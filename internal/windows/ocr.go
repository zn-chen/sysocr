//go:build windows

package windows

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lwindowsapp -lshcore -lole32

#include "ocr.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// TextBlock 表示识别到的文本块及其边界框。
type TextBlock struct {
	Text   string
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Result 包含 OCR 识别结果。
type Result struct {
	Blocks []TextBlock
}

// Recognize 使用 Windows.Media.Ocr API 对图片数据进行 OCR 识别。
func Recognize(data []byte, languages []string) (*Result, error) {
	if len(data) == 0 {
		return nil, errors.New("empty image data")
	}

	// 准备数据指针
	dataPtr := (*C.uchar)(unsafe.Pointer(&data[0]))
	dataLen := C.int(len(data))

	// 准备语言参数
	var langsPtr **C.char
	var langsCount C.int
	if len(languages) > 0 {
		cLangs := make([]*C.char, len(languages))
		for i, lang := range languages {
			cLangs[i] = C.CString(lang)
		}
		defer func() {
			for _, cl := range cLangs {
				C.free(unsafe.Pointer(cl))
			}
		}()
		langsPtr = &cLangs[0]
		langsCount = C.int(len(languages))
	}

	// 调用 C 函数
	cResult := C.sysocr_recognize(dataPtr, dataLen, langsPtr, langsCount)
	defer C.sysocr_free_result(cResult)

	// 检查错误
	if cResult.error != nil {
		return nil, errors.New(C.GoString(cResult.error))
	}

	// 转换结果
	result := &Result{
		Blocks: make([]TextBlock, int(cResult.count)),
	}

	if cResult.count > 0 && cResult.blocks != nil {
		blocks := unsafe.Slice(cResult.blocks, int(cResult.count))
		for i, b := range blocks {
			result.Blocks[i] = TextBlock{
				Text:   C.GoString(b.text),
				X:      float64(b.x),
				Y:      float64(b.y),
				Width:  float64(b.width),
				Height: float64(b.height),
			}
		}
	}

	return result, nil
}
