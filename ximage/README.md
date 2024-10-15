# ximage

ximage is a Go package that provides image compression functionality for JPEG and PNG formats.

## Features

- Compress JPEG and PNG images
- Adjustable quality for JPEG compression
- Maximum size limit for compressed images
- Support for both byte slices and io.Reader inputs

## Installation

```bash
go get github.com/seefs001/xox/ximage
```

## Usage

### CompressImage

Compresses an image from a byte slice.

```go
func CompressImage(data []byte, quality int, maxSizeInMB float64) ([]byte, error)
```

Parameters:
- `data`: The input image as a byte slice
- `quality`: Compression quality (1-100) for JPEG images (ignored for PNG)
- `maxSizeInMB`: Maximum size of the compressed image in megabytes

Returns:
- Compressed image as a byte slice
- Error, if any

Example:

```go
import (
    "fmt"
    "io/ioutil"
    "github.com/seefs001/xox/ximage"
)

func main() {
    // Read image file
    data, err := ioutil.ReadFile("input.jpg")
    if err != nil {
        fmt.Println("Error reading file:", err)
        return
    }

    // Compress image
    compressed, err := ximage.CompressImage(data, 75, 1.0)
    if err != nil {
        fmt.Println("Error compressing image:", err)
        return
    }

    // Save compressed image
    err = ioutil.WriteFile("output.jpg", compressed, 0644)
    if err != nil {
        fmt.Println("Error writing file:", err)
        return
    }

    fmt.Println("Image compressed successfully")
}
```

### CompressImageReader

Compresses an image from an io.Reader.

```go
func CompressImageReader(r io.Reader, quality int, maxSizeInMB float64) ([]byte, error)
```

Parameters:
- `r`: The input image as an io.Reader
- `quality`: Compression quality (1-100) for JPEG images (ignored for PNG)
- `maxSizeInMB`: Maximum size of the compressed image in megabytes

Returns:
- Compressed image as a byte slice
- Error, if any

Example:

```go
import (
    "fmt"
    "os"
    "io/ioutil"
    "github.com/seefs001/xox/ximage"
)

func main() {
    // Open image file
    file, err := os.Open("input.png")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // Compress image
    compressed, err := ximage.CompressImageReader(file, 75, 2.0)
    if err != nil {
        fmt.Println("Error compressing image:", err)
        return
    }

    // Save compressed image
    err = ioutil.WriteFile("output.png", compressed, 0644)
    if err != nil {
        fmt.Println("Error writing file:", err)
        return
    }

    fmt.Println("Image compressed successfully")
}
```

## Notes

- For PNG images, the `quality` parameter is ignored, and default compression is applied.
- If the compressed image exceeds the specified `maxSizeInMB`, the package will attempt to further reduce the quality of JPEG images to meet the size limit.
- For unsupported image formats, an error will be returned.
- The package uses the standard Go image libraries for decoding and encoding images.
- Compression is lossy for JPEG images and lossless for PNG images.
- The actual file size after compression may be slightly different from the specified `maxSizeInMB` due to the nature of image compression algorithms.

## Performance Considerations

- For large images or high volumes of images, consider using goroutines to process multiple images concurrently.
- The compression process can be CPU-intensive, especially for large images or when trying to meet strict size limits.

## Limitations

- The package currently supports only JPEG and PNG formats. Other formats will result in an error.
- There's no built-in progress reporting for long-running compressions.
- The package does not preserve EXIF data or other metadata during compression.

## Contributing

Contributions to improve ximage are welcome. Please submit issues and pull requests on the GitHub repository.

## License

This package is distributed under the MIT license. See the LICENSE file for more details.
