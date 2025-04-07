package handler

import (
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	pb "github.com/tinkernels/coredns-pocketbase/handler/pocketbase"
	"github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
	"golang.org/x/net/context"
)

const (
	pluginName = "pocketbase"
)

type ErrUnsupportedRecordType struct {
	RecordType string
}

func (e *ErrUnsupportedRecordType) Error() string {
	return fmt.Sprintf("unsupported record type: %s", e.RecordType)
}

type PocketBaseHandler struct {
	Next plugin.Handler
	// internal
	pbInst *pb.Instance
}

func (handler *PocketBaseHandler) WarmUp() {
	handler.pbInst.Start()
	handler.pbInst.WaitForReady()
}

// ServeDNS implements the plugin.Handler interface.
// It processes DNS queries by:
// 1. Fetching available zones
// 2. Matching the query against available zones
// 3. Fetching and composing appropriate DNS records
// Returns DNS response code and any error encountered
func (handler *PocketBaseHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	return handler.processQuery(ctx, state)
}

func (handler *PocketBaseHandler) processQuery(ctx context.Context, state request.Request) (int, error) {
	qName := state.Name()
	qType := state.Type()

	zones, err := handler.pbInst.FetchZones()
	if err != nil {
		return handler.errorResponse(state, dns.RcodeServerFailure, err)
	}

	qZone := plugin.Zones(zones).Matches(qName)
	if qZone == "" {
		return plugin.NextOrFailure(handler.Name(), handler.Next, ctx, state.W, state.Req)
	}

	records, err := handler.pbInst.FetchRecords(qZone, qName, qType)
	if err != nil {
		return handler.errorResponse(state, dns.RcodeServerFailure, err)
	}

	var recordNotFound bool
	if len(records) == 0 {
		recordNotFound = true
		// no record found but we are going to return a SOA
		recs, err := handler.pbInst.FetchRecords(qZone, "", "SOA")
		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
		records = append(records, recs...)
	}

	if qType == "SOA" {
		recsNs, err := handler.pbInst.FetchRecords(qZone, qName, "NS")
		if err != nil {
			return handler.errorResponse(state, dns.RcodeServerFailure, err)
		}
		records = append(records, recsNs...)
	}

	if qType == "AXFR" {
		return handler.errorResponse(state, dns.RcodeNotImplemented, nil)
	}

	answers, extras, err := handler.composeResponseMsgs(records)
	// handle error type
	if err != nil {
		if err, ok := err.(*ErrUnsupportedRecordType); ok {
			return handler.errorResponse(state, dns.RcodeNotImplemented, err)
		}
		return handler.errorResponse(state, dns.RcodeServerFailure, err)
	}

	rMsg := new(dns.Msg)
	rMsg.SetReply(state.Req)
	rMsg.Authoritative = true
	rMsg.RecursionAvailable = false
	rMsg.Compress = true

	if !recordNotFound {
		rMsg.Answer = append(rMsg.Answer, answers...)
	} else {
		rMsg.Ns = append(rMsg.Ns, answers...)
		rMsg.Rcode = dns.RcodeNameError
	}
	rMsg.Extra = append(rMsg.Extra, extras...)

	state.SizeAndDo(rMsg)
	rMsg = state.Scrub(rMsg)
	_ = state.W.WriteMsg(rMsg)
	return dns.RcodeSuccess, nil
}

func (handler *PocketBaseHandler) composeResponseMsgs(records []*model.Record) (answers []dns.RR, extras []dns.RR, err error) {
	answers = make([]dns.RR, 0, 10)
	extras = make([]dns.RR, 0, 10)

	for _, record := range records {
		var answer dns.RR
		var err error
		switch record.RecordType {
		case "A":
			answer, extras, err = handler.pbInst.ComposeARecord(record)
		case "AAAA":
			answer, extras, err = handler.pbInst.ComposeAAAARecord(record)
		case "CNAME":
			answer, extras, err = handler.pbInst.ComposeCNAMERecord(record)
		case "SOA":
			answer, extras, err = handler.pbInst.ComposeSOARecord(record)
		case "SRV":
			answer, extras, err = handler.pbInst.ComposeSRVRecord(record)
		case "NS":
			answer, extras, err = handler.pbInst.ComposeNSRecord(record)
		case "MX":
			answer, extras, err = handler.pbInst.ComposeMXRecord(record)
		case "TXT":
			answer, extras, err = handler.pbInst.ComposeTXTRecord(record)
		case "CAA":
			answer, extras, err = handler.pbInst.ComposeCAARecord(record)
		default:
			return nil, nil, &ErrUnsupportedRecordType{RecordType: record.RecordType}
		}

		if err != nil {
			return nil, nil, err
		}
		if answer != nil {
			answers = append(answers, answer)
		}
	}
	return
}

// Name implements the Handler interface.
func (handler *PocketBaseHandler) Name() string { return pluginName }

func NewWithConfig(config *Config) (handler *PocketBaseHandler, err error) {
	finalConfig := config.MixWithEnv()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	handler = &PocketBaseHandler{
		pbInst: nil,
	}

	pbInstance := pb.NewWithDataDir(finalConfig.DataDir).
		WithSuUserName(finalConfig.SuEmail).
		WithSuPassword(finalConfig.SuPassword).
		WithListen(finalConfig.Listen).
		WithDefaultTtl(finalConfig.DefaultTtl)

	handler.pbInst = pbInstance

	return handler, nil
}

func (handler *PocketBaseHandler) errorResponse(state request.Request, rCode int, err error) (int, error) {
	msg := new(dns.Msg)
	msg.SetRcode(state.Req, rCode)
	msg.Authoritative, msg.RecursionAvailable, msg.Compress = true, false, true

	state.SizeAndDo(msg)
	_ = state.W.WriteMsg(msg)
	// Return success as the rCode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
