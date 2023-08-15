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

	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/query"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

func BenchmarkQueryWithSingleFilterWithTwoRecords(b *testing.B) {
	db, err := kvs.NewMemKVDB()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("WHITE"))
	}
}

func BenchmarkQueryWithMultiFilterWithTwoRecords(b *testing.B) {
	db, err := kvs.NewMemKVDB()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	store.Save(kvs.RootOwner{}, &Balloon{Color: "WHITE", Size: 366})
	store.Save(kvs.RootOwner{}, &Balloon{Color: "RED", Size: 695})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("WHITE").Filter("size").Eq(306))
	}
}

func BenchmarkQueryWithMultiFilterWithFiveHunderedRecordsWithMatchingFilter(b *testing.B) {
	db, err := kvs.NewMemKVDB()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	store := storage.New(db)
	defer store.Close()

	for i := 0; i < 500; i++ {
		color := "RED"
		if i%2 == 0 {
			color = "BLUE"
		}
		store.Save(kvs.RootOwner{}, &Balloon{Color: color, Size: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.Run[Balloon](store, kvs.RootOwner{}, query.New().Filter("color").Eq("RED").Filter("size").Eq(306, 422, 211))
	}
}
