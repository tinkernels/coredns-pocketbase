package coredns_pocketbase

import (
	"strconv"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/tinkernels/coredns-pocketbase/handler"
)

func init() {
	caddy.RegisterPlugin("pocketbase", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	pluginInst, err := parseConfig(c)
	if err != nil {
		return plugin.Error("pocketbase", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		pluginInst.Next = next
		return pluginInst
	})
	return nil
}

func parseConfig(c *caddy.Controller) (h *handler.PocketBaseHandler, err error) {
	conf := handler.NewConfig()

	c.Next()
	if c.NextBlock() {
		for {
			switch c.Val() {
			case "listen":
				if c.NextArg() {
					conf = conf.WithListen(c.Val())
				}
			case "data_dir":
				if c.NextArg() {
					conf = conf.WithDataDir(c.Val())
				}
			case "su_email":
				if c.NextArg() {
					conf = conf.WithSuEmail(c.Val())
				}
			case "su_password":
				if c.NextArg() {
					conf = conf.WithSuPassword(c.Val())
				}
			case "default_ttl":
				if c.NextArg() {
					v := c.Val()
					intV, err := strconv.Atoi(v)
					if err == nil {
						conf = conf.WithDefaultTtl(intV)
					} else {
						log.Warningf("default_ttl is not an integer %+v, using default value of %d",
							intV,
							handler.DefaultConfigVal4DefaultTtl())
					}
				}
			case "cache_capacity":
				if c.NextArg() {
					v := c.Val()
					intV, err := strconv.Atoi(v)
					if err == nil {
						conf = conf.WithCacheCapacity(intV)
					} else {
						log.Warningf("default_ttl is not an integer %+v, using default value of %d",
							intV,
							0)
					}
				}
			default:
				if c.Val() != "}" {
					return nil, c.Errf("unknown property '%s'", c.Val())
				}
			}

			if !c.Next() {
				break
			}
		}
	}

	h, err = handler.NewWithConfig(conf)
	return
}
