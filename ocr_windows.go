//go:build windows

package sysocr

import (
	"errors"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/zn-chen/sysocr/internal/winrt"
)

// 忽略未使用的语言参数（暂时使用用户配置语言）
var _ = func(languages []string) {}

// Recognize 对提供的图片进行 OCR 识别。
func Recognize(opts Options) (*Result, error) {
	// 将输入转换为字节数据
	data, err := resolveInput(opts.Input)
	if err != nil {
		return nil, err
	}

	// 初始化 Windows Runtime
	if err := winrt.Initialize(); err != nil {
		return nil, err
	}

	// 执行 OCR
	return recognizeWithWinRT(data, opts.Languages)
}

// recognizeWithWinRT 使用 Windows Runtime API 进行 OCR
func recognizeWithWinRT(data []byte, languages []string) (*Result, error) {
	// 创建 InMemoryRandomAccessStream
	stream, err := winrt.CreateInMemoryRandomAccessStream()
	if err != nil {
		return nil, errors.New("failed to create memory stream: " + err.Error())
	}
	defer stream.Release()

	// 创建 DataWriter 并写入数据
	writerFactory, err := winrt.GetDataWriterFactory()
	if err != nil {
		return nil, errors.New("failed to get DataWriter factory: " + err.Error())
	}
	defer writerFactory.Release()

	writer, err := writerFactory.CreateDataWriter(stream)
	if err != nil {
		return nil, errors.New("failed to create DataWriter: " + err.Error())
	}
	defer writer.Release()

	// 写入图片数据
	writer.WriteBytes(data)

	// 存储数据
	storeOp, err := writer.StoreAsync()
	if err != nil {
		return nil, errors.New("failed to store data: " + err.Error())
	}
	if err := storeOp.Wait(30 * time.Second); err != nil {
		return nil, errors.New("store operation timeout: " + err.Error())
	}
	storeOp.Release()

	// 刷新
	flushOp, err := writer.FlushAsync()
	if err != nil {
		return nil, errors.New("failed to flush: " + err.Error())
	}
	if err := flushOp.Wait(30 * time.Second); err != nil {
		return nil, errors.New("flush operation timeout: " + err.Error())
	}
	flushOp.Release()

	writer.DetachStream()

	// 将流位置重置到开始（通过 Seek）
	streamVtbl := stream.VTable()
	_, _, _ = syscallN(streamVtbl.Seek, uintptr(unsafe.Pointer(stream)), 0)

	// 创建 BitmapDecoder
	decoderStatics, err := winrt.GetBitmapDecoderStatics()
	if err != nil {
		return nil, errors.New("failed to get BitmapDecoder: " + err.Error())
	}
	defer decoderStatics.Release()

	// 调用 CreateAsync
	createAsyncVtbl := decoderStatics.VTable()
	var createOp *winrt.IAsyncOperation
	hr, _, _ := syscallN(createAsyncVtbl.CreateAsync,
		uintptr(unsafe.Pointer(decoderStatics)),
		uintptr(unsafe.Pointer(stream)),
		uintptr(unsafe.Pointer(&createOp)))
	if hr != 0 || createOp == nil {
		return nil, errors.New("failed to create BitmapDecoder async operation")
	}
	defer createOp.Release()

	if err := createOp.Wait(30 * time.Second); err != nil {
		return nil, errors.New("BitmapDecoder creation timeout: " + err.Error())
	}

	decoderInsp, err := createOp.GetResults()
	if err != nil || decoderInsp == nil {
		return nil, errors.New("failed to get BitmapDecoder result")
	}
	defer decoderInsp.Release()

	// 获取 IBitmapFrameWithSoftwareBitmap 接口
	var frameWithBitmap *winrt.IBitmapFrameWithSoftwareBitmap
	var ptr unsafe.Pointer
	decoderInsp.QueryInterface(&winrt.IID_IBitmapFrameWithSoftwareBitmap, &ptr)
	if ptr == nil {
		return nil, errors.New("failed to get IBitmapFrameWithSoftwareBitmap interface")
	}
	frameWithBitmap = (*winrt.IBitmapFrameWithSoftwareBitmap)(ptr)
	defer frameWithBitmap.Release()

	// 获取 SoftwareBitmap
	frameVtbl := frameWithBitmap.VTable()
	var getBitmapOp *winrt.IAsyncOperation
	hr, _, _ = syscallN(frameVtbl.GetSoftwareBitmapAsync,
		uintptr(unsafe.Pointer(frameWithBitmap)),
		uintptr(unsafe.Pointer(&getBitmapOp)))
	if hr != 0 || getBitmapOp == nil {
		return nil, errors.New("failed to get SoftwareBitmap async operation")
	}
	defer getBitmapOp.Release()

	if err := getBitmapOp.Wait(30 * time.Second); err != nil {
		return nil, errors.New("GetSoftwareBitmap timeout: " + err.Error())
	}

	bitmapInsp, err := getBitmapOp.GetResults()
	if err != nil || bitmapInsp == nil {
		return nil, errors.New("failed to get SoftwareBitmap result")
	}
	defer bitmapInsp.Release()

	bitmap := (*winrt.ISoftwareBitmap)(unsafe.Pointer(bitmapInsp))
	imageWidth := float64(bitmap.GetPixelWidth())
	imageHeight := float64(bitmap.GetPixelHeight())

	// 创建 OCR 引擎
	ocrStatics, err := winrt.GetOcrEngineStatics()
	if err != nil {
		return nil, errors.New("failed to get OcrEngine statics: " + err.Error())
	}
	defer ocrStatics.Release()

	engine, err := ocrStatics.TryCreateFromUserProfileLanguages()
	if err != nil || engine == nil {
		return nil, errors.New("failed to create OCR engine: no supported language available")
	}
	defer engine.Release()

	// 执行 OCR
	recognizeOp, err := engine.RecognizeAsync(bitmap)
	if err != nil {
		return nil, errors.New("failed to start OCR: " + err.Error())
	}
	defer recognizeOp.Release()

	if err := recognizeOp.Wait(60 * time.Second); err != nil {
		return nil, errors.New("OCR timeout: " + err.Error())
	}

	ocrResultInsp, err := recognizeOp.GetResults()
	if err != nil || ocrResultInsp == nil {
		return nil, errors.New("failed to get OCR result")
	}
	defer ocrResultInsp.Release()

	ocrResult := (*winrt.IOcrResult)(unsafe.Pointer(ocrResultInsp))

	// 收集结果
	result := &Result{
		Blocks: make([]TextBlock, 0),
	}

	var textBuilder strings.Builder

	// 获取所有行
	lines, err := ocrResult.GetLines()
	if err != nil {
		return nil, errors.New("failed to get OCR lines: " + err.Error())
	}
	defer lines.Release()

	lineCount := lines.GetSize()
	for i := uint32(0); i < lineCount; i++ {
		lineInsp, err := lines.GetAt(i)
		if err != nil || lineInsp == nil {
			continue
		}

		line := (*winrt.IOcrLine)(unsafe.Pointer(lineInsp))

		// 获取行文本
		lineText := line.GetText()

		// 获取所有单词来计算行的边界框
		words, err := line.GetWords()
		if err != nil {
			line.Release()
			continue
		}

		// 计算行的边界框（合并所有单词的边界框）
		var minX, minY, maxX, maxY float32
		first := true
		wordCount := words.GetSize()
		for j := uint32(0); j < wordCount; j++ {
			wordInsp, err := words.GetAt(j)
			if err != nil || wordInsp == nil {
				continue
			}

			word := (*winrt.IOcrWord)(unsafe.Pointer(wordInsp))
			rect := word.GetBoundingRect()

			if first {
				minX = rect.X
				minY = rect.Y
				maxX = rect.X + rect.Width
				maxY = rect.Y + rect.Height
				first = false
			} else {
				if rect.X < minX {
					minX = rect.X
				}
				if rect.Y < minY {
					minY = rect.Y
				}
				if rect.X+rect.Width > maxX {
					maxX = rect.X + rect.Width
				}
				if rect.Y+rect.Height > maxY {
					maxY = rect.Y + rect.Height
				}
			}

			word.Release()
		}

		words.Release()

		// 创建行级别的文本块
		if !first { // 确保有至少一个单词
			block := TextBlock{
				Text: lineText,
				BoundingBox: BoundingBox{
					X:      float64(minX) / imageWidth,
					Y:      float64(minY) / imageHeight,
					Width:  float64(maxX-minX) / imageWidth,
					Height: float64(maxY-minY) / imageHeight,
				},
			}
			result.Blocks = append(result.Blocks, block)

			if textBuilder.Len() > 0 {
				textBuilder.WriteString("\n")
			}
			textBuilder.WriteString(lineText)
		}

		line.Release()
	}

	result.Text = textBuilder.String()
	return result, nil
}

// syscallN 是 syscall.SyscallN 的包装
func syscallN(trap uintptr, args ...uintptr) (r1, r2 uintptr, err error) {
	return syscall.SyscallN(trap, args...)
}
