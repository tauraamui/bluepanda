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

package storage_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/redpanda/pkg/kvs"
	"github.com/tauraamui/redpanda/pkg/kvs/storage"
)

type Balloon struct {
	ID    uint32 `mdb:"ignore"`
	Color string
	Size  int
}

func (b Balloon) TableName() string { return "balloons" }

type Cake struct {
	ID       uint32 `mdb:"ignore"`
	Type     string
	Calories int
}

func (b Cake) TableName() string { return "cakes" }

func TestStoreAndLoadMultipleBalloonsSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}
	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	bs, err := storage.LoadAll[Balloon](store, kvs.RootOwner{})
	is.NoErr(err)

	is.Equal(len(bs), 3)

	is.Equal(bs[0], Balloon{ID: 0, Color: "RED", Size: 695})
	is.Equal(bs[1], Balloon{ID: 1, Color: "YELLOW", Size: 112})
	is.Equal(bs[2], Balloon{ID: 2, Color: "WHITE", Size: 366})
}

func TestStoreMultipleAndUpdateSingleBalloonsSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}

	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	smallYellowBalloon.Color = "PINK"
	is.NoErr(store.Update(kvs.RootOwner{}, &smallYellowBalloon, smallYellowBalloon.ID))

	bs, err := storage.LoadAll[Balloon](store, kvs.RootOwner{})
	is.NoErr(err)

	is.True(len(bs) == 3)

	is.Equal(bs[0], Balloon{ID: 0, Color: "RED", Size: 695})
	is.Equal(bs[1], Balloon{ID: 1, Color: "PINK", Size: 112})
	is.Equal(bs[2], Balloon{ID: 2, Color: "WHITE", Size: 366})
}

func TestStoreAndDeleteSingleBalloonSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}
	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	bs, err := storage.LoadAll[Balloon](store, kvs.RootOwner{})
	is.NoErr(err)

	is.True(len(bs) == 3)

	is.NoErr(store.Delete(kvs.RootOwner{}, &smallYellowBalloon, smallYellowBalloon.ID))

	bs, err = storage.LoadAll[Balloon](store, kvs.RootOwner{})
	is.NoErr(err)

	is.True(len(bs) == 2)
}

func TestStoreLoadMultipleLoadIndividualBalloonsSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}
	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	bs0 := Balloon{}
	is.NoErr(storage.Load(store, &bs0, kvs.RootOwner{}, 0))

	bs1 := Balloon{}
	is.NoErr(storage.Load(store, &bs1, kvs.RootOwner{}, 1))

	bs2 := Balloon{}
	is.NoErr(storage.Load(store, &bs2, kvs.RootOwner{}, 2))

	is.Equal(bs0, Balloon{ID: 0, Color: "RED", Size: 695})
	is.Equal(bs1, Balloon{ID: 1, Color: "YELLOW", Size: 112})
	is.Equal(bs2, Balloon{ID: 2, Color: "WHITE", Size: 366})
}

func TestStoreMultipleBalloonsSuccess(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}
	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	is.Equal(bigRedBalloon.ID, uint32(0))
	is.Equal(smallYellowBalloon.ID, uint32(1))
	is.Equal(mediumWhiteBalloon.ID, uint32(2))
}

func TestStoreMultipleBalloonsAndCakesInSuccessionRetainsCorrectRowIDs(t *testing.T) {
	is := is.New(t)

	db, err := kvs.NewMemKVDB()
	is.NoErr(err)
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	bigRedBalloon := Balloon{Color: "RED", Size: 695}
	disguistingVeganCake := Cake{Type: "INEDIBLE", Calories: -38}
	smallYellowBalloon := Balloon{Color: "YELLOW", Size: 112}
	healthyishCarrotCake := Cake{Type: "CARROT", Calories: 280}
	mediumWhiteBalloon := Balloon{Color: "WHITE", Size: 366}
	redVelvetCake := Cake{Type: "RED_VELVET", Calories: 410}

	is.NoErr(store.Save(kvs.RootOwner{}, &bigRedBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &smallYellowBalloon))
	is.NoErr(store.Save(kvs.RootOwner{}, &mediumWhiteBalloon))

	is.NoErr(store.Save(kvs.RootOwner{}, &disguistingVeganCake))
	is.NoErr(store.Save(kvs.RootOwner{}, &healthyishCarrotCake))
	is.NoErr(store.Save(kvs.RootOwner{}, &redVelvetCake))

	is.Equal(bigRedBalloon.ID, uint32(0))
	is.Equal(disguistingVeganCake.ID, uint32(0))
	is.Equal(smallYellowBalloon.ID, uint32(1))
	is.Equal(healthyishCarrotCake.ID, uint32(1))
	is.Equal(mediumWhiteBalloon.ID, uint32(2))
	is.Equal(redVelvetCake.ID, uint32(2))
}
