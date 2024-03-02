//go:build android
// +build android

package vulkan

/*
#cgo android LDFLAGS: -Wl,--no-warn-mismatch
#cgo android CFLAGS: -DVK_USE_PLATFORM_ANDROID_KHR -D_NDK_MATH_NO_SOFTFP=1 -mfpu=vfp

#include <android/native_window.h>

#include "vulkan/vulkan.h"
#include "vk_wrapper.h"
#include "vk_bridge.h"
*/
import "C"
import "unsafe"

const (
	// UsePlatformAndroid as defined in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html
	UsePlatformAndroid = 1
	// KhrAndroidSurface as defined in vulkan/vulkan.h:3669
	KhrAndroidSurface = 1
	// KhrAndroidSurfaceSpecVersion as defined in vulkan/vulkan.h:3672
	KhrAndroidSurfaceSpecVersion = 6
	// KhrAndroidSurfaceExtensionName as defined in vulkan/vulkan.h:3673
	KhrAndroidSurfaceExtensionName = "VK_KHR_android_surface"
)

// CreateWindowSurface creates a Vulkan surface (VK_KHR_android_surface) for ANativeWindow from Android NDK.
func CreateWindowSurface(instance Instance, nativeWindow uintptr, pAllocator *AllocationCallbacks, pSurface *Surface) Result {
	cinstance, _ := *(*C.VkInstance)(unsafe.Pointer(&instance)), cgoAllocsUnknown
	cpAllocator, _ := (*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)), cgoAllocsUnknown
	cpSurface, _ := (*C.VkSurfaceKHR)(unsafe.Pointer(pSurface)), cgoAllocsUnknown
	pCreateInfo := &AndroidSurfaceCreateInfo{
		SType:  StructureTypeAndroidSurfaceCreateInfo,
		Window: (*ANativeWindow)(unsafe.Pointer(nativeWindow)),
	}
	cpCreateInfo, _ := pCreateInfo.PassRef()
	__ret := C.callVkCreateAndroidSurfaceKHR(cinstance, cpCreateInfo, cpAllocator, cpSurface)
	__v := (Result)(__ret)
	return __v
}

// GetRequiredInstanceExtensions should be used to query instance extensions required for surface initialization.
func GetRequiredInstanceExtensions() []string {
	return []string{
		"VK_KHR_surface\x00",
		"VK_KHR_android_surface\x00",
	}
}

// CreateAndroidSurface function as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#vkCreateAndroidSurfaceKHR
func CreateAndroidSurface(instance Instance, pCreateInfo *AndroidSurfaceCreateInfo, pAllocator *AllocationCallbacks, pSurface *Surface) Result {
	cinstance, _ := *(*C.VkInstance)(unsafe.Pointer(&instance)), cgoAllocsUnknown
	cpCreateInfo, _ := pCreateInfo.PassRef()
	cpAllocator, _ := (*C.VkAllocationCallbacks)(unsafe.Pointer(pAllocator)), cgoAllocsUnknown
	cpSurface, _ := (*C.VkSurfaceKHR)(unsafe.Pointer(pSurface)), cgoAllocsUnknown
	__ret := C.callVkCreateAndroidSurfaceKHR(cinstance, cpCreateInfo, cpAllocator, cpSurface)
	__v := (Result)(__ret)
	return __v
}

// allocAndroidSurfaceCreateInfoMemory allocates memory for type C.VkAndroidSurfaceCreateInfoKHR in C.
// The caller is responsible for freeing the this memory via C.free.
func allocAndroidSurfaceCreateInfoMemory(n int) unsafe.Pointer {
	mem, err := C.calloc(C.size_t(n), (C.size_t)(sizeOfAndroidSurfaceCreateInfoValue))
	if err != nil {
		panic("memory alloc error: " + err.Error())
	}
	return mem
}

const sizeOfAndroidSurfaceCreateInfoValue = unsafe.Sizeof([1]C.VkAndroidSurfaceCreateInfoKHR{})

// Ref returns the underlying reference to C object or nil if struct is nil.
func (x *AndroidSurfaceCreateInfo) Ref() *C.VkAndroidSurfaceCreateInfoKHR {
	if x == nil {
		return nil
	}
	return x.refeca5c35c
}

// Free invokes alloc map's free mechanism that cleanups any allocated memory using C free.
// Does nothing if struct is nil or has no allocation map.
func (x *AndroidSurfaceCreateInfo) Free() {
	if x != nil && x.allocseca5c35c != nil {
		x.allocseca5c35c.(*cgoAllocMap).Free()
		x.refeca5c35c = nil
	}
}

// NewAndroidSurfaceCreateInfoRef creates a new wrapper struct with underlying reference set to the original C object.
// Returns nil if the provided pointer to C object is nil too.
func NewAndroidSurfaceCreateInfoRef(ref unsafe.Pointer) *AndroidSurfaceCreateInfo {
	if ref == nil {
		return nil
	}
	obj := new(AndroidSurfaceCreateInfo)
	obj.refeca5c35c = (*C.VkAndroidSurfaceCreateInfoKHR)(unsafe.Pointer(ref))
	return obj
}

// PassRef returns the underlying C object, otherwise it will allocate one and set its values
// from this wrapping struct, counting allocations into an allocation map.
func (x *AndroidSurfaceCreateInfo) PassRef() (*C.VkAndroidSurfaceCreateInfoKHR, *cgoAllocMap) {
	if x == nil {
		return nil, nil
	} else if x.refeca5c35c != nil {
		return x.refeca5c35c, nil
	}
	memeca5c35c := allocAndroidSurfaceCreateInfoMemory(1)
	refeca5c35c := (*C.VkAndroidSurfaceCreateInfoKHR)(memeca5c35c)
	allocseca5c35c := new(cgoAllocMap)
	var csType_allocs *cgoAllocMap
	refeca5c35c.sType, csType_allocs = (C.VkStructureType)(x.SType), cgoAllocsUnknown
	allocseca5c35c.Borrow(csType_allocs)

	var cpNext_allocs *cgoAllocMap
	refeca5c35c.pNext, cpNext_allocs = *(*unsafe.Pointer)(unsafe.Pointer(&x.PNext)), cgoAllocsUnknown
	allocseca5c35c.Borrow(cpNext_allocs)

	var cflags_allocs *cgoAllocMap
	refeca5c35c.flags, cflags_allocs = (C.VkAndroidSurfaceCreateFlagsKHR)(x.Flags), cgoAllocsUnknown
	allocseca5c35c.Borrow(cflags_allocs)

	var cwindow_allocs *cgoAllocMap
	refeca5c35c.window, cwindow_allocs = *(**C.ANativeWindow)(unsafe.Pointer(&x.Window)), cgoAllocsUnknown
	allocseca5c35c.Borrow(cwindow_allocs)

	x.refeca5c35c = refeca5c35c
	x.allocseca5c35c = allocseca5c35c
	return refeca5c35c, allocseca5c35c

}

// PassValue does the same as PassRef except that it will try to dereference the returned pointer.
func (x AndroidSurfaceCreateInfo) PassValue() (C.VkAndroidSurfaceCreateInfoKHR, *cgoAllocMap) {
	if x.refeca5c35c != nil {
		return *x.refeca5c35c, nil
	}
	ref, allocs := x.PassRef()
	return *ref, allocs
}

// Deref uses the underlying reference to C object and fills the wrapping struct with values.
// Do not forget to call this method whether you get a struct for C object and want to read its values.
func (x *AndroidSurfaceCreateInfo) Deref() {
	if x.refeca5c35c == nil {
		return
	}
	x.SType = (StructureType)(x.refeca5c35c.sType)
	x.PNext = (unsafe.Pointer)(unsafe.Pointer(x.refeca5c35c.pNext))
	x.Flags = (AndroidSurfaceCreateFlags)(x.refeca5c35c.flags)
	x.Window = (*ANativeWindow)(unsafe.Pointer(x.refeca5c35c.window))
}

// ANativeWindow as declared in android/native_window.h:36
type ANativeWindow C.ANativeWindow

// AndroidSurfaceCreateFlags type as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#vkAndroidSurfaceCreateFlagsKHR
type AndroidSurfaceCreateFlags uint32

// AndroidSurfaceCreateInfo as declared in https://www.khronos.org/registry/vulkan/specs/1.0-wsi_extensions/xhtml/vkspec.html#vkAndroidSurfaceCreateInfoKHR
type AndroidSurfaceCreateInfo struct {
	SType          StructureType
	PNext          unsafe.Pointer
	Flags          AndroidSurfaceCreateFlags
	Window         *ANativeWindow
	refeca5c35c    *C.VkAndroidSurfaceCreateInfoKHR
	allocseca5c35c interface{}
}

// Ref returns a reference to C object as it is.
func (x *ANativeWindow) Ref() *C.ANativeWindow {
	if x == nil {
		return nil
	}
	return (*C.ANativeWindow)(unsafe.Pointer(x))
}

// Free cleanups the referenced memory using C free.
func (x *ANativeWindow) Free() {
	if x != nil {
		C.free(unsafe.Pointer(x))
	}
}

// NewANativeWindowRef converts the C object reference into a raw struct reference without wrapping.
func NewANativeWindowRef(ref unsafe.Pointer) *ANativeWindow {
	return (*ANativeWindow)(ref)
}

// NewANativeWindow allocates a new C object of this type and converts the reference into
// a raw struct reference without wrapping.
func NewANativeWindow() *ANativeWindow {
	return (*ANativeWindow)(allocANativeWindowMemory(1))
}

// allocANativeWindowMemory allocates memory for type C.ANativeWindow in C.
// The caller is responsible for freeing the this memory via C.free.
func allocANativeWindowMemory(n int) unsafe.Pointer {
	mem, err := C.calloc(C.size_t(n), (C.size_t)(sizeOfANativeWindowValue))
	if err != nil {
		panic("memory alloc error: " + err.Error())
	}
	return mem
}

const sizeOfANativeWindowValue = unsafe.Sizeof([1]C.ANativeWindow{})

// PassRef returns a reference to C object as it is or allocates a new C object of this type.
func (x *ANativeWindow) PassRef() *C.ANativeWindow {
	if x == nil {
		x = (*ANativeWindow)(allocANativeWindowMemory(1))
	}
	return (*C.ANativeWindow)(unsafe.Pointer(x))
}
