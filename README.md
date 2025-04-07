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
    [su_username SU_USERNAME]
    [su_password SU_PASSWORD]
    [default_ttl DEFAULT_TTL]
    [cache_capacity CACHE_CAPACITY]
}
```

- `liten` pocketbase listen http address, default to 0.0.0.0:8090,
- `data_dir` dir to store pocketbase data,
- `su_email` superuser login name, can be overwritten by environment variable `COREDNS_PB_SUPERUSER_EMAIL`, default to su@pocketbase.internal,
- `su_password` superuser password, can be overwritten by environment variable `COREDNS_PB_SUPERUSER_PWD`, default to pwd@pocketbase.internal,
- `default_ttl` default ttl to use, default to 30s,
- `cache_capacity` zone data cache capacity, 0 to disable cache, default to 0.

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

## Database Setup

This plugin doesn't create or migrate database schema for its use yet. To create the database and tables, use the
following table structure (note the table name prefix):

```sql
CREATE TABLE `coredns_records`
(
    `id`          INT          NOT NULL AUTO_INCREMENT,
    `zone`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `ttl`         INT DEFAULT NULL,
    `content`     TEXT,
    `record_type` VARCHAR(255) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = INNODB AUTO_INCREMENT = 6 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;
```

## Record setup

Each record served by this plugin, should belong to the zone it is allowed to server by CoreDNS. Here are some examples:

```sql
-- Insert batch #1
INSERT INTO coredns_records (zone, name, ttl, content, record_type)
VALUES ('example.org.', 'foo', 30, '{"ip": "1.1.1.1"}', 'A'),
       ('example.org.', 'foo', '60', '{"ip": "1.1.1.0"}', 'A'),
       ('example.org.', 'foo', 30, '{"text": "hello"}', 'TXT'),
       ('example.org.', 'foo', 30, '{"host" : "foo.example.org.","priority" : 10}', 'MX');
```

These can be queries using `dig` like this:

```shell script
$ dig A MX foo.example.org 
```

### Acknowledgements and Credits

This plugin, is inspired by https://github.com/wenerme/coredns-pdsql and https://github.com/arvancloud/redis

### Development

To develop this plugin further, make sure you can compile CoreDNS locally and get this repo (
`go get github.com/cloud66-oss/coredns_mysql`). You can switch the CoreDNS mod file to look for the plugin code locally
while you're developing it:

Put `replace github.com/cloud66-oss/coredns_mysql => LOCAL_PATH_TO_THE_SOURCE_CODE` at the end of the `go.mod` file in
CoreDNS code.

Pull requests and bug reports are welcome!

