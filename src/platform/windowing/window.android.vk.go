//go:build android

/******************************************************************************/
/* window.android.vk.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

func getInstanceExtensions() []string {
	return []string{"VK_KHR_surface\x00", "VK_KHR_android_surface\x00"}
}
