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

package query

import (
	"github.com/tauraamui/bluepanda/pkg/kvs"
	"github.com/tauraamui/bluepanda/pkg/kvs/storage"
)

type Query struct {
	filters []Filter
}

type operator int64

const (
	undefined operator = iota
	equal
	lessthan
)

func (op operator) String() string {
	switch op {
	case equal:
		return "equal"
	default:
		return "undefined"
	}
}

type Filter struct {
	q         *Query
	fieldName string
	op        operator
	values    []any
}

func (f Filter) cmp(d []byte) bool {
	for _, v := range f.values {
		if kvs.CompareBytesToAny(d, v) {
			return true
		}
	}
	return false
}

func New() *Query {
	return &Query{}
}

func Run[T storage.Value](s storage.Store, owner kvs.UUID, q *Query) ([]T, error) {
	return storage.LoadAllWithEvaluator[T](s, owner, func(e kvs.Entry) bool {
		if q == nil || len(q.filters) == 0 {
			return true
		}

		captured := true
		for i, filter := range q.filters {
			if i > 0 && !captured {
				return false
			}
			if filter.fieldName == e.ColumnName {
				if filter.op == equal {
					if !filter.cmp(e.Data) {
						captured = false
					}
				}
			}
		}

		return captured
	})
}

func (q *Query) Filter(fieldName string) *Filter {
	q = q.clone()
	filter := Filter{q: q, fieldName: fieldName}
	q.filters = append(q.filters, filter)
	return &q.filters[len(q.filters)-1]
}

func (f *Filter) Eq(value ...any) *Query {
	f.values = value
	f.op = equal
	return f.q
}

func (f *Filter) Lt(value ...any) *Query {
	f.values = value
	f.op = lessthan
	return f.q
}

func (q *Query) clone() *Query {
	x := *q
	// Copy the contents of the slice-typed fields to a new backing store.
	if len(q.filters) > 0 {
		x.filters = make([]Filter, len(q.filters))
		copy(x.filters, q.filters)
	}
	return &x
}
