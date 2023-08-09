package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tauraamui/kvs/v2"
	"github.com/tauraamui/redpanda/internal/logging"
)

func handleInserts(log logging.Logger, store kvs.KVDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ttype := c.Params("type")
		uuidx := c.Params("uuid")

		data := map[string]any{}
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
