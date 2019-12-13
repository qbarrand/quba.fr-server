package cache

import (
	"path/filepath"
	"strconv"
	"strings"
)

type ImageFileKey struct {
	filename string
	format   string
	height   int
	quality  uint
	width    int
}

func NewImageFileKey(f string, w, h int, q uint, fmt string) *ImageFileKey {
	return &ImageFileKey{
		filename: f,
		format:   fmt,
		height:   h,
		quality:  q,
		width:    w,
	}
}

func (ifk *ImageFileKey) Key() string {
	ext := filepath.Ext(ifk.filename)
	name := strings.TrimSuffix(ifk.filename, ext)

	widthStr := strconv.Itoa(ifk.width)
	heightStr := strconv.Itoa(ifk.height)
	qualityStr := strconv.Itoa(int(ifk.quality))

	return strings.Join([]string{name, widthStr, heightStr, qualityStr, ifk.format}, "_") + ext
}
