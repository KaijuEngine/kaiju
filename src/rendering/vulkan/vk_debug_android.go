// +build android

package vulkan

/*
#cgo android LDFLAGS: -llog

#include "vulkan/vulkan.h"
#include <android/log.h>

VkBool32 debugReportCallback(
  VkDebugReportFlagsEXT msgFlags,
  VkDebugReportObjectTypeEXT objType,
  uint64_t srcObject, size_t location,
  int32_t msgCode, const char * pLayerPrefix,
  const char * pMsg, void * pUserData )
{
  if (msgFlags & VK_DEBUG_REPORT_ERROR_BIT_EXT) {
    __android_log_print(ANDROID_LOG_ERROR,
                        "VkAndroid",
                        "ERROR: [%s] Code %i : %s",
                        pLayerPrefix, msgCode, pMsg);
  } else if (msgFlags & VK_DEBUG_REPORT_WARNING_BIT_EXT) {
    __android_log_print(ANDROID_LOG_WARN,
                        "VkAndroid",
                        "WARNING: [%s] Code %i : %s",
                        pLayerPrefix, msgCode, pMsg);
  } else if (msgFlags & VK_DEBUG_REPORT_PERFORMANCE_WARNING_BIT_EXT) {
    __android_log_print(ANDROID_LOG_WARN,
                        "VkAndroid",
                        "PERFORMANCE WARNING: [%s] Code %i : %s",
                        pLayerPrefix, msgCode, pMsg);
  } else if (msgFlags & VK_DEBUG_REPORT_INFORMATION_BIT_EXT) {
    __android_log_print(ANDROID_LOG_INFO,
                        "VkAndroid", "INFO: [%s] Code %i : %s",
                        pLayerPrefix, msgCode, pMsg);
  } else if (msgFlags & VK_DEBUG_REPORT_DEBUG_BIT_EXT) {
    __android_log_print(ANDROID_LOG_VERBOSE,
                        "VkAndroid", "DEBUG: [%s] Code %i : %s",
                        pLayerPrefix, msgCode, pMsg);
  }

  // Returning false tells the layer not to stop when the event occurs, so
  // they see the same behavior with and without validation layers enabled.
  return VK_FALSE;
}
*/
import "C"

var DebugReportCallbackAndroid = C.debugReportCallback
