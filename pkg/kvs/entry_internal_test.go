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
	"testing"

	"github.com/matryer/is"
)

func TestConvertToBytes(t *testing.T) {
	is := is.New(t)

	// Test input and expected output.
	tests := []struct {
		input    interface{}
		expected []byte
	}{
		{[]byte{1, 2, 3}, []byte{1, 2, 3}},
		{"hello", []byte("hello")},
		{struct{ A int }{5}, []byte("{\"A\":5}")},
	}

	// Iterate over the tests and compare the output of convertToBytes to the expected output.
	for _, test := range tests {
		result, err := convertToBytes(test.input)
		is.NoErr(err)
		is.Equal(result, test.expected)
	}
}

func TestConvertBytesFromBytes(t *testing.T) {
	is := is.New(t)

	input := []byte{1, 2, 3}
	var destination []byte
	err := convertFromBytes(input, &destination)
	is.NoErr(err)
	is.Equal(destination, input)

}

func TestConvertStringFromBytes(t *testing.T) {
	is := is.New(t)

	input := []byte("hello")
	var destination string
	err := convertFromBytes(input, &destination)
	is.NoErr(err)
	is.Equal(destination, string(input))
}

func TestConvertStructFromBytes(t *testing.T) {
	is := is.New(t)

	type TestStruct struct {
		A int
		B string
	}
	input := []byte("{\"A\":5,\"B\":\"hello\"}")
	var destination TestStruct
	err := convertFromBytes(input, &destination)
	is.NoErr(err)
	is.Equal(destination, TestStruct{A: 5, B: "hello"})
}
