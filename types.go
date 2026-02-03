package sysocr

// BoundingBox 表示文本在图片中的位置。
type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// TextBlock 表示识别到的文本块及其位置信息。
type TextBlock struct {
	Text        string
	BoundingBox BoundingBox
}

// Result 包含 OCR 识别结果。
type Result struct {
	Blocks []TextBlock
	Text   string // 所有文本拼接
}

// Input 指定图片来源，三个字段只能设置其中一个。
type Input struct {
	FilePath string // 本地文件路径
	URL      string // 远程 URL (http/https)
	Data     []byte // 内存中的图片数据
}

// Options 配置 OCR 识别参数。
type Options struct {
	Input     Input
	Languages []string // 可选：语言提示（如 "zh-Hans", "en"）
}
