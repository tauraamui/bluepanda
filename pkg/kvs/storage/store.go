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

package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v3"
	pb "github.com/tauraamui/bluepanda/pkg/api"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Value interface {
	TableName() string
}

type Store struct {
	db   kvs.KVDB
	bpdb pb.BluePandaClient
	pks  map[string]*badger.Sequence
}

func New(db kvs.KVDB) Store {
	return Store{db: db, pks: map[string]*badger.Sequence{}}
}

func Connect(addr string) (*Store, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Store{bpdb: pb.NewBluePandaClient(conn)}, nil
}

func (s Store) Save(owner kvs.UUID, value Value) error {
	if s.bpdb != nil {
		ridResp, err := s.bpdb.NextRowID(context.Background(), &pb.NextRowIDRequest{Uuid: owner.String(), Type: value.TableName()})
		if err != nil {
			return err
		}
		ridResp.GetId()
		// call "save value" but have it invoke an RPC store method
	}
	rowID, err := nextRowID(s.db, owner, value.TableName(), s.pks)
	if err != nil {
		return err
	}

	return saveValue(s.db, value.TableName(), owner, rowID, value)
}

func (s Store) Update(owner kvs.UUID, value Value, rowID uint32) error {
	return saveValue(s.db, value.TableName(), owner, rowID, value)
}

func saveValue(db kvs.KVDB, tableName string, ownerID kvs.UUID, rowID uint32, v Value) error {
	if v == nil {
		return nil
	}
	entries := kvs.ConvertToEntries(tableName, ownerID, rowID, v)
	for _, e := range entries {
		if err := kvs.Store(db, e); err != nil {
			return err
		}
	}

	return kvs.LoadID(v, rowID)
}

func (s Store) Delete(owner kvs.UUID, value Value, rowID uint32) error {
	db := s.db

	blankEntries := kvs.ConvertToBlankEntries(value.TableName(), owner, rowID, value)
	for _, ent := range blankEntries {
		db.Update(func(txn *badger.Txn) error {
			return txn.Delete(ent.Key())
		})
	}

	return nil
}

func Load[T Value](s Store, dest T, owner kvs.UUID, rowID uint32) error {
	db := s.db

	blankEntries := kvs.ConvertToBlankEntries(dest.TableName(), owner, rowID, dest)
	for _, ent := range blankEntries {
		db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(ent.Key())
			if err != nil {
				return err
			}

			if err := item.Value(func(val []byte) error {
				ent.Data = val
				return nil
			}); err != nil {
				return err
			}

			return nil
		})

		ent.RowID = rowID

		if err := kvs.LoadEntry(dest, ent); err != nil {
			return err
		}
	}

	return kvs.LoadID(dest, rowID)
}

func extractRowFromKey(k string) (int, error) {
	rowPos := strings.LastIndex(k, ".")
	return strconv.Atoi(k[rowPos+1:])
}

func LoadAll[T Value](s Store, owner kvs.UUID) ([]T, error) {
	return loadAllWithPredicate[T](s, owner, func(e kvs.Entry) bool { return true })
}

func LoadAllWithEvaluator[T Value](s Store, owner kvs.UUID, pred func(e kvs.Entry) bool) ([]T, error) {
	return loadAllWithPredicate[T](s, owner, pred)
}

func loadAllWithPredicate[T Value](s Store, owner kvs.UUID, pred func(e kvs.Entry) bool) ([]T, error) {
	db := s.db
	dest := []T{}
	v := *new(T)

	exclusions := map[int]int{}

	blankEntries := kvs.ConvertToBlankEntries(v.TableName(), owner, 0, v)
	for _, ent := range blankEntries {
		// iterate over all stored values for this entry
		prefix := ent.PrefixKey()
		if err := db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()

			var destinationindex uint32 = 0
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				if _, ok := exclusions[int(destinationindex)]; ok {
					destinationindex++
					continue
				}

				item := it.Item()
				if err := loadItemDataIntoEntry(&ent, item.Value); err != nil {
					return err
				}

				if err := forEachEntryItem(destinationindex, ent, item.Key(), &dest, &exclusions, pred); err != nil {
					return err
				}

				destinationindex++
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	finalDest := []T{}
	for i, v := range dest {
		if _, ok := exclusions[i]; ok {
			continue
		}
		finalDest = append(finalDest, v)
	}

	return finalDest, nil
}

func (s Store) Close() (err error) {
	if s.pks == nil {
		return
	}
	for _, seq := range s.pks {
		if seq == nil {
			continue
		}
		// TODO:(tauraamui) should collect all errors into group here rather than immediately escape
		if err = seq.Release(); err != nil {
			return
		}
	}

	s.pks = nil

	return
}

func forEachEntryItem[T Value](index uint32, ent kvs.Entry, key []byte, dest *[]T, exclusions *map[int]int, pred func(e kvs.Entry) bool) error {
	if len(*dest) == 0 || index >= uint32(len(*dest)) {
		rowID, err := extractRowFromKey(string(key))
		if err != nil {
			return err
		}

		*dest = append(*dest, *new(T))

		// for reasons, we have to just keep assigning the current "field we're on" as the full entry's ID
		if err := kvs.LoadID(&(*dest)[index], uint32(rowID)); err != nil {
			return err
		}
	}

	ent.RowID = index

	excluded := false
	if pred != nil && !pred(ent) {
		(*exclusions)[int(index)] = int(index)
		excluded = true
	}

	if !excluded {
		if err := kvs.LoadEntry(&(*dest)[index], ent); err != nil {
			return err
		}
	}
	return nil
}

func loadItemDataIntoEntry(ent *kvs.Entry, fn func(func(val []byte) error) error) error {
	return fn(func(val []byte) error {
		ent.Data = val
		return nil
	})
}

func nextRowID(db kvs.KVDB, owner kvs.UUID, tableName string, pks map[string]*badger.Sequence) (uint32, error) {
	seq, err := resolveSequence(db, fmt.Sprintf("%s.%s", owner, tableName), pks)
	if err != nil {
		return 0, err
	}

	s, err := seq.Next()
	if err != nil {
		return 0, err
	}
	return uint32(s), nil
}

func nextSequence(seq *badger.Sequence) (uint32, error) {
	s, err := seq.Next()
	if err != nil {
		return 0, err
	}
	return uint32(s), nil
}

func resolveSequence(db kvs.KVDB, sequenceKey string, pks map[string]*badger.Sequence) (*badger.Sequence, error) {
	seq, ok := pks[sequenceKey]
	var err error
	if !ok {
		seq, err = db.GetSeq([]byte(sequenceKey), 1)
		if err != nil {
			return nil, err
		}
		pks[sequenceKey] = seq
	}

	return seq, nil
}
