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

package kvs

import (
	"fmt"
	"io"
	"os"

	"github.com/dgraph-io/badger/v3"
)

type KVDB struct {
	conn *badger.DB
}

func NewKVDB(db *badger.DB) (KVDB, error) {
	return newKVDB(db)
}

func NewMemKVDB() (KVDB, error) {
	return newKVDB(nil)
}

func newKVDB(db *badger.DB) (KVDB, error) {
	if db == nil {
		db, err := badger.Open(badger.DefaultOptions("").WithLogger(nil).WithInMemory(true))
		if err != nil {
			return KVDB{}, err
		}
		return KVDB{conn: db}, nil
	}

	return KVDB{conn: db}, nil
}

func (db KVDB) GetSeq(key []byte, bandwidth uint64) (*badger.Sequence, error) {
	return db.conn.GetSequence(key, bandwidth)
}

func (db KVDB) View(f func(txn *badger.Txn) error) error {
	return db.conn.View(f)
}

func (db KVDB) Update(f func(txn *badger.Txn) error) error {
	return db.conn.Update(f)
}

func (db KVDB) DumpTo(w io.Writer) error {
	return db.conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Fprintf(w, "key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db KVDB) DumpToStdout() error {
	return db.DumpTo(os.Stdout)
}

func (db KVDB) Close() error {
	return db.conn.Close()
}
