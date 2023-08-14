package service

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tauraamui/kvs/v2"
	"github.com/tauraamui/redpanda/internal/logging"
)

type typedEntry struct {
	t reflect.Type
	e kvs.Entry
}

func handleFetch(log logging.Logger, store kvs.KVDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := map[string]string{}
		json.Unmarshal(c.Body(), &data)

		blankEntries := convertToBlankTypesEntries(ttype, resolveOwnerID(uuidx), uint32(0), data)

		dest := []rawData{}
		for _, ent := range blankEntries {
			// iterate over all stored values for this entry
			prefix := ent.e.PrefixKey()
			if err := store.View(func(txn *badger.Txn) error {
				it := txn.NewIterator(badger.DefaultIteratorOptions)
				defer it.Close()

				var destinationindex uint32 = 0
				for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
					item := it.Item()
					if err := item.Value(func(val []byte) error {
						ent.e.Data = val
						return nil
					}); err != nil {
						return err
					}

					if len(dest) == 0 || destinationindex >= uint32(len(dest)) {
						dest = append(dest, rawData{})
					}

					v := reflect.New(ent.t).Interface()
					fmt.Printf("LEN: %d\n", len(ent.e.Data))
					if err := convertFromBytes(ent.e.Data, v); err != nil {
						return err
					}
					fmt.Printf("V: %+v\n", v)
					dest[destinationindex][ent.e.ColumnName] = v

					destinationindex++
				}
				return nil
			}); err != nil {
				return err
			}
		}

		log.Info().Msg("loaded entry successfully...")

		return c.JSON(dest)
	}
}

type rawData map[string]any

func handleInserts(log logging.Logger, store kvs.KVDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := rawData{}
		json.Unmarshal(c.Body(), &data)

		entries := convertToEntries(ttype, resolveOwnerID(uuidx), uint32(0), data, true)

		for _, entry := range entries {
			if err := kvs.Store(store, entry); err != nil {
				log.Error().Msgf("failed to store entry: %v", err)
				return c.SendStatus(http.StatusInternalServerError)
			}
		}

		log.Info().Msg("stored entry successfully...")

		return nil
	}
}

func resolveOwnerID(v string) kvs.UUID {
	if v == "root" {
		return kvs.RootOwner{}
	}
	return uuid.MustParse(v)
}

func loadItemDataIntoEntry(ent *kvs.Entry, fn func(func(val []byte) error) error) error {
	return fn(func(val []byte) error {
		return convertFromBytes(val, &ent.Data)
	})
}

func convertToBlankTypesEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data map[string]string) []typedEntry {
	entries := []typedEntry{}
	for k, v := range data {
		e := kvs.Entry{
			TableName:  tableName,
			ColumnName: strings.ToLower(k),
			OwnerUUID:  ownerUUID,
			RowID:      rowID,
		}

		entries = append(entries, typedEntry{t: nameToTypeOf(v), e: e})
	}
	return entries
}

func nameToTypeOf(name string) reflect.Type {
	switch name {
	case "string":
		return reflect.TypeOf("")
	case "int":
		return reflect.TypeOf(int(0))
	}
	return nil
}

func convertToBlankEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data map[string]any) []kvs.Entry {
	return convertToEntries(tableName, ownerUUID, rowID, data, false)
}

func convertToEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data map[string]any, includeData bool) []kvs.Entry {
	entries := []kvs.Entry{}

	for k, v := range data {
		e := kvs.Entry{
			TableName:  tableName,
			ColumnName: strings.ToLower(k),
			OwnerUUID:  ownerUUID,
			RowID:      rowID,
		}

		if includeData {
			bd, err := convertToBytes(v)
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
	switch v := i.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case int32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case int64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(v))
		return buf, nil
	case uint:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case uint32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, v)
		return buf, nil
	case uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, v)
		return buf, nil
	case float32:
		bits := math.Float32bits(v)
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, bits)
		return buf, nil
	case float64:
		bits := math.Float64bits(v)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, bits)
		return buf, nil
	case bool:
		if v {
			return []byte{1}, nil
		} else {
			return []byte{0}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

func convertFromBytes(data []byte, i interface{}) error {
	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	switch v := i.(type) {
	case *[]byte:
		*v = data
		return nil
	case *string:
		*v = string(data)
		return nil
	case *int:
		if len(data) == 2 {
			data = append(data, 0x0000)
			data = append(data, 0x0000)
		}
		if len(data) < 4 {
			return fmt.Errorf("insufficient data for int")
		}
		*v = int(binary.BigEndian.Uint32(data))
		return nil
	case *int32:
		if len(data) < 4 {
			return fmt.Errorf("insufficient data for int32")
		}
		*v = int32(binary.BigEndian.Uint32(data))
		return nil
	case *int64:
		if len(data) < 8 {
			return fmt.Errorf("insufficient data for int64")
		}
		*v = int64(binary.BigEndian.Uint64(data))
		return nil
	case *uint:
		if len(data) < 4 {
			return fmt.Errorf("insufficient data for uint")
		}
		*v = uint(binary.BigEndian.Uint32(data))
		return nil
	case *uint32:
		if len(data) < 4 {
			return fmt.Errorf("insufficient data for uint32")
		}
		*v = binary.BigEndian.Uint32(data)
		return nil
	case *uint64:
		if len(data) < 8 {
			return fmt.Errorf("insufficient data for uint64")
		}
		*v = binary.BigEndian.Uint64(data)
		return nil
	case *float32:
		if len(data) < 4 {
			return fmt.Errorf("insufficient data for float32")
		}
		bits := binary.BigEndian.Uint32(data)
		*v = math.Float32frombits(bits)
		return nil
	case *float64:
		if len(data) < 8 {
			return fmt.Errorf("insufficient data for float64")
		}
		bits := binary.BigEndian.Uint64(data)
		*v = math.Float64frombits(bits)
		return nil
	case *bool:
		if len(data) < 1 {
			return fmt.Errorf("insufficient data for bool")
		}
		*v = data[0] != 0
		return nil
	default:
		return fmt.Errorf("unsupported type")
	}
}
