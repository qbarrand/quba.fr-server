package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FsCache string

func New(path string) FsCache {
	return FsCache(path)
}

func (c FsCache) Add(key Key, r io.Reader, mainColor, hash string) error {
	path := c.getDataFile(key)

	// Create the parent directory
	if err := os.MkdirAll(filepath.Dir(path), os.ModeDir|0755); err != nil {
		return fmt.Errorf("could not create the parent directory: %v", err)
	}

	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	if _, err := bufio.NewReader(r).WriteTo(fd); err != nil {
		return err
	}

	m := &jsonMetadata{
		HashStr:      hash,
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
