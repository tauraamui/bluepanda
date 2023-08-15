package service

import (
	"bytes"
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
	"github.com/tauraamui/redpanda/internal/logging"
	"github.com/tauraamui/redpanda/pkg/kvs"
)

const JSONNumber = byte(99)

type typedEntry struct {
	t reflect.Type
	e kvs.Entry
}

func handleFetch(log logging.Logger, store kvs.KVDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := []string{}
		log.Debug().Msgf("%s", c.Body())
		json.Unmarshal(c.Body(), &data)

		blankEntries := convertToBlankTypesEntries(ttype, resolveOwnerID(uuidx), uint32(0), data)

		dest := []rawData{}
		for _, ent := range blankEntries {
			// iterate over all stored values for this entry
			prefix := ent.PrefixKey()
			if err := store.View(func(txn *badger.Txn) error {
				it := txn.NewIterator(badger.DefaultIteratorOptions)
				defer it.Close()

				var destinationindex uint32 = 0
				for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
					item := it.Item()
					if err := item.Value(func(val []byte) error {
						ent.Data = val
						return nil
					}); err != nil {
						return err
					}
					ent.Meta = item.UserMeta()

					if len(dest) == 0 || destinationindex >= uint32(len(dest)) {
						dest = append(dest, rawData{})
					}

					var v any
					if ent.Meta != JSONNumber {
						v = reflect.New(reflect.TypeOf(createInstanceOfKind(reflect.Kind(ent.Meta)))).Interface()
						if err := convertFromBytes(ent.Data, v); err != nil {
							return err
						}
					} else {
						v = json.Number(string(ent.Data))
					}
					dest[destinationindex][ent.ColumnName] = v

					destinationindex++
				}
				return nil
			}); err != nil {
				return err
			}
		}

		log.Debug().Msg("loaded entry successfully...")

		return c.JSON(dest)
	}
}

type PKS map[string]*badger.Sequence

type rawData map[string]any

func handleInserts(log logging.Logger, store kvs.KVDB, gpks PKS) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := rawData{}
		decoder := json.NewDecoder(bytes.NewReader(c.Body()))
		decoder.UseNumber()
		if err := decoder.Decode(&data); err != nil {
			return err
		}

		rowID, err := nextRowID(store, resolveOwnerID(uuidx), ttype, gpks)
		if err != nil {
			return err
		}
		entries := convertToEntries(ttype, resolveOwnerID(uuidx), rowID, data, true)

		for _, entry := range entries {
			if err := kvs.Store(store, entry); err != nil {
				log.Error().Msgf("failed to store entry: %v", err)
				return c.SendStatus(http.StatusInternalServerError)
			}
		}

		log.Debug().Msg("stored entry successfully...")

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

func convertToBlankTypesEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data []string) []kvs.Entry {
	entries := []kvs.Entry{}
	for _, k := range data {
		e := kvs.Entry{
			TableName:  tableName,
			ColumnName: strings.ToLower(k),
			OwnerUUID:  ownerUUID,
			RowID:      rowID,
		}

		entries = append(entries, e)
	}
	return entries
}

func convertToBlankEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data map[string]any) []kvs.Entry {
	return convertToEntries(tableName, ownerUUID, rowID, data, false)
}

func convertToEntries(tableName string, ownerUUID kvs.UUID, rowID uint32, data map[string]any, includeData bool) []kvs.Entry {
	entries := []kvs.Entry{}

	for k, v := range data {
		jsonNum, isJSONNumber := v.(json.Number)
		e := kvs.Entry{
			TableName:  tableName,
			ColumnName: strings.ToLower(k),
			OwnerUUID:  ownerUUID,
			RowID:      rowID,
			Meta:       byte(reflect.TypeOf(v).Kind()),
		}

		if isJSONNumber {
			e.Meta = JSONNumber
		}

		if includeData {
			if !isJSONNumber {

				bd, err := convertToBytes(v)
				if err != nil {
					return entries
				}
				e.Data = bd
			} else {
				e.Data = []byte(jsonNum.String())
			}
		}

		entries = append(entries, e)
	}

	return entries
}

func createInstanceOfKind(kind reflect.Kind) any {
	switch kind {
	case reflect.Bool:
		return false
	case reflect.Int:
		return int(0)
	case reflect.Int8:
		return int8(0)
	case reflect.Int16:
		return int16(0)
	case reflect.Int32:
		return int32(0)
	case reflect.Int64:
		return int64(0)
	case reflect.Uint:
		return uint(0)
	case reflect.Uint8:
		return uint8(0)
	case reflect.Uint16:
		return uint16(0)
	case reflect.Uint32:
		return uint32(0)
	case reflect.Uint64:
		return uint64(0)
	case reflect.Uintptr:
		return uintptr(0)
	case reflect.Float32:
		return float32(0)
	case reflect.Float64:
		return float64(0)
	case reflect.Complex64:
		return complex64(0)
	case reflect.Complex128:
		return complex128(0)
	case reflect.Interface:
		return new(interface{})
	case reflect.String:
		return ""
	default:
		return nil
	}
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
	case json.Number:
		return []byte(v), nil
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

func nextRowID(db kvs.KVDB, owner kvs.UUID, tableName string, pks map[string]*badger.Sequence) (uint32, error) {
	seq, err := resolveSequence(db, fmt.Sprintf("%s.%s", owner, tableName), pks)
	if err != nil {
		return 0, err
	}

	s, err := seq.Next()
	if err != nil {
		return 0, err
	}
	return uint32(s), nil
}

func nextSequence(seq *badger.Sequence) (uint32, error) {
	s, err := seq.Next()
	if err != nil {
		return 0, err
	}
	return uint32(s), nil
}

func resolveSequence(db kvs.KVDB, sequenceKey string, pks map[string]*badger.Sequence) (*badger.Sequence, error) {
	seq, ok := pks[sequenceKey]
	var err error
	if !ok {
		seq, err = db.GetSeq([]byte(sequenceKey), 1)
		if err != nil {
			return nil, err
		}
		pks[sequenceKey] = seq
	}

	return seq, nil
}
