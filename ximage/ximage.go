package ximage

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

var (
	ErrInvalidImage    = errors.New("invalid image data")
	ErrUnsupportedType = errors.New("unsupported image type")
)

// CompressImage compresses the given image byte slice.
// It supports JPEG and PNG formats.
// The quality parameter is used for JPEG compression (1-100).
// For PNG, it's ignored and the default compression is used.
// The maxSizeInMB parameter specifies the maximum size of the compressed image in megabytes.
func CompressImage(data []byte, quality int, maxSizeInMB float64) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrInvalidImage
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	quality = normalizeQuality(quality)

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, ErrUnsupportedType
	}

	if err != nil {
		return nil, err
	}

	maxSizeInBytes := int64(maxSizeInMB * 1024 * 1024)
	if buf.Len() > int(maxSizeInBytes) && (format == "jpeg" || format == "jpg") {
		return compressJPEGWithSizeLimit(img, quality, maxSizeInBytes)
	}

	return buf.Bytes(), nil
}

// CompressImageReader compresses the image from the given reader.
// It supports JPEG and PNG formats.
// The quality parameter is used for JPEG compression (1-100).
// For PNG, it's ignored and the default compression is used.
// The maxSizeInMB parameter specifies the maximum size of the compressed image in megabytes.
func CompressImageReader(r io.Reader, quality int, maxSizeInMB float64) ([]byte, error) {
	if r == nil {
		return nil, ErrInvalidImage
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return CompressImage(data, quality, maxSizeInMB)
}

func normalizeQuality(quality int) int {
	if quality < 1 {
		return 1
	}
	if quality > 100 {
		return 100
	}
	return quality
}

func compressJPEGWithSizeLimit(img image.Image, quality int, maxSizeInBytes int64) ([]byte, error) {
	var buf bytes.Buffer
	for quality > 1 {
		buf.Reset()
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
		if buf.Len() <= int(maxSizeInBytes) {
			break
		}
		quality -= 5
	}
	return buf.Bytes(), nil
}

func IsSupported(format string) bool {
	switch format {
	case "jpeg", "jpg", "png":
		return true
	default:
		return false
	}
}

func GetImageFormat(data []byte) (string, error) {
	if len(data) == 0 {
		return "", ErrInvalidImage
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return format, nil
}
