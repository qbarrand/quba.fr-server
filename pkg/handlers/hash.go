package handlers

import (
	"encoding/hex"
	"hash/fnv"
)

func hashBytes(b []byte) (string, error) {
	h := fnv.New64a()

	if _, err := h.Write(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
