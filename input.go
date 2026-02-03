package sysocr

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	ErrNoInput       = errors.New("sysocr: no input specified")
	ErrMultipleInput = errors.New("sysocr: multiple inputs specified, only one allowed")
)

// resolveInput 将 Input 转换为原始图片字节数据。
func resolveInput(input Input) ([]byte, error) {
	count := 0
	if input.FilePath != "" {
		count++
	}
	if input.URL != "" {
		count++
	}
	if input.Data != nil {
		count++
	}

	if count == 0 {
		return nil, ErrNoInput
	}
	if count > 1 {
		return nil, ErrMultipleInput
	}

	if input.Data != nil {
		return input.Data, nil
	}

	if input.FilePath != "" {
		return os.ReadFile(input.FilePath)
	}

	if input.URL != "" {
		return fetchURL(input.URL)
	}

	return nil, ErrNoInput
}

// fetchURL 从远程 URL 获取图片数据。
func fetchURL(url string) ([]byte, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, errors.New("sysocr: URL must start with http:// or https://")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("sysocr: failed to fetch URL: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}
