# SysOCR

[![Go Reference](https://pkg.go.dev/badge/github.com/zn-chen/sysocr.svg)](https://pkg.go.dev/github.com/zn-chen/sysocr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A cross-platform OCR library that uses native OS APIs for text recognition.

[中文文档](README_zh.md)

## Features

- **Cross-platform**: Supports macOS and Windows 10/11
- **Native APIs**: Uses built-in OS OCR engines, no external dependencies
- **Simple API**: Unified Go interface, recognize text in one line
- **Multiple inputs**: Local files, remote URLs, or in-memory data
- **Position info**: Returns text content with bounding box coordinates

## Platform Support

| Platform | OCR Engine | Minimum Version |
|----------|------------|-----------------|
| macOS | Vision Framework | macOS 10.15+ |
| Windows | Windows.Media.Ocr | Windows 10 |

## Installation

```bash
go get github.com/zn-chen/sysocr
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zn-chen/sysocr"
)

func main() {
    // Recognize from file
    result, err := sysocr.Recognize(sysocr.Options{
        Input: sysocr.Input{FilePath: "/path/to/image.png"},
    })
    if err != nil {
        panic(err)
    }

    // Print recognized text
    fmt.Println(result.Text)

    // Print text blocks with positions
    for _, block := range result.Blocks {
        fmt.Printf("Text: %s\n", block.Text)
        fmt.Printf("Position: (%.2f, %.2f) Size: %.2f x %.2f\n",
            block.BoundingBox.X, block.BoundingBox.Y,
            block.BoundingBox.Width, block.BoundingBox.Height)
    }
}
```

## API Reference

### Recognize

```go
func Recognize(opts Options) (*Result, error)
```

Performs OCR recognition on the provided image.

### Options

```go
type Options struct {
    Input     Input
    Languages []string // Optional: language hints (e.g., "zh-Hans", "en")
}
```

### Input

Specifies the image source. Only one field should be set:

```go
type Input struct {
    FilePath string // Local file path
    URL      string // Remote URL (http/https)
    Data     []byte // In-memory image data
}
```

### Result

```go
type Result struct {
    Blocks []TextBlock // List of recognized text blocks
    Text   string      // All text concatenated
}
```

### TextBlock

```go
type TextBlock struct {
    Text        string
    BoundingBox BoundingBox
}
```

### BoundingBox

Coordinates are normalized to 0-1 range:

```go
type BoundingBox struct {
    X      float64 // Top-left X (0-1)
    Y      float64 // Top-left Y (0-1)
    Width  float64 // Width (0-1)
    Height float64 // Height (0-1)
}
```

## Usage Examples

### From File

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{FilePath: "/path/to/image.png"},
})
```

### From URL

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{URL: "https://example.com/image.png"},
})
```

### From Memory

```go
imageData, _ := os.ReadFile("image.png")
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{Data: imageData},
})
```

### With Language Hints

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input:     sysocr.Input{FilePath: "image.png"},
    Languages: []string{"zh-Hans", "en"},
})
```

## Running Examples

```bash
# From local file
go run ./examples/basic /path/to/image.png

# From URL
go run ./examples/basic https://example.com/image.png
```

## Building

```bash
# macOS
go build ./...

# Windows (no special configuration needed)
go build ./...
```

## Technical Details

### macOS

- Uses Vision Framework's `VNRecognizeTextRequest`
- CGO bridge to Objective-C code
- Returns line-level text blocks

### Windows

- Uses Windows.Media.Ocr API
- Pure Go implementation via syscall to WinRT
- No CGO required, no MSVC compiler needed
- Returns line-level text blocks

## License

MIT License
