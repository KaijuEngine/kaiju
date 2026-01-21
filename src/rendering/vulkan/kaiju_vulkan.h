#ifndef KAIJU_VULKAN_H
#define KAIJU_VULKAN_H

/*
 * Wrapper header to force use of the vendored Vulkan headers.
 * This avoids SDK header shadowing when CGO_CFLAGS adds -I paths.
 */
#include "vulkan/vulkan.h"

#endif /* KAIJU_VULKAN_H */
