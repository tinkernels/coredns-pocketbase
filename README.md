# pocketbase

PocketBase backend for CoreDNS

## Name

pocketbase - PocketBase backend for CoreDNS

## Description

This plugin uses PocketBase as a backend to store DNS records. These will then can served by CoreDNS. The backend uses a
simple single table data structure that can add and remove records from the DNS server.

## Syntax

```
pocketbase {
    [liten LISTEN]
    [data_dir DATA_DIR]
    [su_email SU_EMAIL]
    [su_password SU_PASSWORD]
    [default_ttl DEFAULT_TTL]
    [cache_capacity CACHE_CAPACITY]
}
```

- `liten` pocketbase listening http address, default to `[::]:8090`,
- `data_dir` dir to store pocketbase data, default to `pb_data`,
- `su_email` superuser login email, can be overwritten by environment variable `COREDNS_PB_SUPERUSER_EMAIL`, default to `su@pocketbase.internal`,
- `su_password` superuser password, can be overwritten by environment variable `COREDNS_PB_SUPERUSER_PWD`, default to `pwd@pocketbase.internal`,
- `default_ttl` default ttl to use, default to `30`,
- `cache_capacity` zone data cache capacity, `0` to disable cache, default to `0`.

## Supported Record Types

A, AAAA, CNAME, SOA, TXT, NS, MX, CAA and SRV. Wildcard records are supported as well. This backend doesn't support AXFR
requests.

## Setup (as an external plugin)

Add this as an external plugin in `plugin.cfg` file:

```
pocketbase:github.com/tinkernels/coredns-pocketbase
```

*P.S.place pocketbase above cache plugin is recommended.*

then run

```shell script
$ go generate
$ go build
```

Add any required modules to CoreDNS code as prompted.

### Acknowledgements and Credits

This plugin, is inspired by

- [https://github.com/wenerme/coredns-pdsql](https://github.com/wenerme/coredns-pdsql)
- [https://github.com/arvancloud/redis](https://github.com/arvancloud/redis)
- [https://github.com/cloud66-oss/coredns_mysql](https://github.com/cloud66-oss/coredns_mysql)
