package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/query"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

type HotAirBalloon struct {
	ID     uint32 `mdb:"ignore"`
	UUID   kvs.UUID
	Flying bool
	MaxCap int
}

func (b HotAirBalloon) TableName() string { return "hotairballoons" }

type Passenger struct {
	ID        uint32 `mdb:"ignore"`
	FirstName string
	Surname   string
	Married   bool
	Age       int
}

func (p Passenger) TableName() string { return "passengers" }

func queries() {
	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	blimp := HotAirBalloon{
		UUID:   uuid.New(),
		Flying: false,
		MaxCap: 5,
	}

	store.Save(kvs.RootOwner{}, &blimp)
	store.Save(blimp.UUID, &Passenger{FirstName: "Brian", Surname: "Hax", Age: 3})
	store.Save(blimp.UUID, &Passenger{FirstName: "Amy", Surname: "Hax", Age: 26, Married: true})
	store.Save(blimp.UUID, &Passenger{FirstName: "Mark", Surname: "West", Age: 58})
	store.Save(blimp.UUID, &Passenger{FirstName: "Rory", Surname: "Hax", Age: 27, Married: true})

	haxFamalam, err := query.Run[Passenger](store, blimp.UUID, query.New().Filter("surname").Eq("Hax"))
	if err != nil {
		panic(err)
	}

	for _, p := range haxFamalam {
		fmt.Printf("ROWID: %d, %+v\n", p.ID, p)
	}

}
