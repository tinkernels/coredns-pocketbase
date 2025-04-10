package pocketbase

import (
	"encoding/json"
	"time"

	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	m "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
)

// Composer is responsible for composing DNS records from PocketBase data.
// It provides methods to convert PocketBase record data into DNS resource records.
type Composer struct {
	inst *Instance
}

// NewComposer creates a new Composer instance with the given PocketBase instance.
func NewComposer(inst *Instance) *Composer {
	return &Composer{
		inst: inst,
	}
}

// ComposeARecord creates a DNS A record from a PocketBase record.
// It returns the composed A record and any additional records needed.
func (inst *Instance) ComposeARecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.A)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.ARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal A record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if retRec.Ip == nil {
		log.Debugf("A record is nil, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}
	r.A = retRec.Ip
	log.Debugf("Composed A record, zone: %s, name: %s, ip: %s", rec.Zone, rec.Name, retRec.Ip)
	return r, nil, nil
}

// ComposeAAAARecord creates a DNS AAAA record from a PocketBase record.
// It returns the composed AAAA record and any additional records needed.
func (inst *Instance) ComposeAAAARecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.AAAA)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeAAAA,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.AAAARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal AAAA record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if retRec.Ip == nil {
		log.Debugf("AAAA record is nil, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.AAAA = retRec.Ip
	log.Debugf("Composed AAAA record, zone: %s, name: %s, ip: %s", rec.Zone, rec.Name, retRec.Ip)
	return r, nil, nil
}

// ComposeTXTRecord creates a DNS TXT record from a PocketBase record.
// It returns the composed TXT record and any additional records needed.
func (inst *Instance) ComposeTXTRecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.TXT)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.TXTRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal TXT record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if len(retRec.Text) == 0 {
		log.Debugf("TXT record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.Txt = split255(retRec.Text)
	log.Debugf("Composed TXT record, zone: %s, name: %s, text: %s", rec.Zone, rec.Name, retRec.Text)
	return r, nil, nil
}

// ComposeCNAMERecord creates a DNS CNAME record from a PocketBase record.
// It returns the composed CNAME record and any additional records needed.
func (inst *Instance) ComposeCNAMERecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.CNAME)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeCNAME,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.CNAMERecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal CNAME record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		log.Debugf("CNAME record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}
	r.Target = dns.Fqdn(retRec.Host)
	log.Debugf("Composed CNAME record, zone: %s, name: %s, target: %s", rec.Zone, rec.Name, retRec.Host)
	return r, nil, nil
}

// ComposeNSRecord creates a DNS NS record from a PocketBase record.
// It returns the composed NS record and any additional records needed.
func (inst *Instance) ComposeNSRecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.NS)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeNS,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.NSRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal NS record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		log.Debugf("NS record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.Ns = retRec.Host
	extras, err = inst.Hosts(rec.Zone, r.Ns)
	if err != nil {
		log.Errorf("Failed to compose NS record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}
	log.Debugf("Composed NS record, zone: %s, name: %s, target: %s", rec.Zone, rec.Name, r.Ns)
	return r, extras, nil
}

// ComposeMXRecord creates a DNS MX record from a PocketBase record.
// It returns the composed MX record and any additional records needed.
func (inst *Instance) ComposeMXRecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.MX)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeMX,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.MXRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal MX record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		log.Debugf("MX record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.Mx = retRec.Host
	r.Preference = retRec.Preference
	extras, err = inst.Hosts(rec.Zone, retRec.Host)
	if err != nil {
		log.Errorf("Failed to compose MX record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}
	log.Debugf("Composed MX record, zone: %s, name: %s, target: %s", rec.Zone, rec.Name, r.Mx)
	return r, extras, nil
}

// ComposeSRVRecord creates a DNS SRV record from a PocketBase record.
// It returns the composed SRV record and any additional records needed.
func (inst *Instance) ComposeSRVRecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.SRV)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeSRV,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.SRVRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal SRV record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if len(retRec.Target) == 0 {
		log.Debugf("SRV record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.Target = retRec.Target
	r.Weight = retRec.Weight
	r.Port = retRec.Port
	r.Priority = retRec.Priority
	log.Debugf("Composed SRV record, zone: %s, name: %s, target: %s", rec.Zone, rec.Name, r.Target)
	return r, nil, nil
}

// ComposeSOARecord creates a DNS SOA record from a PocketBase record.
// It returns the composed SOA record and any additional records needed.
func (inst *Instance) ComposeSOARecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.SOA)
	var retRec *m.SOARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal SOA record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if retRec.Ns == "" {
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rec.Name),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    inst.tryRefillTtl(rec),
		}
		// TODO: get from config
		r.Ns = "ns1." + rec.Name
		r.Mbox = "hostmaster." + rec.Name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = inst.tryRefillTtl(rec)
	} else {
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rec.Zone),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    inst.tryRefillTtl(rec),
		}
		r.Ns = retRec.Ns
		r.Mbox = dns.Fqdn(retRec.MBox)
		r.Refresh = retRec.Refresh
		r.Retry = retRec.Retry
		r.Expire = retRec.Expire
		r.Minttl = retRec.MinTtl
	}
	r.Serial = serial()
	log.Debugf("Composed SOA record, zone: %s, name: %s, serial: %d", rec.Zone, rec.Name, r.Serial)
	return r, nil, nil
}

// ComposeCAARecord creates a DNS CAA record from a PocketBase record.
// It returns the composed CAA record and any additional records needed.
func (inst *Instance) ComposeCAARecord(rec *m.Record) (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.CAA)
	r.Hdr = dns.RR_Header{
		Name:   rec.Name,
		Rrtype: dns.TypeCAA,
		Class:  dns.ClassINET,
		Ttl:    inst.tryRefillTtl(rec),
	}
	var retRec *m.CAARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		log.Errorf("Failed to unmarshal CAA record, zone: %s, name: %s, err: %+v", rec.Zone, rec.Name, err)
		return nil, nil, err
	}

	if retRec.Value == "" || retRec.Tag == "" {
		log.Debugf("CAA record is empty, zone: %s, name: %s", rec.Zone, rec.Name)
		return nil, nil, nil
	}

	r.Flag = retRec.Flag
	r.Tag = retRec.Tag
	r.Value = retRec.Value
	log.Debugf("Composed CAA record, zone: %s, name: %s, flag: %d, tag: %s, value: %s", rec.Zone, rec.Name, r.Flag, r.Tag, r.Value)
	return r, nil, nil
}

// tryRefillTtl returns the TTL value for a record.
// If the record's TTL is not set (0), it returns the default TTL from the instance configuration.
func (inst *Instance) tryRefillTtl(rec *m.Record) uint32 {
	if rec.Ttl == 0 {
		log.Debugf("Record TTL is 0, zone: %s, name: %s, using default TTL: %d", rec.Zone, rec.Name, inst.defaultTtl)
		return uint32(inst.defaultTtl)
	}
	log.Debugf("Record TTL is not 0, zone: %s, name: %s, using TTL: %d", rec.Zone, rec.Name, rec.Ttl)
	return rec.Ttl
}

// serial generates a serial number for SOA records based on the current Unix timestamp.
func serial() uint32 {
	log.Debug("Generating serial number...")
	return uint32(time.Now().Unix())
}

// This is required for TXT records which have a maximum length of 255 characters per chunk.
func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	var sx []string
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break

		}
		p, i = p+255, i+255
	}

	return sx
}
