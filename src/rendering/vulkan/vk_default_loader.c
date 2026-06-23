#include "vk_default_loader.h"
#include "kaiju_vulkan.h"
#include <stdlib.h> // getenv

#if defined(_WIN32)
    #include <windows.h>
    typedef void* (*vkGetInstanceProcAddrFn)(void* lib, const char* procname);
    static vkGetInstanceProcAddrFn symLoader;
    static void* loaderWrap(VkInstance instance, const char* vkproc) {
        return (*symLoader)(instance, vkproc);
    }
#elif defined(__android__) || defined(__ANDROID__) || defined(__linux__) || defined(__unix__) || defined(unix)
    #include <dlfcn.h>
    static void* (*symLoader)(void* lib, const char* procname);
    static void* loaderWrap(VkInstance instance, const char* vkproc) {
        return (*symLoader)(instance, vkproc);
    }
#elif defined(__APPLE__) && defined(__MACH__)
    #include <dlfcn.h>
    static void* (*symLoader)(void* lib, const char* procname);
    static void* loaderWrap(VkInstance instance, const char* vkproc) {
        return (*symLoader)(instance, vkproc);
    }
#endif

void* getDefaultProcAddr() {
    #if defined(_WIN32)
        HMODULE libvulkan = LoadLibrary(TEXT("vulkan-1.dll"));
        if (libvulkan == NULL) {
            return NULL;
        }
        symLoader = (vkGetInstanceProcAddrFn)GetProcAddress(libvulkan, "vkGetInstanceProcAddr");
        if (symLoader == NULL) {
            return NULL;
        }
        return &loaderWrap;
    #elif defined(__APPLE__) && defined(__MACH__)
        // Default: load MoltenVK directly (unchanged release behavior). Opt-in via
        // KAIJU_VULKAN_USE_LOADER: prefer the real Vulkan loader (libvulkan.dylib),
        // which loads MoltenVK as its ICD AND inserts explicit layers such as
        // VK_LAYER_KHRONOS_validation (configured via VK_ADD_LAYER_PATH /
        // VK_ICD_FILENAMES). dlopening MoltenVK directly bypasses the loader, so
        // layers can never be enumerated; the loader is therefore required to run
        // validation. See docs/engine/vulkan_validation_layers.md.
        void* libvulkan = NULL;
        if (getenv("KAIJU_VULKAN_USE_LOADER")) {
            libvulkan = dlopen("libvulkan.dylib", RTLD_NOW | RTLD_LOCAL);
            if (libvulkan == NULL) {
                libvulkan = dlopen("libvulkan.1.dylib", RTLD_NOW | RTLD_LOCAL);
            }
        }
        if (libvulkan == NULL) {
            libvulkan = dlopen("libMoltenVK.dylib", RTLD_NOW | RTLD_LOCAL);
        }
        if (libvulkan == NULL) {
            return NULL;
        }
        symLoader = dlsym(libvulkan, "vkGetInstanceProcAddr");
        if (symLoader == NULL) {
            return NULL;
        }
        return &loaderWrap;
    #elif defined(__android__) || defined(__ANDROID__) || defined(__linux__) || defined(__unix__) || defined(unix)
        void* libvulkan = dlopen("libvulkan.so", RTLD_NOW | RTLD_LOCAL);
        if (libvulkan == NULL) {
            return NULL;
        }
        symLoader = dlsym(libvulkan, "vkGetInstanceProcAddr");
        if (symLoader == NULL) {
            return NULL;
        }
        return &loaderWrap;
    #else
        // Unknown operating system
        return NULL;
    #endif
}
