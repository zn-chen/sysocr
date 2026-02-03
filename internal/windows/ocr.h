#ifndef SYSOCR_WINDOWS_H
#define SYSOCR_WINDOWS_H

#ifdef __cplusplus
extern "C" {
#endif

// OCRTextBlock 表示单个文本块及其位置信息
typedef struct {
    char* text;
    double x;
    double y;
    double width;
    double height;
} OCRTextBlock;

// OCRResult 表示 OCR 识别结果
typedef struct {
    OCRTextBlock* blocks;
    int count;
    char* error;
} OCRResult;

// sysocr_recognize 执行 OCR 识别
// data: 图片数据
// length: 数据长度
// languages: 语言提示数组
// lang_count: 语言数量
OCRResult sysocr_recognize(const unsigned char* data, int length, const char** languages, int lang_count);

// sysocr_free_result 释放 OCRResult 占用的内存
void sysocr_free_result(OCRResult result);

#ifdef __cplusplus
}
#endif

#endif
