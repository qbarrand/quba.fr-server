package cache

import "io"

type WriterToCloser interface {
	io.Closer
	io.WriterTo
}

type Cache interface {
	Add(Key, io.Reader, string, string) error
	Get(Key) (io.ReadCloser, Metadata, error)
}

type Key interface {
	Key() string
}

type Metadata interface {
	MainColor() string
	Hash() string
}
