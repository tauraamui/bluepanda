package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

type SmallChild struct {
	ID           uint32 `mdb:"ignore"`
	UUID         kvs.UUID
	HungryMetric uint32
	Norished     bool
}

type Cake struct {
	ID       uint32 `mdb:"ignore"`
	UUID     kvs.UUID
	Type     string
	Calories int
}

func (b Cake) TableName() string { return "cakes" }

type Candle struct {
	Cake kvs.UUID
	ID   uint32 `mdb:"ignore"`
	Lit  bool
}

func (b Candle) TableName() string { return "candles" }

func hierarchy() {
	db, err := kvs.NewMemKVDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	child := SmallChild{
		UUID:         uuid.New(),
		HungryMetric: 100,
	}

	disguistingVeganCake := Cake{UUID: uuid.New(), Type: "INEDIBLE", Calories: -38}
	healthyishCarrotCake := Cake{UUID: uuid.New(), Type: "CARROT", Calories: 280}
	redVelvetCake := Cake{UUID: uuid.New(), Type: "RED_VELVET", Calories: 410}

	store.Save(child.UUID, &disguistingVeganCake)
	store.Save(child.UUID, &healthyishCarrotCake)
	store.Save(child.UUID, &redVelvetCake)

	store.Save(disguistingVeganCake.UUID, &Candle{Cake: disguistingVeganCake.UUID, Lit: true})
	store.Save(healthyishCarrotCake.UUID, &Candle{Cake: healthyishCarrotCake.UUID, Lit: true})
	store.Save(redVelvetCake.UUID, &Candle{Cake: redVelvetCake.UUID, Lit: true})

	bs, err := storage.LoadAll[Cake](store, child.UUID)
	for _, cake := range bs {
		fmt.Printf("ROWID: %d, %+v\n", cake.ID, cake)

		candles, err := storage.LoadAll[Cake](store, cake.UUID)
		if err != nil {
			panic(err)
		}

		for _, candle := range candles {
			fmt.Printf("ROWID: %d, %+v\n", candle.ID, candle)
		}
	}
}
