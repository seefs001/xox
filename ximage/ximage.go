package ximage

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

// CompressImage compresses the given image byte slice.
// It supports JPEG and PNG formats.
// The quality parameter is used for JPEG compression (1-100).
// For PNG, it's ignored and the default compression is used.
// The maxSizeInMB parameter specifies the maximum size of the compressed image in megabytes.
func CompressImage(data []byte, quality int, maxSizeInMB float64) ([]byte, error) {
	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Create a buffer to store the compressed image
	var buf bytes.Buffer

	// Compress based on the image format
	switch format {
	case "jpeg":
		if quality < 1 {
			quality = 1
		} else if quality > 100 {
			quality = 100
		}
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, img)
	default:
		// If the format is not supported, return the original data
		return data, nil
	}

	if err != nil {
		return nil, err
	}

	// Check if the compressed image exceeds the maximum size
	maxSizeInBytes := int64(maxSizeInMB * 1024 * 1024)
	if buf.Len() > int(maxSizeInBytes) {
		// If it exceeds, compress again with lower quality
		for quality > 1 && buf.Len() > int(maxSizeInBytes) {
			buf.Reset()
			quality -= 5
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
			if err != nil {
				return nil, err
			}
		}
	}

	// Return the compressed image as a byte slice
	return buf.Bytes(), nil
}

// CompressImageReader compresses the image from the given reader.
// It supports JPEG and PNG formats.
// The quality parameter is used for JPEG compression (1-100).
// For PNG, it's ignored and the default compression is used.
// The maxSizeInMB parameter specifies the maximum size of the compressed image in megabytes.
func CompressImageReader(r io.Reader, quality int, maxSizeInMB float64) ([]byte, error) {
	// Read all data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Use the CompressImage function to compress the data
	return CompressImage(data, quality, maxSizeInMB)
}
