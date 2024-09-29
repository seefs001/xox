package ximage_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/seefs001/xox/ximage"
)

func TestCompressImage(t *testing.T) {
	tests := []struct {
		name        string
		imageFormat string
		quality     int
		maxSizeInMB float64
	}{
		{"JPEG_Normal", "jpeg", 75, 1.0},
		{"JPEG_HighQuality", "jpeg", 100, 2.0},
		{"JPEG_LowQuality", "jpeg", 10, 0.5},
		{"PNG_Normal", "png", 75, 1.0},
		{"JPEG_ExceedMaxSize", "jpeg", 100, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test image
			img := createTestImage(100, 100)

			// Encode the image
			var buf bytes.Buffer
			var err error
			switch tt.imageFormat {
			case "jpeg":
				err = jpeg.Encode(&buf, img, nil)
			case "png":
				err = png.Encode(&buf, img)
			}
			if err != nil {
				t.Fatalf("Failed to encode image: %v", err)
			}

			// Compress the image
			compressed, err := ximage.CompressImage(buf.Bytes(), tt.quality, tt.maxSizeInMB)
			if err != nil {
				t.Fatalf("CompressImage failed: %v", err)
			}

			// Check if the compressed image is within the max size
			maxSizeInBytes := int(tt.maxSizeInMB * 1024 * 1024)
			if len(compressed) > maxSizeInBytes {
				t.Errorf("Compressed image size (%d bytes) exceeds max size (%d bytes)", len(compressed), maxSizeInBytes)
			}

			// Verify that the compressed image can be decoded
			_, _, err = image.Decode(bytes.NewReader(compressed))
			if err != nil {
				t.Errorf("Failed to decode compressed image: %v", err)
			}
		})
	}
}

func TestCompressImageReader(t *testing.T) {
	tests := []struct {
		name        string
		imageFormat string
		quality     int
		maxSizeInMB float64
	}{
		{"JPEG_Normal", "jpeg", 75, 1.0},
		{"PNG_Normal", "png", 75, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test image
			img := createTestImage(100, 100)

			// Encode the image
			var buf bytes.Buffer
			var err error
			switch tt.imageFormat {
			case "jpeg":
				err = jpeg.Encode(&buf, img, nil)
			case "png":
				err = png.Encode(&buf, img)
			}
			if err != nil {
				t.Fatalf("Failed to encode image: %v", err)
			}

			// Create a reader from the buffer
			reader := bytes.NewReader(buf.Bytes())

			// Compress the image using CompressImageReader
			compressed, err := ximage.CompressImageReader(reader, tt.quality, tt.maxSizeInMB)
			if err != nil {
				t.Fatalf("CompressImageReader failed: %v", err)
			}

			// Check if the compressed image is within the max size
			maxSizeInBytes := int(tt.maxSizeInMB * 1024 * 1024)
			if len(compressed) > maxSizeInBytes {
				t.Errorf("Compressed image size (%d bytes) exceeds max size (%d bytes)", len(compressed), maxSizeInBytes)
			}

			// Verify that the compressed image can be decoded
			_, _, err = image.Decode(bytes.NewReader(compressed))
			if err != nil {
				t.Errorf("Failed to decode compressed image: %v", err)
			}
		})
	}
}

func TestCompressImage_UnsupportedFormat(t *testing.T) {
	// Create a dummy image data
	data := []byte("This is not a valid image")

	// Try to compress the invalid image data
	compressed, err := ximage.CompressImage(data, 75, 1.0)
	if err == nil {
		t.Error("Expected an error for unsupported format, but got nil")
	}
	if compressed != nil {
		t.Error("Expected nil compressed data for unsupported format")
	}
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, image.White)
		}
	}
	return img
}
