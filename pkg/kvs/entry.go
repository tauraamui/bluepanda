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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

type Entry struct {
	TableName  string
	ColumnName string
	OwnerUUID  UUID
	RowID      uint32
	Data       []byte
	Meta       byte
}

func (e Entry) PrefixKey() []byte {
	return []byte(fmt.Sprintf("%s.%s.%s", e.TableName, e.ColumnName, e.resolveOwnerID()))
}

func (e Entry) Key() []byte {
	return []byte(fmt.Sprintf("%s.%s.%s.%d", e.TableName, e.ColumnName, e.resolveOwnerID(), e.RowID))
}

func (e Entry) resolveOwnerID() string {
	if e.OwnerUUID == nil {
		e.OwnerUUID = RootOwner{}
	}
	return e.OwnerUUID.String()
}

func Store(db KVDB, e Entry) error {
	return db.conn.Update(func(txn *badger.Txn) error {
		be := badger.NewEntry([]byte(e.Key()), e.Data)
		return txn.SetEntry(be.WithMeta(e.Meta))
	})
}

func Get(db KVDB, e *Entry) error {
	return db.conn.View(func(txn *badger.Txn) error {
		lookupKey := e.Key()
		item, err := txn.Get(lookupKey)
		if err != nil {
			return fmt.Errorf("%s: %s", strings.ToLower(err.Error()), lookupKey)
		}

		if err := item.Value(func(val []byte) error {
			e.Data = val
			return nil
		}); err != nil {
			return err
		}
		e.Meta = item.UserMeta()

		return nil
	})
}

func ConvertToBlankEntries(tableName string, ownerID UUID, rowID uint32, x any) []Entry {
	v := reflect.ValueOf(x)
	return convertToEntries(tableName, ownerID, rowID, v, false)
}

func ConvertToEntries(tableName string, ownerID UUID, rowID uint32, x any) []Entry {
	v := reflect.ValueOf(x)
	return convertToEntries(tableName, ownerID, rowID, v, true)
}

type UUID interface {
	String() string
}

type RootOwner struct{}

func (o RootOwner) String() string { return "root" }

func LoadEntry(s interface{}, entry Entry) error {
	// convert the interface value to a reflect.Value so we can access its fields
	val := reflect.ValueOf(s).Elem()

	field, err := resolveFieldRef(val, entry.ColumnName)
	if err != nil {
		return err
	}

	// convert the entry's Data field to the type of the target field
	if err := convertFromBytes(entry.Data, field.Addr().Interface()); err != nil {
		return fmt.Errorf("failed to convert entry data to field type: %v", err)
	}

	return nil
}

func LoadID(s any, rowID uint32) error {
	// convert the interface value to a reflect.Value so we can access its fields
	val := reflect.ValueOf(s).Elem()

	field, err := resolveFieldRef(val, "ID")
	if err != nil {
		return err
	}

	// convert the entry's Data field to the type of the target field
	if err := assignUint32(rowID, field.Addr().Interface()); err != nil {
		return fmt.Errorf("failed to convert entry data to field type: %v", err)
	}

	return nil
}

func resolveFieldRef(v reflect.Value, nameToMatch string) (reflect.Value, error) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if strings.EqualFold(field.Name, nameToMatch) {
			return v.Field(i), nil
		}
	}

	return reflect.Zero(reflect.TypeOf(v)), fmt.Errorf("struct does not have a field with name %q", nameToMatch)
}

func LoadEntries(s interface{}, entries []Entry) error {
	for _, entry := range entries {
		if err := LoadEntry(s, entry); err != nil {
			return err
		}
	}

	return nil
}

func convertToEntries(tableName string, ownerUUID UUID, rowID uint32, v reflect.Value, includeData bool) []Entry {
	entries := []Entry{}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		vv := reflect.Indirect(v)
		f := vv.Type().Field(i)

		fOpts := resolveFieldOptions(f)
		if fOpts.Ignore {
			continue
		}

		e := Entry{
			TableName:  tableName,
			ColumnName: strings.ToLower(f.Name),
			OwnerUUID:  ownerUUID,
			RowID:      rowID,
		}

		if includeData {
			bd, err := convertToBytes(v.Field(i).Interface())
			if err != nil {
				return entries
			}
			e.Data = bd
		}

		entries = append(entries, e)
	}

	return entries
}

func convertToBytes(i interface{}) ([]byte, error) {
	// Check the type of the interface.
	switch v := i.(type) {
	case []byte:
		// Return the input as a []byte if it is already a []byte.
		return v, nil
	case string:
		// Convert the string to a []byte and return it.
		return []byte(v), nil
	default:
		// Use json.Marshal to convert the interface to a []byte.
		return json.Marshal(v)
	}
}

func assignUint32(data uint32, dest any) error {
	// Check that the destination argument is a pointer.
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	switch v := dest.(type) {
	case *uint32:
		*v = data
		return nil
	}

	return errors.New("struct field ID is not of type uint32")
}

func CompareBytesToAny(a []byte, i interface{}) bool {
	switch v := i.(type) {
	case []byte:
		return bytes.Equal(a, v)
	case string:
		return string(a) == v
	default:
		val := reflect.ValueOf(i)
		if val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		newVal := reflect.New(val.Type())
		err := json.Unmarshal(a, newVal.Interface())
		if err != nil {
			return false
		}
		return reflect.DeepEqual(newVal.Elem().Interface(), i)
	}
}

func convertFromBytes(data []byte, i interface{}) error {
	// Check that the destination argument is a pointer.
	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	// Check the type of the interface.
	switch v := i.(type) {
	case *[]byte:
		// Set the value of the interface to the []byte if it is a pointer to a []byte.
		*v = data
		return nil
	case *string:
		// Convert the []byte to a string and set the value of the interface to the string.
		*v = string(data)
		return nil
	case *UUID:
		// Convert the []byte to a UUID instance and set the value of the interface to it.
		uuidv, err := uuid.ParseBytes(data)
		if err != nil {
			return err
		}
		*v = uuidv
		return nil
	default:
		// Use json.Unmarshal to convert the []byte to the interface.
		return json.Unmarshal(data, v)
	}
}

type mdbFieldOptions struct {
	Ignore bool
}

func resolveFieldOptions(f reflect.StructField) mdbFieldOptions {
	mdbTagValue := f.Tag.Get("mdb")
	return mdbFieldOptions{
		Ignore: strings.Contains(mdbTagValue, "ignore"),
	}
}
