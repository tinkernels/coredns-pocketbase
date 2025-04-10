package pocketbase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
)

const (
	recordWildcardSymbol = "*"
	recordWildcardPrefix = recordWildcardSymbol + "."
	recordCollectionName = "coredns_records"
)

// FetchRecords retrieves DNS records from PocketBase for a given zone, name, and record types.
// It first checks the cache if enabled, then queries the database if not found in cache.
// For wildcard records, it recursively checks parent domains until a match is found.
// Returns a slice of records and any error encountered.
func (inst *Instance) FetchRecords(zone string, name string, types ...string) (recs []*m.Record, err error) {
	coll, err := inst.pb.FindCollectionByNameOrId(recordCollectionName)
	if err != nil {
		log.Errorf("Failed fetching collection [%s], err: %+v",
			recordCollectionName, err)
		return nil, err
	}

	if len(types) == 1 {
		recs, err = inst.fetchSingleTypeRecords(coll, zone, name, types[0])
	} else {
		for _, t := range types {
			tmpRecs, _ := inst.fetchSingleTypeRecords(coll, zone, name, t)
			if len(tmpRecs) > 0 {
				recs = append(recs, tmpRecs...)
			}
		}
	}

	return
}

// fetchSingleTypeRecords retrieves DNS records of a single type from PocketBase for a given zone and name.
func (inst *Instance) fetchSingleTypeRecords(coll *core.Collection, zone string, name string, recordType string) (recs []*m.Record, err error) {
	recs = inst.doFetchSingleTypeRecords(coll, zone, name, recordType)
	// If no records found, chase cname records.
	if len(recs) == 0 && recordType != "CNAME" {
		log.Debugf("No records found in db, zone: [%s], name: [%s], will try chase CNAME", zone, name)
		recs = inst.resolveCNAMEs(coll, zone, name, recordType)
	}
	// If no records found, check for wildcard records.
	if len(recs) == 0 && name != zone {
		log.Debugf("No records found in db, zone: [%s], name: [%s], will try wildcard records", zone, name)
		recs, err = inst.fetchWildCardRecords(coll, zone, name, recordType)
	}
	return recs, err
}

func (inst *Instance) doFetchSingleTypeRecords(coll *core.Collection, zone string, name string, recordType string) (recs []*m.Record) {
	// if cache is enabled, try to get from cache
	if inst.cacheCapacity > 0 {
		log.Debugf("Fetchingrecords from cache, zone: [%s], name: [%s], type: [%s]", zone, name, recordType)
		recs, ok := inst.recordsCache.Get(fmt.Sprintf(RecordsCacheKeyFormat, zone, name, recordType))
		if ok {
			log.Debugf("Found [%d] records in cache, zone: [%s], name: [%s], type: [%s]", len(recs), zone, name, recordType)
			return recs
		}
	}
	q := inst.pb.RecordQuery(coll).
		Select("name", "zone", "ttl", "record_type", "content").
		Where(dbx.NewExp("zone = {:zone}", dbx.Params{"zone": zone})).
		AndWhere(dbx.NewExp("name = {:name}", dbx.Params{"name": name})).
		AndWhere(dbx.NewExp("record_type = {:record_type}", dbx.Params{"record_type": recordType}))

	err := q.All(&recs)
	if err != nil {
		log.Errorf("Fetching records from db failed, zone: [%s], name: [%s], type: %s, err: %+v", zone, name, recordType, err)
		return nil
	}
	log.Debugf("Fetched with SQL: %s; Parameters: %+v",
		q.Build().SQL(), q.Build().Params())
	log.Debugf("Records [%d] fetched from db, zone: [%s], name: [%s], type: %s", len(recs), zone, name, recordType)
	// if cache is enabled, set to cache
	if inst.cacheCapacity > 0 {
		log.Debugf("Setting [%d] records to cache, zone: [%s], name: [%s], type: %s", len(recs), zone, name, recordType)
		inst.recordsCache.Set(fmt.Sprintf(RecordsCacheKeyFormat, zone, name, recordType), recs)
	}
	return recs
}

func (inst *Instance) resolveCNAMEs(coll *core.Collection, zone string, name string, recordType string) (recs []*m.Record) {
	cnameZone, cname := zone, name
	for { // First get CNAME records for the name
		cnameRecs := inst.doFetchSingleTypeRecords(coll, cnameZone, cname, "CNAME")

		// If no CNAME records found, return empty slice
		if len(cnameRecs) == 0 {
			break
		}

		// Take only the first CNAME record since multiple CNAMEs for the same name are illegal
		cnameRec := cnameRecs[0]

		// Get the target name from CNAME content
		var cnameRecord m.CNAMERecord
		jsonErr := json.Unmarshal([]byte(cnameRec.Content), &cnameRecord)
		if jsonErr != nil {
			log.Errorf("Failed to unmarshal CNAME record, zone: %s, name: %s, err: %+v",
				cnameRec.Zone, cnameRec.Name, jsonErr)
			break
		}
		targetName, targetZone := cnameRecord.Host, cnameRecord.Zone
		if targetName == "" || targetZone == "" {
			log.Errorf("Invalid CNAME record, zone: %s, name: %s, content: %s",
				cnameRec.Zone, cnameRec.Name, cnameRec.Content)
			break
		}
		recs = append(recs, cnameRec)

		log.Debugf("Resolved CNAME, name: [%s], target name: [%s], target zone: [%s]",
			name, targetName, targetZone)

		targetRecs := inst.doFetchSingleTypeRecords(coll, targetZone, targetName, recordType)

		if len(targetRecs) == 0 {
			cnameZone, cname = targetZone, targetName
			continue
		}
		recs = append(recs, targetRecs...)
		break
	}
	return recs
}

// fetchWildCardRecords attempts to find wildcard records
// recursively until it finds matching records.
// e.g. x.y.z -> *.y.z -> *.z -> *
func (inst *Instance) fetchWildCardRecords(coll *core.Collection, zone string, name string, recordType string) (recs []*m.Record, err error) {
	name = strings.TrimPrefix(name, recordWildcardPrefix)

	target := recordWildcardSymbol
	i, end := dns.NextLabel(name, 0)
	if !end {
		target = recordWildcardPrefix + name[i:]
	}

	if target == recordWildcardSymbol {
		log.Debugf("Wildcard name is not allowed, name: %s", target)
		return nil, nil
	}

	return inst.fetchSingleTypeRecords(coll, zone, target, recordType)
}

// FetchZones retrieves all unique DNS zones from PocketBase.
// It first checks the cache if enabled, then queries the database if not found in cache.
// Returns a slice of zone names and any error encountered.
func (inst *Instance) FetchZones() (zones []string, err error) {
	// if cache is enabled, try to get from cache
	if inst.cacheCapacity > 0 {
		zones, ok := inst.zonesCache.Get(ZonesCacheKey)
		if ok {
			log.Debugf("Found %d zones in cache", len(zones))
			return zones, nil
		}
	}
	zones, err = inst.fetchZonesFromDb()
	if err != nil {
		log.Errorf("Failed to fetch zones from db, err: %+v", err)
		return nil, err
	}
	// if cache is enabled, set to cache
	if inst.cacheCapacity > 0 {
		inst.zonesCache.Set(ZonesCacheKey, zones)
	}
	return
}

func (inst *Instance) fetchZonesFromDb() (zones []string, err error) {
	coll, err := inst.pb.FindCollectionByNameOrId(recordCollectionName)
	if err != nil {
		log.Errorf("Failed fetching collection [%s], err: %+v", recordCollectionName, err)
		return nil, err
	}
	var zonesContainer []struct {
		Zone string `db:"zone" json:"zone"`
	}
	err = inst.pb.RecordQuery(coll).
		Select("zone").
		Distinct(true).
		All(&zonesContainer)
	if err == nil {
		for _, z := range zonesContainer {
			log.Debugf("Found zone in db, zone: %s", z.Zone)
			zones = append(zones, z.Zone)
		}
	}
	return
}

// Hosts retrieves and composes DNS resource records for a given zone and name.
// It supports A, AAAA, and CNAME record types.
// Returns a slice of DNS resource records and any error encountered.
func (inst *Instance) Hosts(zone string, name string) (answers []dns.RR, err error) {
	recs, err := inst.FetchRecords(zone, name, "A", "AAAA", "CNAME")
	if err != nil {
		log.Errorf("Failed to fetch records, zone: [%s], name: [%s], err: %+v", zone, name, err)
		return nil, err
	}

	answers = make([]dns.RR, 0)

	for _, rec := range recs {
		switch rec.RecordType {
		case "A":
			aRec, _, err := inst.ComposeARecord(rec)
			if err != nil {
				log.Errorf("Failed to compose A record, zone: [%s], name: [%s], err: %+v", zone, name, err)
				return nil, err
			}
			answers = append(answers, aRec)
		case "AAAA":
			aaaaRec, _, err := inst.ComposeAAAARecord(rec)
			if err != nil {
				log.Errorf("Failed to compose AAAA record, zone: [%s], name: [%s], err: %+v", zone, name, err)
				return nil, err
			}
			answers = append(answers, aaaaRec)
		case "CNAME":
			cnameRec, _, err := inst.ComposeCNAMERecord(rec)
			if err != nil {
				log.Errorf("Failed to compose CNAME record, zone: [%s], name: [%s], err: %+v", zone, name, err)
				return nil, err
			}
			answers = append(answers, cnameRec)
		}
	}
	return
}
