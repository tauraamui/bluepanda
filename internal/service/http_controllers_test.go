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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/matryer/is"
	"github.com/tauraamui/bluepanda/internal/logging"
	"github.com/tauraamui/bluepanda/internal/mock"
	"github.com/tauraamui/bluepanda/pkg/kvs"
)

type data struct {
	Name string
	Size uint32
}

func TestHandleFetch(t *testing.T) {
	register, store, test, shutdown := setup()
	defer shutdown()

	is := is.New(t)

	is.NoErr(insertEntry(store, "fruit", "name", 0, []byte("mango"), reflect.String))
	is.NoErr(insertEntry(store, "fruit", "size", 0, func() []byte {
		bits := math.Float64bits(99.48)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, bits)
		return buf
	}(), reflect.Float64))

	is.NoErr(insertEntry(store, "fruit", "name", 1, []byte("strawberry"), reflect.String))
	is.NoErr(insertEntry(store, "fruit", "size", 1, func() []byte {
		bits := math.Float64bits(15)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, bits)
		return buf
	}(), reflect.Float64))

	is.NoErr(insertEntry(store, "fruit", "name", 2, []byte("grape"), reflect.String))
	is.NoErr(insertEntry(store, "fruit", "size", 2, []byte("n/a"), reflect.String))

	logWriter := mock.LogWriter{}
	register("POST", "/fetch/:type/:uuid", handleFetch(logging.New(&logWriter), store))

	resp, err := test(buildPostRequest("/fetch/fruit/root", mustMarshal([]string{"name", "size"})))

	is.NoErr(err)

	is.Equal(resp.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(resp.Body)
	is.NoErr(err)

	is.Equal(string(body), `[{"name":"mango","size":99.48},{"name":"strawberry","size":15},{"name":"grape","size":"n/a"}]`)
}

func TestHandleInserts(t *testing.T) {
	register, store, test, shutdown := setup()
	defer shutdown()

	is := is.New(t)

	logWriter := mock.LogWriter{}
	register("POST", "/insert/:type/:uuid", handleInserts(logging.New(&logWriter), store, PKS{}))

	resp, err := test(buildPostRequest("/insert/fruit/root", mustMarshal(data{
		Name: "mango",
		Size: 99,
	})))

	is.NoErr(err)

	is.Equal(resp.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(resp.Body)
	is.NoErr(err)

	is.Equal(string(body), "")
}

func insertEntry(store kvs.KVDB, tbl, col string, rID uint32, data []byte, meta reflect.Kind) error {
	return kvs.Store(store, kvs.Entry{
		TableName:  tbl,
		ColumnName: col,
		OwnerUUID:  kvs.RootOwner{},
		RowID:      rID,
		Data:       data,
		Meta:       byte(meta),
	})
}

func setup() (register, kvs.KVDB, test, func() error) {
	app := fiber.New()
	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}

	return app.Add, db, app.Test, func() error {
		app.Shutdown()
		db.Close()
		return nil
	}
}

func mustMarshal(v any) []byte {
	d, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to marshal data: %v\n", err)
		os.Exit(1)
	}

	return d
}

func buildPostRequest(url string, data []byte) *http.Request {
	req := httptest.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

type register func(method, path string, handlers ...fiber.Handler) fiber.Router
type test func(req *http.Request, msTimeout ...int) (*http.Response, error)
