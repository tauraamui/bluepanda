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

package query_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/query"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

type Balloon struct {
	ID    uint32 `mdb:"ignore"`
	Color string
	Size  int
}

func (b Balloon) TableName() string { return "balloons" }

func TestQueryFilterWithSinglePredicateSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("WHITE"))
	is.NoErr(err)
	is.Equal(len(bs), 1)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "WHITE")
	is.Equal(bs[0].Size, 366)

	is = is.New(t)

	bs, err = query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("RED"))
	is.NoErr(err)
	is.Equal(len(bs), 1)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "RED")
	is.Equal(bs[0].Size, 695)
}

func TestQueryFilterWithSinglePredicateMultipleValuesSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("GREEN", "WHITE", "CYAN", "PURPLE", "RED", "GOLD"))
	is.NoErr(err)
	is.Equal(len(bs), 2)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "WHITE")
	is.Equal(bs[0].Size, 366)

	is.Equal(bs[1].Color, "RED")
	is.Equal(bs[1].Size, 695)
}

func TestQueryFilterWithSinglePredicateFailure(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		is.NoErr(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("deef"))
	is.NoErr(err)
	is.Equal(len(bs), 0)

	bs, err = query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("rgrr"))
	is.NoErr(err)
	is.Equal(len(bs), 0)
}

func TestQueryFilterWithMultiplePredicateSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("WHITE").Filter("size").Eq(366))
	is.NoErr(err)
	is.Equal(len(bs), 1)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "WHITE")
	is.Equal(bs[0].Size, 366)

	is = is.New(t)

	bs, err = query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("RED").Filter("size").Eq(695))
	is.NoErr(err)
	is.Equal(len(bs), 1)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "RED")
	is.Equal(bs[0].Size, 695)
}

func TestQueryFilterWithMultiplePredicateMultipleValuesSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("ZIMA_BLUE", "BLACK", "WHITE", "RED").Filter("size").Eq(222, 366, 948))
	is.NoErr(err)
	is.Equal(len(bs), 1)
	is = is.NewRelaxed(t)
	is.Equal(bs[0].Color, "WHITE")
	is.Equal(bs[0].Size, 366)
}

func TestQueryFilterWithMultiplePredicateFailure(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("WHITE").Filter("size").Eq(110))
	is.NoErr(err)
	is.Equal(len(bs), 0)

	bs, err = query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("RED").Filter("size").Eq(548))
	is.NoErr(err)
	is.Equal(len(bs), 0)
}

func TestQueryFilterWithMultiplePredicateMultipleValuesFailure(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})

	bs, err := query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("GREY", "PINK").Filter("size").Eq(110, 366))
	is.NoErr(err)
	is.Equal(len(bs), 0)
}
