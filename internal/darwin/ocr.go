//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework Vision -framework CoreGraphics

#include "ocr.h"
*/
import "C"

// TODO: implement CGO bridge
