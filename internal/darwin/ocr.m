#import <Foundation/Foundation.h>
#import <Vision/Vision.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ImageIO/ImageIO.h>
#include "ocr.h"
#include <stdlib.h>
#include <string.h>

OCRResult sysocr_recognize(const unsigned char* data, int length, const char** languages, int lang_count) {
    OCRResult result = {0};

    @autoreleasepool {
        // 从原始字节创建 NSData
        NSData *imageData = [NSData dataWithBytes:data length:length];
        if (imageData == nil) {
            result.error = strdup("failed to create NSData from image bytes");
            return result;
        }

        // 创建 CGImageSource
        CGImageSourceRef imageSource = CGImageSourceCreateWithData((__bridge CFDataRef)imageData, NULL);
        if (imageSource == NULL) {
            result.error = strdup("failed to create image source");
            return result;
        }

        // 创建 CGImage
        CGImageRef cgImage = CGImageSourceCreateImageAtIndex(imageSource, 0, NULL);
        CFRelease(imageSource);
        if (cgImage == NULL) {
            result.error = strdup("failed to create CGImage");
            return result;
        }

        // 创建文字识别请求
        __block NSMutableArray<VNRecognizedTextObservation *> *observations = [NSMutableArray array];
        __block NSError *recognitionError = nil;

        VNRecognizeTextRequest *request = [[VNRecognizeTextRequest alloc] initWithCompletionHandler:^(VNRequest *req, NSError *error) {
            if (error != nil) {
                recognitionError = error;
                return;
            }
            for (VNRecognizedTextObservation *obs in req.results) {
                [observations addObject:obs];
            }
        }];

        // 配置请求参数
        request.recognitionLevel = VNRequestTextRecognitionLevelAccurate;
        request.usesLanguageCorrection = YES;

        // 设置识别语言（如果提供）
        if (languages != NULL && lang_count > 0) {
            NSMutableArray<NSString *> *langArray = [NSMutableArray arrayWithCapacity:lang_count];
            for (int i = 0; i < lang_count; i++) {
                [langArray addObject:[NSString stringWithUTF8String:languages[i]]];
            }
            request.recognitionLanguages = langArray;
        }

        // 创建请求处理器并执行
        VNImageRequestHandler *handler = [[VNImageRequestHandler alloc] initWithCGImage:cgImage options:@{}];
        NSError *performError = nil;
        BOOL success = [handler performRequests:@[request] error:&performError];

        CGImageRelease(cgImage);

        if (!success || performError != nil) {
            NSError *err = performError ?: recognitionError;
            const char *errMsg = err ? [[err localizedDescription] UTF8String] : "unknown error";
            result.error = strdup(errMsg);
            return result;
        }

        if (recognitionError != nil) {
            result.error = strdup([[recognitionError localizedDescription] UTF8String]);
            return result;
        }

        // 将 observations 转换为 OCRTextBlock 数组
        NSUInteger count = [observations count];
        if (count == 0) {
            result.blocks = NULL;
            result.count = 0;
            return result;
        }

        result.blocks = (OCRTextBlock *)malloc(sizeof(OCRTextBlock) * count);
        if (result.blocks == NULL) {
            result.error = strdup("failed to allocate memory");
            return result;
        }
        result.count = (int)count;

        for (NSUInteger i = 0; i < count; i++) {
            VNRecognizedTextObservation *obs = observations[i];
            VNRecognizedText *topCandidate = [[obs topCandidates:1] firstObject];

            // 获取文本
            NSString *text = topCandidate ? topCandidate.string : @"";
            result.blocks[i].text = strdup([text UTF8String]);

            // 获取边界框（Vision 使用归一化坐标，原点在左下角）
            CGRect bbox = obs.boundingBox;
            result.blocks[i].x = bbox.origin.x;
            // 从左下角原点转换为左上角原点
            result.blocks[i].y = 1.0 - bbox.origin.y - bbox.size.height;
            result.blocks[i].width = bbox.size.width;
            result.blocks[i].height = bbox.size.height;
        }
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
