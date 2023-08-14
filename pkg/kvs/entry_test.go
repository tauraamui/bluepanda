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

package kvs_test

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/tauraamui/redpanda/pkg/kvs"
)

func TestEntryStoreValuesInTable(t *testing.T) {
	is := is.New(t)

	e := kvs.Entry{
		TableName:  "users",
		ColumnName: "email",
		OwnerUUID:  uuid.UUID{},
		Data:       []byte{0x33},
		Meta:       byte(reflect.Float64),
	}

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	seq, err := db.GetSeq(e.PrefixKey(), 100)
	is.NoErr(err) // error occurred on getting db sequence
	defer seq.Release()

	id, err := seq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value

	e.RowID = uint32(id)

	is.NoErr(kvs.Store(db, e)) // error occurred when calling store

	newEntry := kvs.Entry{
		TableName:  e.TableName,
		ColumnName: e.ColumnName,
		OwnerUUID:  uuid.UUID{},
		RowID:      e.RowID,
		Data:       nil,
	}
	is.NoErr(kvs.Get(db, &newEntry))

	is.Equal(newEntry.Data, []byte{0x33})
	is.Equal(newEntry.Meta, byte(reflect.Float64))
}

type uuidstr string

func (u uuidstr) String() string { return string(u) }

func TestGettingEntryOutOfTableErrorIncorrectKey(t *testing.T) {
	is := is.New(t)

	owner := uuidstr("11")
	e := kvs.Entry{
		TableName:  "user",
		ColumnName: "email",
		OwnerUUID:  owner,
		Data:       []byte{0x33},
		Meta:       byte(reflect.String),
	}

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	seq, err := db.GetSeq(e.PrefixKey(), 100)
	is.NoErr(err) // error occurred on getting db sequence
	defer seq.Release()

	id, err := seq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value

	e.RowID = uint32(id)

	is.NoErr(kvs.Store(db, e)) // error occurred when calling store

	newEntry := kvs.Entry{
		TableName:  e.TableName,
		ColumnName: "emailz",
		OwnerUUID:  owner,
		RowID:      e.RowID,
		Data:       nil,
	}

	err = kvs.Get(db, &newEntry)
	is.True(err != nil)
	is.Equal(err.Error(), "key not found: user.emailz.11.0")
	is.Equal(newEntry.Data, nil)
	is.Equal(newEntry.Meta, uint8(0x0))
}

func TestConvertToEntries(t *testing.T) {
	is := is.New(t)

	source := struct {
		Foo string
		Bar int
	}{
		Foo: "Foo",
		Bar: 4,
	}

	owner := uuidstr("39")
	e := kvs.ConvertToEntries("test", owner, 0, source)
	is.Equal(len(e), 2)

	is = is.NewRelaxed(t)

	is.Equal(kvs.Entry{
		OwnerUUID:  owner,
		TableName:  "test",
		ColumnName: "foo",
		Data:       []byte{70, 111, 111},
	}, e[0])

	is.Equal(kvs.Entry{
		OwnerUUID:  owner,
		TableName:  "test",
		ColumnName: "bar",
		Data:       []byte{52},
	}, e[1])
}

func TestLoadEntriesIntoStruct(t *testing.T) {
	// Define a struct type to use for the test
	type TestStruct struct {
		Field1 string
		Field2 int
		Field3 bool
	}

	// Create a slice of Entry values to use as input
	entries := []kvs.Entry{
		{ColumnName: "field1", Data: []byte("hello")},
		{ColumnName: "field2", Data: []byte("123")},
		{ColumnName: "field3", Data: []byte("true")},
	}

	s := TestStruct{}

	is := is.New(t)

	is.NoErr(kvs.LoadEntries(&s, entries)) // LoadEntries returned an error
	// Check that the values of the TestStruct fields were updated correctly
	expected := TestStruct{Field1: "hello", Field2: 123, Field3: true}
	is.Equal(s, expected) // Use the Equal method of the is package to compare the values
}

func TestSequences(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	fruitEntry := kvs.Entry{
		TableName:  "fruits",
		ColumnName: "color",
	}

	chocolateEntry := kvs.Entry{
		TableName:  "chocolate",
		ColumnName: "flavour",
	}

	fruitSeq, err := db.GetSeq(fruitEntry.PrefixKey(), 100)
	is.NoErr(err) // error occurred on getting db sequence
	defer fruitSeq.Release()

	chocolateSeq, err := db.GetSeq(chocolateEntry.PrefixKey(), 100)
	is.NoErr(err) // error occurred on getting db sequence
	defer fruitSeq.Release()

	id, err := fruitSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(0))

	id, err = chocolateSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(0))

	id, err = fruitSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(1))

	id, err = chocolateSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(1))

	id, err = fruitSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(2))

	id, err = chocolateSeq.Next()
	is.NoErr(err) // error occurred when aquiring next iter value
	is.Equal(id, uint64(2))
}

func TestCompareStringWithBytes(t *testing.T) {
	is := is.New(t)

	input := []byte("hello")
	is.True(kvs.CompareBytesToAny(input, "hello"))
}

func TestCompareBytesWithBytes(t *testing.T) {
	is := is.New(t)

	input := []byte("{\"A\":5,\"B\":\"hello\"}")
	is.True(kvs.CompareBytesToAny(input, input))
}

func TestCompareBytesWithStruct(t *testing.T) {
	is := is.New(t)

	type TestStruct struct {
		A int
		B string
	}
	input := []byte("{\"A\":5,\"B\":\"hello\"}")
	is.True(kvs.CompareBytesToAny(input, TestStruct{A: 5, B: "hello"}))
}
