//go:build darwin

package sysocr

import (
	"strings"

	"github.com/zn-chen/sysocr/internal/darwin"
)

// Recognize 对提供的图片进行 OCR 识别。
func Recognize(opts Options) (*Result, error) {
	// 将输入转换为字节数据
	data, err := resolveInput(opts.Input)
	if err != nil {
		return nil, err
	}

	// 调用平台特定实现
	darwinResult, err := darwin.Recognize(data, opts.Languages)
	if err != nil {
		return nil, err
	}

	// 转换为公共类型
	result := &Result{
		Blocks: make([]TextBlock, len(darwinResult.Blocks)),
	}

	var textBuilder strings.Builder
	for i, b := range darwinResult.Blocks {
		result.Blocks[i] = TextBlock{
			Text: b.Text,
			BoundingBox: BoundingBox{
				X:      b.X,
				Y:      b.Y,
				Width:  b.Width,
				Height: b.Height,
			},
		}
		if i > 0 {
			textBuilder.WriteString("\n")
		}
		textBuilder.WriteString(b.Text)
	}
	result.Text = textBuilder.String()

	return result, nil
}
