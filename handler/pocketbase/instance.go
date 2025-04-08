package pocketbase

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/tinkernels/coredns-pocketbase/handler/pocketbase/cache"
	_ "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/pb_migrations"
)

const (
	// ZonesCacheRefreshInterval defines how often the zones cache should be refreshed.
	ZonesCacheRefreshInterval = 10 * time.Second
	// ZonesCacheKey is the key used to store zones in the cache.
	ZonesCacheKey = "zones"
	// RecordsCacheKeyFormat defines the format for record cache keys.
	RecordsCacheKeyFormat = "records.%s.%s.%s"
)

// Instance represents a PocketBase instance with DNS-specific configurations and caching.
// It manages the connection to PocketBase and provides methods for DNS record operations.
type Instance struct {
	pb            *pocketbase.PocketBase
	suEmail       string
	suPassword    string
	listen        string
	defaultTtl    int
	cacheCapacity int
	// internal
	zonesCache   *cache.ZonesCache
	recordsCache *cache.RecordsCache
	readyChan    chan struct{}
	composer     *Composer
}

// NewWithDataDir creates a new Instance with the specified data directory.
// The data directory will be converted to an absolute path if it's relative.
func NewWithDataDir(dataDir string) *Instance {
	finalDataDir := toAbsPath(dataDir)
	log.Print("instance dataDir: ", finalDataDir)
	inst := &Instance{
		pb: pocketbase.NewWithConfig(pocketbase.Config{
			DefaultDataDir: finalDataDir,
		}),
		readyChan: make(chan struct{}),
	}
	inst.composer = NewComposer(inst)

	return inst
}

// WithSuUserName sets the superuser email for the Instance.
// This is used for authentication and administrative operations.
func (inst *Instance) WithSuUserName(suEmail string) *Instance {
	inst.suEmail = suEmail
	return inst
}

// WithSuPassword sets the superuser password for the Instance.
// This is used for authentication and administrative operations.
func (inst *Instance) WithSuPassword(suPassword string) *Instance {
	inst.suPassword = suPassword
	return inst
}

// WithDefaultTtl sets the default TTL (Time To Live) for DNS records.
// This value will be used when no specific TTL is provided for a record.
func (inst *Instance) WithDefaultTtl(defaultTtl int) *Instance {
	inst.defaultTtl = defaultTtl
	return inst
}

// WithListen sets the address and port where the PocketBase server should listen.
// The format should be "address:port".
func (inst *Instance) WithListen(listen string) *Instance {
	inst.listen = listen
	return inst
}

// WithCacheCapacity sets the capacity for the records cache.
// If capacity is greater than 1, both zones and records caching will be enabled.
// If there's an error initializing the cache, caching will be disabled.
func (inst *Instance) WithCacheCapacity(capacity int) *Instance {
	inst.cacheCapacity = capacity
	if capacity > 1 {
		var err error
		inst.zonesCache, err = cache.NewZonesCache()
		// if error, disable cache
		if err != nil {
			inst.cacheCapacity = 0
			return inst
		}
		inst.recordsCache, err = cache.NewRecordsCache(inst.cacheCapacity)
		if err != nil {
			inst.cacheCapacity = 0
			return inst
		}
	}
	return inst
}

// initZonesCacheRefreshSchedule starts a background goroutine that periodically
// refreshes the zones cache at the specified interval.
func (inst *Instance) initZonesCacheRefreshSchedule() {
	if inst.cacheCapacity > 0 {
		go func() {
			for {
				time.Sleep(ZonesCacheRefreshInterval)
				zones, err := inst.FetchZones()
				if err != nil {
					log.Print(err)
					continue
				}
				inst.zonesCache.Set(ZonesCacheKey, zones)
			}
		}()
	}
}

// Start initializes and starts the PocketBase server.
// It configures the server to run in development mode and sets up necessary hooks.
// Returns an error if the server fails to start.
func (inst *Instance) Start() error {
	args := []string{"--dev", "serve", "--http", inst.listen}
	inst.pb.RootCmd.SetArgs(args)

	inst.pb.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			log.Print("skip installer...")
			e.InstallerFunc = nil
			err := inst.initTheOnlySuperuser(inst.suEmail, inst.suPassword)
			if err != nil {
				return err
			}
			inst.initZonesCacheRefreshSchedule()
			close(inst.readyChan)
			return e.Next()
		},
	})

	err := inst.pb.Bootstrap()
	if err != nil {
		return err
	}

	go func() {
		_ = inst.pb.Start()
	}()

	return nil
}

func (inst *Instance) bindRecordCreateEvent() {
	if inst.cacheCapacity <= 0 {
		return
	}

	cachePopFunc := func(e *core.RecordEvent) error {
		zone := e.Record.GetString("zone")
		name := e.Record.GetString("name")
		typ := e.Record.GetString("record_type")

		cacheKey := fmt.Sprintf(RecordsCacheKeyFormat, zone, name, typ)
		inst.recordsCache.Delete(cacheKey)

		// remove special cache key used in query
		if typ == "A" || typ == "AAAA" || typ == "CNAME" {
			cacheKey = fmt.Sprintf(RecordsCacheKeyFormat, zone, name,
				fmt.Sprintf(RecordsCacheKeyFormat, zone, name, strings.Join([]string{"A", "AAAA", "CNAME"}, ",")))
			inst.recordsCache.Delete(cacheKey)
		}

		inst.zonesCache.Delete(ZonesCacheKey)

		return e.Next()
	}

	inst.pb.OnRecordAfterCreateSuccess(recordCollectionName).BindFunc(cachePopFunc)
	inst.pb.OnRecordAfterUpdateSuccess(recordCollectionName).BindFunc(cachePopFunc)
	inst.pb.OnRecordAfterDeleteSuccess(recordCollectionName).BindFunc(cachePopFunc)
}

// initTheOnlySuperuser ensures there is exactly one superuser with the specified credentials.
// It will create a new superuser if none exists, or update the password if the user exists.
// All other superusers will be deleted to maintain a single superuser configuration.
func (inst *Instance) initTheOnlySuperuser(suEmail, suPwd string) error {
	superusers, err := inst.pb.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	log.Print("init superuser...")

	record, _ := inst.pb.FindAuthRecordByEmail(core.CollectionNameSuperusers, suEmail)
	if record != nil {
		// update the password
		record.Set("password", suPwd)
		err = inst.pb.Save(record)
		if err != nil {
			return err
		}
	} else {
		// create the superuser
		record := core.NewRecord(superusers)
		record.Set("email", suEmail)
		record.Set("password", suPwd)
		err = inst.pb.Save(record)
		if err != nil {
			return nil
		}
	}

	// delete all other superusers
	var existingSuperusers []core.Record
	err = inst.pb.RecordQuery(superusers).All(&existingSuperusers)
	if err != nil {
		return err
	}

	if len(existingSuperusers) > 0 {
		for _, superuser := range existingSuperusers {
			if superuser.GetString("email") != suEmail {
				if err = inst.pb.Delete(&superuser); err == nil {
					return err
				}
			}
		}
	}
	return nil
}

// WaitForReady blocks until the PocketBase instance is fully initialized and ready to serve requests.
func (inst *Instance) WaitForReady() {
	<-inst.readyChan
}

// toAbsPath converts a relative path to an absolute path.
// If the input path is already absolute, it is returned unchanged.
// The path is resolved relative to the executable's directory.
func toAbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	execDir := filepath.Dir(execPath)
	absPath := filepath.Join(execDir, path)
	return absPath
}
