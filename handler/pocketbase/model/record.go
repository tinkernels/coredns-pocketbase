package model

import (
	"net"
)

// Record represents a DNS record with its basic properties
type Record struct {
	Zone       string `db:"zone" json:"zone"`               // The DNS zone this record belongs to
	Name       string `db:"name" json:"name"`               // The name of the record (without the zone)
	RecordType string `db:"record_type" json:"record_type"` // The type of DNS record (A, AAAA, TXT, etc.)
	Ttl        uint32 `db:"ttl" json:"ttl"`                 // Time to live for the record in seconds
	Content    string `db:"content" json:"content"`         // The content of the record in JSON format
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
