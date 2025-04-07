package pocketbase

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	_ "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/pb_migrations"
)

type Instance struct {
	pb         *pocketbase.PocketBase
	suEmail    string
	suPassword string
	listen     string
	defaultTtl int
	readyChan  chan struct{}
}

func NewWithDataDir(dataDir string) *Instance {
	finalDataDir := toAbsPath(dataDir)
	log.Print("instance dataDir: ", finalDataDir)
	inst := &Instance{
		pb: pocketbase.NewWithConfig(pocketbase.Config{
			DefaultDataDir: finalDataDir,
		}),
		readyChan: make(chan struct{}),
	}

	return inst
}

func (inst *Instance) WithSuUserName(suEmail string) *Instance {
	inst.suEmail = suEmail
	return inst
}

func (inst *Instance) WithSuPassword(suPassword string) *Instance {
	inst.suPassword = suPassword
	return inst
}

func (inst *Instance) WithDefaultTtl(defaultTtl int) *Instance {
	inst.defaultTtl = defaultTtl
	return inst
}

func (inst *Instance) WithListen(listen string) *Instance {
	inst.listen = listen
	return inst
}

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
			close(inst.readyChan)
			return e.Next()
		},
	})

	err := inst.pb.Bootstrap()
	if err != nil {
		return err
	}

	if err := inst.pb.Start(); err != nil {
		log.Print(err)
		return err
	}
	return nil
}

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

// WaitForReady waits for PocketBase to be ready by waiting for the OnServe event to complete
func (inst *Instance) WaitForReady() {
	<-inst.readyChan
}

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
