package cache

import (
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

func (ifk *ImageFileKey) FsPath() string {
	widthStr := strconv.Itoa(ifk.width)
	heightStr := strconv.Itoa(ifk.height)
	qualityStr := strconv.Itoa(int(ifk.quality))

	return strings.Join([]string{ifk.filename, widthStr, heightStr, qualityStr, ifk.format}, "_")
}
