#ifndef DIRECTORY_WIN32_H
#define DIRECTORY_WIN32_H

#include <stdio.h>
#include <stdbool.h>
#include <direct.h>
#include <shlobj.h>

const char* open_file_dialog(const char* startPath, const char* ext, void* hwnd) {
	bool valid = false;
	OPENFILENAMEA ofn = { 0 };       // common dialog box structure
	char szFile[260] = "";       // buffer for file name
	if (ext == NULL) {
		ext = "All Files\0*.*\0\0";
	}
	char filter[256] = "";
	snprintf(filter, sizeof(filter), "%s", ext);
	for (int i = 0; i < sizeof(filter) && filter[0] != 0; i++) {
		if (filter[i] == '\n') {
			filter[i] = '\0';
		}
	}
	// Initialize OPENFILENAME
	ZeroMemory(&ofn, sizeof(ofn));
	ofn.lStructSize = sizeof(ofn);
	ofn.hwndOwner = (HWND)hwnd;
	ofn.lpstrFile = (LPSTR)szFile;
	// Set lpstrFile[0] to '\0' so that GetOpenFileName does not
	// use the contents of szFile to initialize itself.
	ofn.lpstrFile[0] = '\0';
	ofn.nMaxFile = sizeof(szFile);
	ofn.lpstrFilter = filter;
	ofn.nFilterIndex = 1;
	ofn.Flags = OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT | OFN_NOREADONLYRETURN;
	valid = GetOpenFileNameA(&ofn) == TRUE;
	if (valid) {
		return ofn.lpstrFile;
	} else {
		return "";
	}
}

const char* save_file_dialog(const char* startPath, const char* fileName, const char* ext, void* hwnd) {
	bool valid = false;
	OPENFILENAMEA ofn = { 0 };       // common dialog box structure
	char szFile[260] = "";       // buffer for file name
	if (ext == NULL) {
		ext = "All Files\0*.*\0\0";
	}
	if (startPath == NULL) {
		startPath = "\0";
	}
	// Set lpstrFile[0] to '\0' so that GetOpenFileName does not
	// use the contents of szFile to initialize itself.
	if (fileName == NULL) {
		fileName = "\0";
	}
	snprintf(szFile, sizeof(szFile), "%s", fileName);
	char filter[256] = "";
	snprintf(filter, sizeof(filter), "%s", ext);
	for (int i = 0; i < sizeof(filter) && filter[0] != 0; i++) {
		if (filter[i] == '\n') {
			filter[i] = '\0';
		}
	}
	// Initialize OPENFILENAME
	ZeroMemory(&ofn, sizeof(ofn));
	ofn.lStructSize = sizeof(ofn);
	ofn.hwndOwner = (HWND)hwnd;
	ofn.lpstrFile = szFile;
	ofn.nMaxFile = sizeof(szFile);
	ofn.lpstrFilter = filter;
	ofn.nFilterIndex = 1;
	ofn.lpstrInitialDir = startPath;
	ofn.Flags = OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT | OFN_NOREADONLYRETURN;
	valid = GetSaveFileNameA(&ofn) == TRUE;
	if (valid) {
		return ofn.lpstrFile;
	} else {
		return "";
	}
}

#endif
