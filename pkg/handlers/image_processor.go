package handlers

type imageProcessor interface {
	GetImageBlob() []byte
}
