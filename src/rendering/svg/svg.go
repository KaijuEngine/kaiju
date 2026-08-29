package svg

/*
#cgo CFLAGS: -I. -DPLUTOVG_BUILD_STATIC -DPLUTOSVG_BUILD_STATIC
#cgo !windows LDFLAGS: -lm
#include <stdlib.h>
#include <stdint.h>
#include "plutosvg.h"
#include "plutovg.h"

// Renders an SVG from a data buffer into a malloc'd straight-alpha RGBA buffer.
// Passing width/height set the target surface size (and container used to
// resolve the document's intrinsic size); pass 0 for either to fall back to
// the SVG's intrinsic dimensions.
// On success returns the buffer and sets *outW/*outH; caller must free() it.
// On failure returns NULL.
unsigned char* plutosvg_render_to_rgba(const char* data, int length, int width, int height, int* outW, int* outH)
{
    float cw = width > 0 ? (float)width : -1;
    float ch = height > 0 ? (float)height : -1;
    int rw = width > 0 ? width : -1;
    int rh = height > 0 ? height : -1;
    plutosvg_document_t* doc = plutosvg_document_load_from_data(data, length, cw, ch, NULL, NULL);
    if(doc == NULL)
        return NULL;
    plutovg_surface_t* surface = plutosvg_document_render_to_surface(doc, NULL, rw, rh, NULL, NULL, NULL);
    plutosvg_document_destroy(doc);
    if(surface == NULL)
        return NULL;
    int w = plutovg_surface_get_width(surface);
    int h = plutovg_surface_get_height(surface);
    int stride = plutovg_surface_get_stride(surface);
    const unsigned char* src = plutovg_surface_get_data(surface);
    unsigned char* out = (unsigned char*)malloc((size_t)w * h * 4);
    if(out != NULL) {
        for(int y = 0; y < h; ++y) {
            const uint32_t* src_row = (const uint32_t*)(src + stride * y);
            unsigned char* dst_row = out + (size_t)w * 4 * y;
            for(int x = 0; x < w; ++x) {
                uint32_t p = src_row[x];
                uint32_t a = (p >> 24) & 0xFF;
                if(a == 0) {
                    dst_row[x*4+0]=0; dst_row[x*4+1]=0; dst_row[x*4+2]=0; dst_row[x*4+3]=0;
                } else {
                    uint32_t r = (p >> 16) & 0xFF;
                    uint32_t g = (p >> 8) & 0xFF;
                    uint32_t b = p & 0xFF;
                    dst_row[x*4+0] = (unsigned char)((r*255)/a);
                    dst_row[x*4+1] = (unsigned char)((g*255)/a);
                    dst_row[x*4+2] = (unsigned char)((b*255)/a);
                    dst_row[x*4+3] = (unsigned char)a;
                }
            }
        }
    }
    plutovg_surface_destroy(surface);
    *outW = w;
    *outH = h;
    return out;
}
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

type Svg struct {
	Width  int
	Height int
	RGBA   []byte
}

// render decodes and rasterizes raw SVG bytes into straight-alpha RGBA.
// If width or height is <= 0, the SVG's intrinsic dimensions are used.
func render(data []byte, width, height int) (Svg, error) {
	var result Svg
	if len(data) == 0 {
		return result, fmt.Errorf("plutosvg: empty svg data")
	}
	var w, h C.int
	ptr := C.plutosvg_render_to_rgba((*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)), C.int(width), C.int(height), &w, &h)
	if ptr == nil {
		return result, fmt.Errorf("plutosvg: failed to load or render svg")
	}
	defer C.free(unsafe.Pointer(ptr))
	result.Width = int(w)
	result.Height = int(h)
	result.RGBA = C.GoBytes(unsafe.Pointer(ptr), C.int(int(w)*int(h)*4))
	return result, nil
}

// RenderString decodes and rasterizes an SVG provided as a string,
// scaled to the given width/height (<=0 uses the SVG's intrinsic size).
func RenderString(svg string, width, height int) (Svg, error) { return render([]byte(svg), width, height) }

// RenderFile decodes and rasterizes the SVG stored at the given path,
// scaled to the given width/height (<=0 uses the SVG's intrinsic size).
func RenderFile(path string, width, height int) (Svg, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Svg{}, err
	}
	return render(data, width, height)
}
