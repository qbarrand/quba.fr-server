//go:generate mockgen -source image_controller.go -destination mock_handlers/mock_image_controller.go ImageController

package handlers

type ImageController interface {
	Bytes() []byte
	Convert(string) error
	Destroy()
	Format() string
	MainColor() (uint, uint, uint, error)
	Resize(uint, uint) error
	SetQuality(uint) error
	StripEXIF() error
}
