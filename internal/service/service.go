package service

import (
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/tauraamui/kvs/v2"
	"github.com/tauraamui/redpanda/internal/logging"
)

type Server interface {
	Listen(port string) error
	ShutdownWithTimeout(d time.Duration) error
	Cleanup() error
}

type server struct {
	db  kvs.KVDB
	app *fiber.App
}

func New() (Server, error) {
	parentDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	conn, err := badger.Open(badger.DefaultOptions(filepath.Join(parentDir, "redpanda", "data")).WithLogger(nil))
	if err != nil {
		return nil, err
	}

	db, err := kvs.NewKVDB(conn)
	if err != nil {
		return nil, err
	}

	svr := server{
		db:  db,
		app: fiber.New(fiber.Config{DisableStartupMessage: true}),
	}

	log := logging.New()
	svr.app.Post("/insert/:type/:uuid", handleInserts(log, db))
	svr.app.Post("/fetch/:type/:uuid", handleFetch(log, db))

	return svr, nil
}

func (s server) Listen(port string) error {
	return s.app.Listen(port)
}

func (s server) Cleanup() error {
	s.db.DumpToStdout()
	s.db.Close()
	return nil
}

func (s server) Shutdown() error {
	return s.app.Shutdown()
}

func (s server) ShutdownWithTimeout(d time.Duration) error {
	return s.app.ShutdownWithTimeout(d)
}
