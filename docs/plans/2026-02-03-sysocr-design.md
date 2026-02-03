# SysOCR 设计文档

## 概述

SysOCR 是一个跨平台的 Go OCR 库，通过调用操作系统原生 OCR 接口实现文字识别。

- **支持平台**: macOS, Windows 10/11
- **技术栈**: Go → CGO → C/Objective-C
- **发布形式**: Go 库 (`go get github.com/zn-chen/sysocr`)

## 项目结构

```
sysocr/
├── ocr.go              # 公共 API 定义
├── types.go            # 数据类型定义
├── input.go            # 输入处理 (文件/URL/内存)
├── ocr_darwin.go       # macOS 实现入口
├── ocr_windows.go      # Windows 实现入口
├── internal/
│   ├── darwin/         # macOS CGO 实现
│   │   ├── ocr.go
│   │   ├── ocr.h
│   │   └── ocr.m       # Vision Framework
│   └── windows/        # Windows CGO 实现
│       ├── ocr.go
│       ├── ocr.h
│       └── ocr.c       # Windows.Media.Ocr
├── go.mod
└── LICENSE
```

## API 设计

```go
package sysocr

type BoundingBox struct {
    X      float64
    Y      float64
    Width  float64
    Height float64
}

type TextBlock struct {
    Text        string
    BoundingBox BoundingBox
}

type Result struct {
    Blocks []TextBlock
    Text   string      // 所有文本拼接
}

type Input struct {
    FilePath string    // 本地文件路径
    URL      string    // 远程 URL (http/https)
    Data     []byte    // 内存数据
}

type Options struct {
    Input     Input
    Languages []string  // 可选：语言提示
}

func Recognize(opts Options) (*Result, error)
```

## 平台实现

### macOS (Vision Framework)

调用链: `Recognize()` → `internal/darwin/ocr.go` → CGO → `ocr.m`

使用:
- `VNRecognizeTextRequest` 进行文字识别
- `VNImageRequestHandler` 处理图片
- 返回 `VNRecognizedTextObservation`

### Windows (Windows.Media.Ocr)

调用链: `Recognize()` → `internal/windows/ocr.go` → CGO → `ocr.c`

使用:
- Windows Runtime (WinRT) C API
- `Windows.Media.Ocr.OcrEngine`
- 返回 `OcrResult`

## 平台兼容性

使用 Go build tags 控制，非 darwin/windows 平台编译���失败：

```go
//go:build darwin
//go:build windows
```

## 输入支持

1. **本地文件**: 直接读取文件路径
2. **远程 URL**: HTTP/HTTPS 下载后处理
3. **内存数据**: 直接处理 `[]byte`
