package handlers

type imageController interface {
	Bytes() []byte
	Convert(string) error
	Destroy()
	ExifField(string) string
	Format() string
	MainColor() (uint, uint, uint, error)
	Resize(uint, uint) error
	SetQuality(uint) error
	StripEXIF() error
}
