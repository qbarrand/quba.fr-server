package cache

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
)

type FsCache string

func New(path string) FsCache {
	return FsCache(path)
}

func (c FsCache) Add(key Key, r io.Reader, mainColor string) error {
	path := c.getDataFile(key)

	fd, err := os.OpenFile(path, os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := fnv.New64a()

	w := io.MultiWriter(fd, h)

	_, err = bufio.NewReader(r).WriteTo(w)
	if err != nil {
		return err
	}

	m := &jsonMetadata{
		HashStr:      hex.EncodeToString(h.Sum(nil)),
		MainColorStr: mainColor,
	}

	return m.writeFile(getMetadataFile(path))
}

func (c FsCache) Get(key Key) (io.ReadCloser, Metadata, error) {
	path := c.getDataFile(key)

	fd, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	m, err := readJSONMetadata(getMetadataFile(path))
	if err != nil {
		// Close fd here
		fd.Close()

		return nil, nil, err
	}

	return fd, m, nil
}

func (c FsCache) getDataFile(key Key) string {
	return filepath.Join(string(c), key.Key())
}

func getMetadataFile(path string) string {
	return path + ".meta.json"
}

type jsonMetadata struct {
	HashStr      string `json:"hash"`
	MainColorStr string `json:"main_color"`
}

func (m *jsonMetadata) Hash() string {
	return m.HashStr
}

func (m *jsonMetadata) MainColor() string {
	return m.MainColorStr
}

func (m *jsonMetadata) writeFile(path string) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}

	return json.NewEncoder(fd).Encode(m)
}

func readJSONMetadata(path string) (*jsonMetadata, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	m := &jsonMetadata{}

	return m, json.NewDecoder(fd).Decode(m)
}
