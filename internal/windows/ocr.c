#include "ocr.h"
#include <stdlib.h>
#include <string.h>

// TODO: implement Windows.Media.Ocr

OCRResult sysocr_recognize(const unsigned char* data, int length, const char** languages, int lang_count) {
    OCRResult result = {0};
    result.error = strdup("not implemented");
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
