package handler

import (
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	pb "github.com/tinkernels/coredns-pocketbase/handler/pocketbase"
	"golang.org/x/net/context"
)

const (
	pluginName = "pocketbase"
	defaultTtl = 30
)

type PocketBaseHandler struct {
	Next               plugin.Handler
	pocketbaseInstance *pb.Instance
}

func (handler *PocketBaseHandler) hosts(zone string, ns string) ([]dns.RR, error) {
	panic("unimplemented")
}

// ServeDNS implements the plugin.Handler interface.
func (handler *PocketBaseHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	panic("unimplemented")
}

// Name implements the Handler interface.
func (handler *PocketBaseHandler) Name() string { return pluginName }
