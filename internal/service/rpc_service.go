package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/tauraamui/bluepanda/internal/logging"
	"github.com/tauraamui/bluepanda/pkg/api"
	pb "github.com/tauraamui/bluepanda/pkg/api"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"google.golang.org/grpc"
)

type rpcserver struct {
	pb.UnimplementedBluePandaServer
	rpcserver *grpc.Server
	db        kvs.KVDB
	gpks      PKS
}

func NewRPC(log logging.Logger) (Server, error) {
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

	return &rpcserver{db: db, gpks: PKS{}}, nil
}

func (s *rpcserver) Type() string {
	return "gRPC"
}

func (s *rpcserver) Listen(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.rpcserver = grpc.NewServer()
	pb.RegisterBluePandaServer(s.rpcserver, s)
	return s.rpcserver.Serve(lis)
}

func (s *rpcserver) Cleanup(log logging.Logger) error {
	dbg := strings.Builder{}
	s.db.DumpTo(&dbg)
	log.Debug().Msg(dbg.String())
	return s.db.Close()
}

func (s *rpcserver) Shutdown() error {
	s.rpcserver.GracefulStop()
	return nil
}

func (s *rpcserver) ShutdownWithTimeout(d time.Duration) error {
	s.rpcserver.GracefulStop()
	return nil
}

func (s *rpcserver) NextRowID(ctx context.Context, req *pb.NextRowIDRequest) (*pb.NextRowIDResult, error) {
	id, err := nextRowID(s.db, uuid.MustParse(req.GetUuid()), req.GetType(), s.gpks)
	if err != nil {
		return nil, err
	}
	return &pb.NextRowIDResult{Id: id}, nil
}

func (s *rpcserver) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResult, error) {
	if err := s.db.Update(func(txn *badger.Txn) error {
		be := badger.NewEntry([]byte(req.GetKey()), req.GetData())
		mt := func() byte {
			m := req.GetMeta()
			if len(m) == 0 {
				return byte(0)
			}
			return m[0]
		}
		return txn.SetEntry(be.WithMeta(mt()))
	}); err != nil {
		return nil, err
	}
	return &pb.StoreResult{Success: true}, nil
}

func (s *rpcserver) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResult, error) {
	result := pb.GetResult{}
	if err := s.db.View(func(txn *badger.Txn) error {
		lookupKey := req.GetKey()
		item, err := txn.Get([]byte(lookupKey))
		if err != nil {
			return fmt.Errorf("%s: %s", strings.ToLower(err.Error()), lookupKey)
		}

		if err := item.Value(func(val []byte) error {
			result.Data = val
			return nil
		}); err != nil {
			return err
		}
		result.Meta = []byte{item.UserMeta()}
		return nil
	}); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *rpcserver) Fetch(req *pb.FetchRequest, stream pb.BluePanda_FetchServer) error {
	ttype := req.GetType()
	uuidx := req.GetUuid()

	columns := req.GetColumns()

	blankEntries := convertToBlankTypesEntries(ttype, resolveOwnerID(uuidx), uint32(0), columns)

	dest := []rawData{}
	for _, ent := range blankEntries {
		// iterate over all stored values for this entry
		prefix := ent.PrefixKey()
		if err := s.db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			var destinationindex uint32 = 0
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				if err := item.Value(func(val []byte) error {
					ent.Data = val
					return nil
				}); err != nil {
					return err
				}
				ent.Meta = item.UserMeta()

				if len(dest) == 0 || destinationindex >= uint32(len(dest)) {
					dest = append(dest, rawData{})
				}

				var v any
				if ent.Meta != JSONNumber {
					v = reflect.New(reflect.TypeOf(createInstanceOfKind(reflect.Kind(ent.Meta)))).Interface()
					if err := convertFromBytes(ent.Data, v); err != nil {
						return err
					}
				} else {
					v = json.Number(string(ent.Data))
				}
				dest[destinationindex][ent.ColumnName] = v

				destinationindex++
			}
			return nil
		}); err != nil {
			return err
		}
	}

	for i := 0; i < len(dest); i++ {
		data, err := json.Marshal(dest[i])
		if err != nil {
			return err
		}
		if err := stream.Send(&api.FetchResult{
			Json: data,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *rpcserver) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResult, error) {
	ttype := req.GetType()
	uuidx := req.GetUuid()

	data := rawData{}
	decoder := json.NewDecoder(bytes.NewReader(req.GetJson()))
	decoder.UseNumber()

	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	rowID, err := nextRowID(s.db, resolveOwnerID(uuidx), ttype, s.gpks)
	if err != nil {
		return nil, err
	}
	entries := convertToEntries(ttype, resolveOwnerID(uuidx), rowID, data, true)

	for _, entry := range entries {
		if err := kvs.Store(s.db, entry); err != nil {
			return nil, fmt.Errorf("failed to store entry: %v", err)
		}
	}

	log.Debug().Msg("stored entry successfully...")

	return &pb.InsertResult{Status: "successful"}, nil
}
