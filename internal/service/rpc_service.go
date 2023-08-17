package service

import (
	"encoding/json"
	"reflect"

	"github.com/dgraph-io/badger/v3"
	"github.com/tauraamui/bluepanda/pkg/api"
	pb "github.com/tauraamui/bluepanda/pkg/api"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"google.golang.org/grpc"
)

type rpcserver struct {
	pb.UnimplementedBluePandaServer
	store kvs.KVDB
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
		if err := s.store.View(func(txn *badger.Txn) error {
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

func stub() {
	s := grpc.NewServer()
	pb.RegisterBluePandaServer(s, &rpcserver{})
}
