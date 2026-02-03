// Windows OCR 实现
// 使用 Windows.Media.Ocr API (WinRT) 进行文字识别

#include "ocr.h"

#include <windows.h>
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Foundation.Collections.h>
#include <winrt/Windows.Graphics.Imaging.h>
#include <winrt/Windows.Media.Ocr.h>
#include <winrt/Windows.Storage.Streams.h>
#include <shcore.h>

#include <string>
#include <vector>
#include <locale>
#include <codecvt>

#pragma comment(lib, "windowsapp")
#pragma comment(lib, "shcore")

using namespace winrt;
using namespace Windows::Foundation;
using namespace Windows::Graphics::Imaging;
using namespace Windows::Media::Ocr;
using namespace Windows::Storage::Streams;

// UTF-16 转 UTF-8
static std::string wstring_to_utf8(const std::wstring& wstr) {
    if (wstr.empty()) return std::string();
    int size_needed = WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), (int)wstr.size(), NULL, 0, NULL, NULL);
    std::string result(size_needed, 0);
    WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), (int)wstr.size(), &result[0], size_needed, NULL, NULL);
    return result;
}

// UTF-8 转 UTF-16
static std::wstring utf8_to_wstring(const std::string& str) {
    if (str.empty()) return std::wstring();
    int size_needed = MultiByteToWideChar(CP_UTF8, 0, str.c_str(), (int)str.size(), NULL, 0);
    std::wstring result(size_needed, 0);
    MultiByteToWideChar(CP_UTF8, 0, str.c_str(), (int)str.size(), &result[0], size_needed);
    return result;
}

// 从内存数据创建 SoftwareBitmap
static IAsyncOperation<SoftwareBitmap> CreateBitmapFromMemory(const unsigned char* data, int length) {
    // 创建内存流
    InMemoryRandomAccessStream stream;
    DataWriter writer(stream);
    writer.WriteBytes(array_view<const uint8_t>(data, data + length));
    co_await writer.StoreAsync();
    co_await writer.FlushAsync();
    writer.DetachStream();
    stream.Seek(0);

    // 解码图片
    BitmapDecoder decoder = co_await BitmapDecoder::CreateAsync(stream);

    // 获取 SoftwareBitmap，转换为 Gray8 或 BGRA8 格式（OCR 支持的格式）
    SoftwareBitmap bitmap = co_await decoder.GetSoftwareBitmapAsync(
        BitmapPixelFormat::Bgra8,
        BitmapAlphaMode::Premultiplied
    );

    co_return bitmap;
}

// 执行 OCR 识别
static IAsyncOperation<OCRResult> RecognizeAsync(const unsigned char* data, int length, const char** languages, int lang_count) {
    OCRResult result = {0};

    try {
        // 创建 SoftwareBitmap
        SoftwareBitmap bitmap = co_await CreateBitmapFromMemory(data, length);
        if (!bitmap) {
            result.error = _strdup("failed to create bitmap from image data");
            co_return result;
        }

        // 创建 OcrEngine
        OcrEngine engine{nullptr};

        if (languages != nullptr && lang_count > 0) {
            // 使用指定语言
            std::wstring langTag = utf8_to_wstring(languages[0]);
            Windows::Globalization::Language lang(langTag);
            engine = OcrEngine::TryCreateFromLanguage(lang);
        }

        if (!engine) {
            // 回退到用户配置语言
            engine = OcrEngine::TryCreateFromUserProfileLanguages();
        }

        if (!engine) {
            result.error = _strdup("failed to create OCR engine, no supported language available");
            co_return result;
        }

        // 执行 OCR
        OcrResult ocrResult = co_await engine.RecognizeAsync(bitmap);

        // 收集所有单词
        std::vector<OCRTextBlock> blocks;

        // 获取图片尺寸用于归一化坐标
        double imageWidth = static_cast<double>(bitmap.PixelWidth());
        double imageHeight = static_cast<double>(bitmap.PixelHeight());

        for (const auto& line : ocrResult.Lines()) {
            for (const auto& word : line.Words()) {
                OCRTextBlock block;

                // 获取文本
                std::string text = wstring_to_utf8(std::wstring(word.Text()));
                block.text = _strdup(text.c_str());

                // 获取边界框并归一化到 0-1 范围
                auto bbox = word.BoundingRect();
                block.x = bbox.X / imageWidth;
                block.y = bbox.Y / imageHeight;
                block.width = bbox.Width / imageWidth;
                block.height = bbox.Height / imageHeight;

                blocks.push_back(block);
            }
        }

        // 复制结果
        if (!blocks.empty()) {
            result.blocks = (OCRTextBlock*)malloc(sizeof(OCRTextBlock) * blocks.size());
            if (result.blocks) {
                memcpy(result.blocks, blocks.data(), sizeof(OCRTextBlock) * blocks.size());
                result.count = static_cast<int>(blocks.size());
            } else {
                // 内存分配失败，释放已分配的字符串
                for (auto& b : blocks) {
                    free(b.text);
                }
                result.error = _strdup("failed to allocate memory for results");
            }
        }

    } catch (const hresult_error& ex) {
        std::string msg = wstring_to_utf8(std::wstring(ex.message()));
        result.error = _strdup(msg.c_str());
    } catch (const std::exception& ex) {
        result.error = _strdup(ex.what());
    } catch (...) {
        result.error = _strdup("unknown error during OCR recognition");
    }

    co_return result;
}

extern "C" {

OCRResult sysocr_recognize(const unsigned char* data, int length, const char** languages, int lang_count) {
    OCRResult result = {0};

    try {
        // 初始化 WinRT
        winrt::init_apartment(winrt::apartment_type::multi_threaded);

        // 同步等待异步操作完成
        result = RecognizeAsync(data, length, languages, lang_count).get();

    } catch (const hresult_error& ex) {
        std::string msg = wstring_to_utf8(std::wstring(ex.message()));
        result.error = _strdup(msg.c_str());
    } catch (const std::exception& ex) {
        result.error = _strdup(ex.what());
    } catch (...) {
        result.error = _strdup("failed to initialize Windows Runtime");
    }

    return result;
}

void sysocr_free_result(OCRResult result) {
    if (result.blocks != NULL) {
        for (int i = 0; i < result.count; i++) {
            free(result.blocks[i].text);
        }
        free(result.blocks);
    }
    if (result.error != NULL) {
        free(result.error);
    }
}

} // extern "C"
