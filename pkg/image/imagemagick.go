package image

import (
	"fmt"
	"log"

	"gopkg.in/gographics/imagick.v2/imagick"
)

type ImageMagickProcessor struct {
	mw *imagick.MagickWand
}

func NewImagickProcessor(path string) (*ImageMagickProcessor, error) {
	mw := imagick.NewMagickWand()
	err := mw.ReadImage(path)

	return &ImageMagickProcessor{mw: mw}, err
}

func (imp *ImageMagickProcessor) Bytes() []byte {
	return imp.mw.GetImageBlob()
}

func (imp *ImageMagickProcessor) Convert(format string) error {
	return imp.mw.SetFormat(format)
}

func (imp *ImageMagickProcessor) Destroy() {
	imp.mw.Destroy()
}

func (imp *ImageMagickProcessor) ExifField(name string) string {
	return imp.mw.GetImageProperty(name)
}

func (imp *ImageMagickProcessor) Format() string {
	return imp.mw.GetFormat()
}

func (imp *ImageMagickProcessor) MainColor() (uint, uint, uint, error) {
	c := imp.mw.Clone()
	defer c.Destroy()

	if err := c.SetDepth(8); err != nil {
		return 0, 0, 0, fmt.Errorf("could not set the color depth: %v", err)
	}

	if err := c.ScaleImage(1, 1); err != nil {
		return 0, 0, 0, fmt.Errorf("could not scale the image: %v", err)
	}

	_, histo := c.GetImageHistogram()

	p := histo[0]

	r := uint(p.GetRed() * 255)
	g := uint(p.GetGreen() * 255)
	b := uint(p.GetBlue() * 255)

	return r, g, b, nil
}

func (imp *ImageMagickProcessor) Resize(height, width uint) error {
	//
	// Sampling factor
	//

	// if err := mw.SetSamplingFactors([]float64{4, 2, 0}); err != nil {
	// 	log.Printf("Could not set the sampling factors: %v", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	//
	// Resizing
	//

	oHeight := imp.mw.GetImageHeight()
	oWidth := imp.mw.GetImageWidth()

	if width != 0 {
		ratio := float64(width) / float64(oWidth)
		height = uint(float64(oHeight) * ratio)
		goto resize
	}

	if height != 0 {
		ratio := float64(height) / float64(oHeight)
		width = uint(float64(oWidth) * ratio)
		goto resize
	}

resize:
	log.Printf("Resizing to %dx%d", width, height)

	if err := imp.mw.AdaptiveResizeImage(width, height); err != nil {
		return fmt.Errorf("Could not resize the image to %dx%d: %v", width, height, err)
	}

	//
	// Interlace
	//

	//if err := mw.SetInterlaceScheme(imagick.INTERLACE_JPEG); err != nil {
	//	return fmt.Errorf("Could not set the interlace method: %v", err)
	//}

	//
	// Color space
	//

	//if err := mw.SetColorspace(imagick.COLORSPACE_SRGB); err != nil {
	//	return fmt.Errorf("Could not set the color space: %v", err)
	//}

	return nil
}

func (imp *ImageMagickProcessor) SetQuality(quality uint) error {
	currentQuality := imp.mw.GetImageCompressionQuality()

	if quality < currentQuality {
		log.Printf("Lowering the quality from %d to %d", currentQuality, quality)

		if err := imp.mw.SetImageCompressionQuality(quality); err != nil {
			return fmt.Errorf("Could not set the quality to %d: %v", quality, err)
		}
	}

	return nil
}

func (imp *ImageMagickProcessor) StripEXIF() error {
	return imp.mw.StripImage()
}
