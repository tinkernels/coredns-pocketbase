package pocketbase

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/pocketbase/dbx"
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
	// if cache is enabled, try to get from cache
	if inst.cacheCapacity > 0 {
		recs, ok := inst.recordsCache.Get(fmt.Sprintf(RecordsCacheKeyFormat, zone, name, strings.Join(types, ",")))
		if ok {
			return recs, nil
		}
	}
	coll, err := inst.pb.FindCollectionByNameOrId(recordCollectionName)
	if err != nil {
		return nil, err
	}

	q := inst.pb.RecordQuery(coll).
		Select("name", "zone", "ttl", "record_type", "content").
		Where(dbx.NewExp("zone = {:zone}", dbx.Params{"zone": zone})).
		AndWhere(dbx.NewExp("name = {:name}", dbx.Params{"name": name}))

	// if multiple types are provided, OR them together
	if len(types) > 1 {
		typesQ := make([]dbx.Expression, 0)
		for i, typ := range types {
			typesQ = append(typesQ, dbx.NewExp(fmt.Sprintf("record_type = {:record_type%d}", i),
				dbx.Params{fmt.Sprintf("record_type%d", i): typ}))
		}
		q = q.AndWhere(dbx.Or(typesQ...))
	} else {
		// if a single type is provided, AND it
		q = q.AndWhere(dbx.NewExp("record_type = {:record_type}", dbx.Params{"record_type": types[0]}))
	}

	err = q.All(&recs)
	if err != nil {
		return nil, err
	}
	// If no records found, check for wildcard records.
	if len(recs) == 0 && name != zone {
		recs, err = inst.fetchWildCardRecords(zone, name, types...)
	}
	// if cache is enabled, set to cache
	if inst.cacheCapacity > 0 {
		inst.recordsCache.Set(fmt.Sprintf(RecordsCacheKeyFormat, zone, name, strings.Join(types, ",")), recs)
	}
	return
}

// fetchWildCardRecords attempts to find wildcard records
// recursively until it finds matching records.
// e.g. x.y.z -> *.y.z -> *.z -> *
func (inst *Instance) fetchWildCardRecords(zone string, name string, types ...string) (recs []*m.Record, err error) {
	if name == recordWildcardSymbol {
		return nil, nil
	}

	name = strings.TrimPrefix(name, recordWildcardPrefix)

	target := recordWildcardSymbol
	i, shot := dns.NextLabel(name, 0)
	if !shot {
		target = recordWildcardPrefix + name[i:]
	}

	return inst.FetchRecords(zone, target, types...)
}

// FetchZones retrieves all unique DNS zones from PocketBase.
// It first checks the cache if enabled, then queries the database if not found in cache.
// Returns a slice of zone names and any error encountered.
func (inst *Instance) FetchZones() (zones []string, err error) {
	// if cache is enabled, try to get from cache
	if inst.cacheCapacity > 0 {
		zones, ok := inst.zonesCache.Get(ZonesCacheKey)
		if ok {
			return zones, nil
		}
	}
	coll, err := inst.pb.FindCollectionByNameOrId(recordCollectionName)
	if err != nil {
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
			zones = append(zones, z.Zone)
		}
	}
	// if cache is enabled, set to cache
	if inst.cacheCapacity > 0 {
		inst.zonesCache.Set(ZonesCacheKey, zones)
	}
	return
}

// Hosts retrieves and composes DNS resource records for a given zone and name.
// It supports A, AAAA, and CNAME record types.
// Returns a slice of DNS resource records and any error encountered.
func (inst *Instance) Hosts(zone string, name string) (answers []dns.RR, err error) {
	recs, err := inst.FetchRecords(zone, name, "A", "AAAA", "CNAME")
	if err != nil {
		return nil, err
	}

	answers = make([]dns.RR, 0)

	for _, rec := range recs {
		switch rec.RecordType {
		case "A":
			aRec, _, err := inst.ComposeARecord(rec)
			if err != nil {
				return nil, err
			}
			answers = append(answers, aRec)
		case "AAAA":
			aaaaRec, _, err := inst.ComposeAAAARecord(rec)
			if err != nil {
				return nil, err
			}
			answers = append(answers, aaaaRec)
		case "CNAME":
			cnameRec, _, err := inst.ComposeCNAMERecord(rec)
			if err != nil {
				return nil, err
			}
			answers = append(answers, cnameRec)
		}
	}
	return
}
