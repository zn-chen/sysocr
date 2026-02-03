package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zn-chen/sysocr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basic <image_path>")
		fmt.Println("       basic <image_url>")
		os.Exit(1)
	}

	input := os.Args[1]

	var opts sysocr.Options

	// 检测输入是 URL 还是文件路径
	if len(input) > 7 && (input[:7] == "http://" || input[:8] == "https://") {
		opts.Input.URL = input
	} else {
		opts.Input.FilePath = input
	}

	// 执行 OCR 识别并计时
	startTime := time.Now()
	result, err := sysocr.Recognize(opts)
	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Fprintf(os.Stderr, "OCR failed: %v\n", err)
		os.Exit(1)
	}

	// 输出耗时
	fmt.Printf("=== OCR Time: %v ===\n\n", elapsed)

	// 输出结果
	fmt.Println("=== Recognized Text ===")
	fmt.Println(result.Text)
	fmt.Println()

	fmt.Println("=== Text Blocks with Positions ===")
	for i, block := range result.Blocks {
		fmt.Printf("[%d] Text: %s\n", i+1, block.Text)
		fmt.Printf("    Position: (%.4f, %.4f) Size: %.4f x %.4f\n",
			block.BoundingBox.X, block.BoundingBox.Y,
			block.BoundingBox.Width, block.BoundingBox.Height)
		fmt.Println()
	}
}
