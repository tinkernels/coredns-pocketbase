package handler

import (
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	pb "github.com/tinkernels/coredns-pocketbase/handler/pocketbase"
	"golang.org/x/net/context"
)

const (
	pluginName = "pocketbase"
)

type PocketBaseHandler struct {
	Next               plugin.Handler
	pocketbaseInstance *pb.Instance
}

// ServeDNS implements the plugin.Handler interface.
func (handler *PocketBaseHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	panic("unimplemented")
}

// Name implements the Handler interface.
func (handler *PocketBaseHandler) Name() string { return pluginName }

func NewWithConfig(config *Config) (handler *PocketBaseHandler, err error) {
	finalConfig := config.MixWithEnv()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	handler = &PocketBaseHandler{
		pocketbaseInstance: nil,
	}

	pbInstance := pb.NewWithDataDir(finalConfig.DataDir).
		WithSuUserName(finalConfig.SuEmail).
		WithSuPassword(finalConfig.SuPassword).
		WithListen(finalConfig.Listen).
		WithDefaultTtl(finalConfig.DefaultTtl)

	handler.pocketbaseInstance = pbInstance

	return handler, nil
}
