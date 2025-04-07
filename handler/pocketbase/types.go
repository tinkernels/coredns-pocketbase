package pocketbase

import (
	"encoding/json"
	"net"
	"time"

	"github.com/miekg/dns"
)

// Record represents a DNS record with its basic properties
type Record struct {
	Zone       string `db:"zone" json:"zone"`               // The DNS zone this record belongs to
	Name       string `db:"name" json:"name"`               // The name of the record (without the zone)
	RecordType string `db:"record_type" json:"record_type"` // The type of DNS record (A, AAAA, TXT, etc.)
	Ttl        uint32 `db:"ttl" json:"ttl"`                 // Time to live for the record in seconds
	Content    string `db:"content" json:"content"`         // The content of the record in JSON format

	inst *Instance // Reference to the handler managing this record
}

// ARecord represents an A (IPv4) DNS record
type ARecord struct {
	Ip net.IP `json:"ip"` // IPv4 address
}

// AAAARecord represents an AAAA (IPv6) DNS record
type AAAARecord struct {
	Ip net.IP `json:"ip"` // IPv6 address
}

// TXTRecord represents a TXT DNS record
type TXTRecord struct {
	Text string `json:"text"` // Text content of the record
}

// CNAMERecord represents a CNAME DNS record
type CNAMERecord struct {
	Host string `json:"host"` // Target hostname
}

// NSRecord represents an NS (Name Server) DNS record
type NSRecord struct {
	Host string `json:"host"` // Name server hostname
}

// MXRecord represents an MX (Mail Exchange) DNS record
type MXRecord struct {
	Host       string `json:"host"`       // Mail server hostname
	Preference uint16 `json:"preference"` // Priority of the mail server
}

// SRVRecord represents an SRV (Service) DNS record
type SRVRecord struct {
	Priority uint16 `json:"priority"` // Priority of the service
	Weight   uint16 `json:"weight"`   // Weight for load balancing
	Port     uint16 `json:"port"`     // Port number of the service
	Target   string `json:"target"`   // Target hostname
}

// SOARecord represents an SOA (Start of Authority) DNS record
type SOARecord struct {
	Ns      string `json:"ns"`      // Primary name server
	MBox    string `json:"mbox"`    // Email address of the administrator
	Refresh uint32 `json:"refresh"` // Refresh interval in seconds
	Retry   uint32 `json:"retry"`   // Retry interval in seconds
	Expire  uint32 `json:"expire"`  // Expiration time in seconds
	MinTtl  uint32 `json:"minttl"`  // Minimum TTL in seconds
}

// CAARecord represents a CAA (Certification Authority Authorization) DNS record
type CAARecord struct {
	Flag  uint8  `json:"flag"`  // Critical flag
	Tag   string `json:"tag"`   // Property identifier
	Value string `json:"value"` // Property value
}

func (rec *Record) AsARecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.A)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *ARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if retRec.Ip == nil {
		return nil, nil, nil
	}
	r.A = retRec.Ip
	return r, nil, nil
}

func (rec *Record) AsAAAARecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.AAAA)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeAAAA,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *AAAARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if retRec.Ip == nil {
		return nil, nil, nil
	}

	r.AAAA = retRec.Ip
	return r, nil, nil
}

func (rec *Record) AsTXTRecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.TXT)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *TXTRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if len(retRec.Text) == 0 {
		return nil, nil, nil
	}

	r.Txt = split255(retRec.Text)
	return r, nil, nil
}

func (rec *Record) AsCNAMERecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.CNAME)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeCNAME,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *CNAMERecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		return nil, nil, nil
	}
	r.Target = dns.Fqdn(retRec.Host)
	return r, nil, nil
}

func (rec *Record) AsNSRecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.NS)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeNS,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *NSRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		return nil, nil, nil
	}

	r.Ns = retRec.Host
	extras, err = rec.inst.hosts(rec.Zone, r.Ns)
	if err != nil {
		return nil, nil, err
	}
	return r, extras, nil
}

func (rec *Record) AsMXRecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.MX)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeMX,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *MXRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if len(retRec.Host) == 0 {
		return nil, nil, nil
	}

	r.Mx = retRec.Host
	r.Preference = retRec.Preference
	extras, err = rec.inst.hosts(rec.Zone, retRec.Host)
	if err != nil {
		return nil, nil, err
	}

	return r, extras, nil
}

func (rec *Record) AsSRVRecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.SRV)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeSRV,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *SRVRecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if len(retRec.Target) == 0 {
		return nil, nil, nil
	}

	r.Target = retRec.Target
	r.Weight = retRec.Weight
	r.Port = retRec.Port
	r.Priority = retRec.Priority
	return r, nil, nil
}

func (rec *Record) AsSOARecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.SOA)
	var retRec *SOARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if retRec.Ns == "" {
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rec.fqdn()),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    rec.minTtl(),
		}
		r.Ns = "ns1." + rec.Name
		r.Mbox = "hostmaster." + rec.Name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = rec.minTtl()
	} else {
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rec.Zone),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    rec.minTtl(),
		}
		r.Ns = retRec.Ns
		r.Mbox = retRec.MBox
		r.Refresh = retRec.Refresh
		r.Retry = retRec.Retry
		r.Expire = retRec.Expire
		r.Minttl = retRec.MinTtl
	}
	r.Serial = rec.serial()

	return r, nil, nil
}

func (rec *Record) AsCAARecord() (record dns.RR, extras []dns.RR, err error) {
	r := new(dns.CAA)
	r.Hdr = dns.RR_Header{
		Name:   dns.Fqdn(rec.fqdn()),
		Rrtype: dns.TypeCAA,
		Class:  dns.ClassINET,
		Ttl:    rec.minTtl(),
	}
	var retRec *CAARecord
	err = json.Unmarshal([]byte(rec.Content), &retRec)
	if err != nil {
		return nil, nil, err
	}

	if retRec.Value == "" || retRec.Tag == "" {
		return nil, nil, nil
	}

	r.Flag = retRec.Flag
	r.Tag = retRec.Tag
	r.Value = retRec.Value

	return r, nil, nil
}

// minTtl returns the minimum TTL for the record
// If Ttl is not set, returns the default TTL from configuration
func (rec *Record) minTtl() uint32 {
	if rec.Ttl == 0 {
		return uint32(rec.inst.defaultTtl)
	}
	return rec.Ttl
}

// serial generates a serial number for SOA records based on current Unix timestamp
func (rec *Record) serial() uint32 {
	return uint32(time.Now().Unix())
}

// split255 splits a string into chunks of maximum 255 characters
// This is required for TXT records which have a maximum length of 255 characters per chunk
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

// fqdn returns the fully qualified domain name for the record
// Combines the record name with its zone if name is not empty
func (rec *Record) fqdn() string {
	if rec.Name == "" {
		return rec.Zone
	}
	return rec.Name + "." + rec.Zone
}
