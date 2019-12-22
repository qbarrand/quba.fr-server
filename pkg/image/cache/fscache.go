package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Item struct {
	Data      io.ReadCloser `json:"-"`
	Hash      string        `json:"hash"`
	MainColor string        `json:"main_color"`
}

func (i *Item) writeToFile(path string, data io.Reader) error {
	metadataPath := getMetadataFile(path)

	fdMeta, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("could not create file %s: %v", metadataPath, err)
	}
	defer fdMeta.Close()

	if err := json.NewEncoder(fdMeta).Encode(i); err != nil {
		return fmt.Errorf("could not encode: %v", err)
	}

	fdData, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file %s, %v", path, err)
	}
	defer fdData.Close()

	if _, err := bufio.NewReader(data).WriteTo(fdData); err != nil {
		return fmt.Errorf("could not write the data to %s: %v", path, err)
	}

	return nil
}

func readItem(path string) (*Item, error) {
	metadataPath := getMetadataFile(path)

	fdMeta, err := os.Open(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("could not read the metadata: %v", err)
	}
	defer fdMeta.Close()

	item := &Item{}

	if err := json.NewDecoder(fdMeta).Decode(item); err != nil {
		return nil, fmt.Errorf("could not decode the metadata at %s: %v", metadataPath, err)
	}

	fdData, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the data: %v", err)
	}

	item.Data = fdData

	return item, nil
}

func getMetadataFile(path string) string {
	return path + ".meta.json"
}

type Key interface {
	FsPath() string
}

type FsCache string

func (c FsCache) Add(key Key, r io.Reader, mainColor, hash string) error {
	path := c.filePath(key)

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

	item := &Item{
		Hash:      hash,
		MainColor: mainColor,
	}

	return item.writeToFile(path, r)
}

func (c FsCache) Get(key Key) (*Item, error) {
	return readItem(
		c.filePath(key),
	)
}

func (c FsCache) filePath(key Key) string {
	return filepath.Join(string(c), key.FsPath())
}
