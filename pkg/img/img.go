package img

import (
	"fmt"
	"log"

	"gopkg.in/gographics/imagick.v2/imagick"
)

func Resize(mw *imagick.MagickWand, height, width, quality uint, format string) error {
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

	if height != 0 || width != 0 {
		oHeight := mw.GetImageHeight()
		oWidth := mw.GetImageWidth()

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

		if err := mw.AdaptiveResizeImage(width, height); err != nil {
			return fmt.Errorf("Could not resize the image to %dx%d: %v", width, height, err)
		}
	}

	//
	// quality
	//

	currentQuality := mw.GetImageCompressionQuality()

	if quality < currentQuality {
		log.Printf("Lowering the quality from %d to %d", currentQuality, quality)

		if err := mw.SetImageCompressionQuality(quality); err != nil {
			return fmt.Errorf("Could not set the quality to %d: %v", quality, err)
		}
	}

	//
	// Strip EXIF data
	//

	if err := mw.StripImage(); err != nil {
		return fmt.Errorf("Could not strip metadata: %v", err)
	}

	//
	// Interlace
	//

	if err := mw.SetInterlaceScheme(imagick.INTERLACE_JPEG); err != nil {
		return fmt.Errorf("Could not set the interlace method: %v", err)
	}

	//
	// Color space
	//

	if err := mw.SetColorspace(imagick.COLORSPACE_SRGB); err != nil {
		return fmt.Errorf("Could not set the color space: %v", err)
	}

	if format != "" {
		if err := mw.SetFormat(format); err != nil {
			return fmt.Errorf("Could not set the format to %q: %v", format, err)
		}
	}

	return nil
}

func GetMainColor(mw *imagick.MagickWand) (uint, uint, uint, error) {
	c := mw.Clone()

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
