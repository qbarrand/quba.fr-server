package cache

import "io"

type Cache interface {
	Add(Key, io.Reader, string) error
	Get(Key) (io.ReadCloser, Metadata, error)
}

type Key interface {
	Key() string
}

type Metadata interface {
	MainColor() string
	Hash() string
}
