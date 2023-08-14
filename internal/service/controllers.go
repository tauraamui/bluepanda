package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tauraamui/kvs/v2"
	"github.com/tauraamui/redpanda/internal/logging"
)

func handleFetch(log logging.Logger, store kvs.KVDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := map[string]any{}
		json.Unmarshal(c.Body(), &data)

		blankEntries := convertToBlankEntries(ttype, resolveOwnerID(uuidx), uint32(0), data)

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

					if len(dest) == 0 || destinationindex >= uint32(len(dest)) {
						dest = append(dest, rawData{})
					}

					dest[destinationindex][ent.ColumnName] = ent.Data

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
	default:
		// Use json.Unmarshal to convert the []byte to the interface.
		return json.Unmarshal(data, v)
	}
}
