package pocketbase

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/pocketbase/dbx"
)

const (
	recordWildcardSymbol = "*"
	recordWildcardPrefix = recordWildcardSymbol + "."
)

func (inst *Instance) findRecords(zone string, name string, types ...string) (recs []*Record, err error) {
	coll, err := inst.pb.FindCollectionByNameOrId("coredns_records")
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
	// add handler into records
	for _, rec := range recs {
		rec.inst = inst
	}
	return
}

// findWildcardRecords attempts to find wildcard records
// recursively until it finds matching records.
// e.g. x.y.z -> *.y.z -> *.z -> *
func (inst *Instance) findWildcardRecords(zone string, name string, types ...string) (recs []*Record, err error) {

	if name == recordWildcardSymbol {
		return nil, nil
	}

	name = strings.TrimPrefix(name, recordWildcardPrefix)

	target := recordWildcardSymbol
	i, shot := dns.NextLabel(name, 0)
	if !shot {
		target = recordWildcardPrefix + name[i:]
	}

	return inst.findRecords(zone, target, types...)
}

func (inst *Instance) findZones() (zones []string, err error) {
	coll, err := inst.pb.FindCollectionByNameOrId("coredns_records")
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
	return
}

func (inst *Instance) hosts(zone string, name string) (answers []dns.RR, err error) {
	recs, err := inst.findRecords(zone, name, "A", "AAAA", "CNAME")
	if err != nil {
		return nil, err
	}

	answers = make([]dns.RR, 0)

	for _, rec := range recs {
		switch rec.RecordType {
		case "A":
			aRec, _, err := rec.AsARecord()
			if err != nil {
				return nil, err
			}
			answers = append(answers, aRec)
		case "AAAA":
			aRec, _, err := rec.AsAAAARecord()
			if err != nil {
				return nil, err
			}
			answers = append(answers, aRec)
		case "CNAME":
			aRec, _, err := rec.AsCNAMERecord()
			if err != nil {
				return nil, err
			}
			answers = append(answers, aRec)
		}
	}
	return
}
