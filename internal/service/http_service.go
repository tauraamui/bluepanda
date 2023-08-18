// Copyright (c) 2023 Adam Prakash Stringer
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted (subject to the limitations in the disclaimer
// below) provided that the following conditions are met:
//
//     * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//
//     * Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer in the
//     documentation and/or other materials provided with the distribution.
//
//     * Neither the name of the copyright holder nor the names of its
//     contributors may be used to endorse or promote products derived from this
//     software without specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY
// THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR
// BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package service

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/tauraamui/bluepanda/internal/logging"
	"github.com/tauraamui/bluepanda/pkg/kvs"
)

type Server interface {
	Listen(port string) error
	ShutdownWithTimeout(d time.Duration) error
	Cleanup(log logging.Logger) error
}

type server struct {
	db  kvs.KVDB
	app *fiber.App
}

func NewHTTP(log logging.Logger) (Server, error) {
	parentDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	conn, err := badger.Open(badger.DefaultOptions(filepath.Join(parentDir, "bluepanda", "data")).WithLogger(nil))
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

	svr.app.Post("/insert/:type/:uuid", handleInserts(log, db, PKS{}))
	svr.app.Post("/fetch/:type/:uuid", handleFetch(log, db))

	return svr, nil
}

func (s server) Listen(port string) error {
	return s.app.Listen(port)
}

func (s server) Cleanup(log logging.Logger) error {
	dbg := strings.Builder{}
	s.db.DumpTo(&dbg)
	log.Debug().Msg(dbg.String())
	s.db.Close()
	return nil
}

func (s server) Shutdown() error {
	return s.app.Shutdown()
}

func (s server) ShutdownWithTimeout(d time.Duration) error {
	return s.app.ShutdownWithTimeout(d)
}
