#include "vk_default_loader.h"
#include "vulkan/vulkan.h"

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
        // On macOS, try to load MoltenVK (which includes the Vulkan loader)
        void* libvulkan = dlopen("libMoltenVK.dylib", RTLD_NOW | RTLD_LOCAL);
        if (libvulkan == NULL) {
            libvulkan = dlopen("/usr/local/lib/libMoltenVK.dylib", RTLD_NOW | RTLD_LOCAL);
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
