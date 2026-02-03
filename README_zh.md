# SysOCR

[![Go Reference](https://pkg.go.dev/badge/github.com/zn-chen/sysocr.svg)](https://pkg.go.dev/github.com/zn-chen/sysocr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

跨平台 OCR 库，调用操作系统原生 OCR 接口进行文字识别。

[English](README.md)

## 特性

- **跨平台支持**: macOS 和 Windows 10/11
- **原生 API**: 使用系统内置 OCR 引擎，无需额外依赖
- **简单易用**: 统一的 Go API，一行代码完成识别
- **多种输入**: 支持本地文件、远程 URL、内存数据
- **位置信息**: 返回文字内容及其在图片中的位置

## 平台支持

| 平台 | OCR 引擎 | 最低版本 |
|------|----------|----------|
| macOS | Vision Framework | macOS 10.15+ |
| Windows | Windows.Media.Ocr | Windows 10 |

## 安装

```bash
go get github.com/zn-chen/sysocr
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/zn-chen/sysocr"
)

func main() {
    // 从文件识别
    result, err := sysocr.Recognize(sysocr.Options{
        Input: sysocr.Input{FilePath: "/path/to/image.png"},
    })
    if err != nil {
        panic(err)
    }

    // 输出识别文本
    fmt.Println(result.Text)

    // 输出每个文本块的位置
    for _, block := range result.Blocks {
        fmt.Printf("Text: %s\n", block.Text)
        fmt.Printf("Position: (%.2f, %.2f) Size: %.2f x %.2f\n",
            block.BoundingBox.X, block.BoundingBox.Y,
            block.BoundingBox.Width, block.BoundingBox.Height)
    }
}
```

## API 文档

### Recognize

```go
func Recognize(opts Options) (*Result, error)
```

执行 OCR 识别。

### Options

```go
type Options struct {
    Input     Input
    Languages []string // 可选：语言提示（如 "zh-Hans", "en"）
}
```

### Input

指定图片来源，三个字段只能设置其中一个：

```go
type Input struct {
    FilePath string // 本地文件路径
    URL      string // 远程 URL (http/https)
    Data     []byte // 内存中的图片数据
}
```

### Result

```go
type Result struct {
    Blocks []TextBlock // 识别到的文本块列表
    Text   string      // 所有文本拼接
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

位置坐标归一化到 0-1 范围：

```go
type BoundingBox struct {
    X      float64 // 左上角 X (0-1)
    Y      float64 // 左上角 Y (0-1)
    Width  float64 // 宽度 (0-1)
    Height float64 // 高度 (0-1)
}
```

## 使用示例

### 从文件识别

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{FilePath: "/path/to/image.png"},
})
```

### 从 URL 识别

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{URL: "https://example.com/image.png"},
})
```

### 从内存数据识别

```go
imageData, _ := os.ReadFile("image.png")
result, err := sysocr.Recognize(sysocr.Options{
    Input: sysocr.Input{Data: imageData},
})
```

### 指定识别语言

```go
result, err := sysocr.Recognize(sysocr.Options{
    Input:     sysocr.Input{FilePath: "image.png"},
    Languages: []string{"zh-Hans", "en"},
})
```

## 运行示例

```bash
# 从本地文件
go run ./examples/basic /path/to/image.png

# 从 URL
go run ./examples/basic https://example.com/image.png
```

## 构建

```bash
# macOS
go build ./...

# Windows（无需特殊配置）
go build ./...
```

## 技术实现

### macOS

- 使用 Vision Framework 的 `VNRecognizeTextRequest`
- 通过 CGO 调用 Objective-C 代码
- 返回行级别的文本块

### Windows

- 使用 Windows.Media.Ocr API
- 纯 Go 实现，通过 syscall 调用 WinRT
- 无需 CGO，无需 MSVC 编译器
- 返回行级别的文本块

## 许可证

MIT License
