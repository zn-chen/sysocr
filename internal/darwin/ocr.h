#ifndef SYSOCR_DARWIN_H
#define SYSOCR_DARWIN_H

typedef struct {
    char* text;
    double x;
    double y;
    double width;
    double height;
} OCRTextBlock;

typedef struct {
    OCRTextBlock* blocks;
    int count;
    char* error;
} OCRResult;

OCRResult sysocr_recognize(const unsigned char* data, int length, const char** languages, int lang_count);
void sysocr_free_result(OCRResult result);

#endif
