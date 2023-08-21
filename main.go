package main

import (
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

type Fruit struct {
	Name string
}

func (f Fruit) TableName() string { return "fruits" }

func main() {
	/*
		store, err := bluepanda.Connect(":3000")
		if err != nil {
			panic(err)
		}
		defer store.Close()
	*/

	store, err := storage.Connect(":3000")

	if err != nil {
		panic(err)
	}
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Fruit{Name: "mango"})

}
