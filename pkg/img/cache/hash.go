package cache

import (
	"encoding/hex"
	"hash/fnv"
)

func HashBytes(b []byte) (string, error) {
	h := fnv.New64a()

	if _, err := h.Write(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
