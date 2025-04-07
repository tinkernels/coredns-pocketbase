package coredns_pocketbase

import (
	"github.com/tinkernels/coredns-pocketbase/handler"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	caddy.RegisterPlugin("pocketbase", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	pbPlugin, err := parseConfig(c)
	if err != nil {
		return plugin.Error("pocketbase", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		pbPlugin.Next = next
		return pbPlugin
	})
	return nil
}

func parseConfig(c *caddy.Controller) (*handler.PocketBaseHandler, error) {
	pb := &handler.PocketBaseHandler{}

	return pb, nil
}
