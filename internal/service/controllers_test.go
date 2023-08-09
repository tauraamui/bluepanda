package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/matryer/is"
	"github.com/tauraamui/kvs/v2"
	"github.com/tauraamui/kvs/v2/storage"
	"github.com/tauraamui/redpanda/internal/logging"
	"github.com/tauraamui/redpanda/internal/mock"
)

type data struct {
	Name string
	Size uint32
}

func TestHandleInserts(t *testing.T) {
	register, store, test, shutdown := setup()
	defer shutdown()

	is := is.New(t)

	logWriter := mock.LogWriter{}
	register("POST", "/:type/:uuid", handleInserts(logging.New(&logWriter), store))

	resp, err := test(buildPostRequest("/fruit/root", mustMarshal(data{
		Name: "mango",
		Size: 99,
	})))

	is.NoErr(err)

	is.Equal(resp.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(resp.Body)
	is.NoErr(err)

	is.Equal(string(body), "")
}

func setup() (register, storage.Store, test, func() error) {
	app := fiber.New()
	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}

	store := storage.New(db)

	return app.Add, storage.New(db), app.Test, func() error {
		app.Shutdown()
		store.Close()
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
