package main

import (
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"
	_ "github.com/tinkernels/coredns-pocketbase/cmd/test/standalone_pocketbase/pb_migrations"
	"os"
	"path/filepath"
)

func main() {
	log.Info("Starting pocketbase...")

	pb := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: defaultDataDir(),
	})

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// migrate command (with go templates)
	migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangGo,
		Automigrate:  true,
		Dir:          defaultMigrationDir(pb),
	})

	args := []string{"--dev", "serve", "--http", "[::]:8090"}
	pb.RootCmd.SetArgs(args)

	// static route to serves files from the provided public dir
	// (if publicDir exists and the route path is not already defined)
	pb.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			return e.Next()
		},
	})

	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
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

// the default pb_migrations dir location is relative to the executable
func defaultMigrationDir(pb *pocketbase.PocketBase) string {
	return filepath.Join(pb.DataDir(), "pb_migrations")
}

// the default pb_data dir location is relative to the executable
func defaultDataDir() string {
	return toAbsPath("../testdata/pb_data")
}
