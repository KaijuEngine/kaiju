#ifndef DIRECTORY_WIN32_H
#define DIRECTORY_WIN32_H

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <wchar.h>
#include <windows.h>
#include <shlobj.h>
#include <shobjidl.h>

/* Dialog mode requested by Go */
typedef enum dialog_mode_t {
	DIALOG_MODE_OPEN_FILE   = 0,
	DIALOG_MODE_OPEN_FILES  = 1,
	DIALOG_MODE_SAVE_FILE   = 2,
	DIALOG_MODE_OPEN_FOLDER = 3,
} dialog_mode_t;

/* Result status returned to Go */
typedef enum dialog_status_t {
	DIALOG_STATUS_ERROR  = -1,
	DIALOG_STATUS_CANCEL =  0,
	DIALOG_STATUS_OK     =  1,
} dialog_status_t;

/*
 * Input payload for run_native_file_dialog.
 * All text is utf8. Optional fields may be NULL.
 */
typedef struct dialog_request_t {
	int mode;
	const char *title_utf8;
	const char *current_directory_utf8;
	const char *filename_utf8;
	const char *root_utf8;
	const char *filters_utf8; /* "Name|*.ext;*.alt\\nOther|*.*" */
	const char *options_utf8; /* "Checkbox|1|\\nCombo|0|A;B;C" */
	int show_hidden;
	void *hwnd;
} dialog_request_t;

/*
 * Output payload for run_native_file_dialog.
 * Caller must release heap fields with free_native_file_dialog_result.
 */
typedef struct dialog_result_t {
	int status;
	int selected_filter_index;
	int hresult;
	int path_count;
	char **paths_utf8;
	char *selected_options_utf8;
	char *error_utf8;
} dialog_result_t;

/* Internal representation for dialog custom controls (parsed from options payload) */
typedef struct dialog_option_control_t {
	DWORD group_id;
	DWORD control_id;
	int is_checkbox;
	int default_value;
	wchar_t *name;
	wchar_t **items;
	int item_count;
} dialog_option_control_t;

/*
 * lightweight COM event sink used while dialog is shown.
 * Used to deny out-of-root folder navigation in dialog_event_on_folder_changing
 * so the UI does not allow browsing outside the configured root.
 */
typedef struct dialog_event_handler_t {
	IFileDialogEvents iface;
	LONG ref_count;
	const wchar_t *root;
} dialog_event_handler_t;

/* convert API-facing utf8 strings to win32 wide strings with owned storage */
static wchar_t *utf8_to_wide_alloc(const char *src) {
	if (src == NULL || src[0] == '\0') {
		return NULL;
	}
	int chars = MultiByteToWideChar(CP_UTF8, 0, src, -1, NULL, 0);
	if (chars <= 0) {
		return NULL;
	}
	wchar_t *dst = (wchar_t *)malloc((size_t)chars * sizeof(wchar_t));
	if (dst == NULL) {
		return NULL;
	}
	if (MultiByteToWideChar(CP_UTF8, 0, src, -1, dst, chars) <= 0) {
		free(dst);
		return NULL;
	}
	return dst;
}

/* convert Win32 wchar back to UTF-8 for Go-facing result payloads */
static char *wide_to_utf8_alloc(const wchar_t *src) {
	if (src == NULL || src[0] == L'\0') {
		return NULL;
	}
	int bytes = WideCharToMultiByte(CP_UTF8, 0, src, -1, NULL, 0, NULL, NULL);
	if (bytes <= 0) {
		return NULL;
	}
	char *dst = (char *)malloc((size_t)bytes);
	if (dst == NULL) {
		return NULL;
	}
	if (WideCharToMultiByte(CP_UTF8, 0, src, -1, dst, bytes, NULL, NULL) <= 0) {
		free(dst);
		return NULL;
	}
	return dst;
}

static char *cstr_dup_alloc(const char *src) {
	if (src == NULL) {
		return NULL;
	}
	size_t len = strlen(src);
	char *dst = (char *)malloc(len + 1);
	if (dst == NULL) {
		return NULL;
	}
	memcpy(dst, src, len + 1);
	return dst;
}

/* in-place tokenizer: replaces delimiter with '\0' and advances cursor */
static char *split_next_token_inplace(char **cursor, char delim) {
	if (cursor == NULL || *cursor == NULL) {
		return NULL;
	}
	char *start = *cursor;
	if (start[0] == '\0') {
		*cursor = NULL;
		return NULL;
	}
	char *p = start;
	while (*p != '\0' && *p != delim) {
		p++;
	}
	if (*p == delim) {
		*p = '\0';
		*cursor = p + 1;
	} else {
		*cursor = NULL;
	}
	return start;
}

static void set_error_message(dialog_result_t *res, const char *msg) {
	if (res == NULL || msg == NULL) {
		return;
	}
	if (res->error_utf8 != NULL) {
		free(res->error_utf8);
	}
	res->error_utf8 = cstr_dup_alloc(msg);
}

static bool push_path(dialog_result_t *res, const wchar_t *path) {
	if (res == NULL || path == NULL) {
		return false;
	}
	char *utf8 = wide_to_utf8_alloc(path);
	if (utf8 == NULL) {
		return false;
	}
	int new_count = res->path_count + 1;
	char **new_paths = (char **)realloc(res->paths_utf8, (size_t)new_count * sizeof(char *));
	if (new_paths == NULL) {
		free(utf8);
		return false;
	}
	res->paths_utf8 = new_paths;
	res->paths_utf8[res->path_count] = utf8;
	res->path_count = new_count;
	return true;
}

static bool append_text(char **buf, size_t *len, size_t *cap, const char *txt) {
	if (txt == NULL) {
		return true;
	}
	size_t n = strlen(txt);
	if (*len + n + 1 > *cap) {
		size_t next = (*cap == 0) ? 256 : *cap;
		while (*len + n + 1 > next) {
			next *= 2;
		}
		char *grown = (char *)realloc(*buf, next);
		if (grown == NULL) {
			return false;
		}
		*buf = grown;
		*cap = next;
	}
	memcpy((*buf) + *len, txt, n);
	*len += n;
	(*buf)[*len] = '\0';
	return true;
}

static bool append_char(char **buf, size_t *len, size_t *cap, char c) {
	char tmp[2] = {c, '\0'};
	return append_text(buf, len, cap, tmp);
}

static void clear_result_paths(dialog_result_t *res) {
	if (res == NULL || res->paths_utf8 == NULL) {
		if (res != NULL) {
			res->path_count = 0;
		}
		return;
	}
	for (int i = 0; i < res->path_count; i++) {
		if (res->paths_utf8[i] != NULL) {
			free(res->paths_utf8[i]);
		}
	}
	free(res->paths_utf8);
	res->paths_utf8 = NULL;
	res->path_count = 0;
}

static bool is_drive_root_path(const wchar_t *path) {
	return path != NULL && wcslen(path) == 3 && path[1] == L':' &&
		(path[2] == L'\\' || path[2] == L'/');
}

/* Normalize a path for root-prefix checks. */
static wchar_t *normalize_full_path_alloc(const wchar_t *path) {
	if (path == NULL || path[0] == L'\0') {
		return NULL;
	}
	DWORD need = GetFullPathNameW(path, 0, NULL, NULL);
	if (need == 0) {
		return NULL;
	}
	wchar_t *full = (wchar_t *)malloc((size_t)(need + 2) * sizeof(wchar_t));
	if (full == NULL) {
		return NULL;
	}
	DWORD got = GetFullPathNameW(path, need + 1, full, NULL);
	if (got == 0 || got > need) {
		free(full);
		return NULL;
	}
	while (got > 0 && (full[got - 1] == L'\\' || full[got - 1] == L'/')) {
		/* keep drive roots like "C:\" intact while normalizing path suffixes */
		if (got == 3 && full[1] == L':') {
			break;
		}
		full[got - 1] = L'\0';
		got--;
	}
	return full;
}

/* root is enforced after selection to avoid returning paths outside the requested scope. */
static bool path_is_within_root(const wchar_t *root, const wchar_t *path) {
	if (root == NULL || root[0] == L'\0') {
		return true;
	}
	wchar_t *full_root = normalize_full_path_alloc(root);
	wchar_t *full_path = normalize_full_path_alloc(path);
	if (full_root == NULL || full_path == NULL) {
		if (full_root != NULL) {
			free(full_root);
		}
		if (full_path != NULL) {
			free(full_path);
		}
		return false;
	}
	size_t root_len = wcslen(full_root);
	bool ok = false;
	if (_wcsnicmp(full_root, full_path, root_len) == 0) {
		/* Drive roots include the separator (e.g. "C:\") so any child path is valid after prefix match. */
		if (is_drive_root_path(full_root)) {
			ok = true;
		}
		/* enforce a full path segment boundary so "C:\foo" does not match "C:\foobar" */
		else if (full_path[root_len] == L'\0' || full_path[root_len] == L'\\' || full_path[root_len] == L'/') {
			ok = true;
		}
	}
	free(full_root);
	free(full_path);
	return ok;
}

static void free_filter_arrays(COMDLG_FILTERSPEC *specs, wchar_t **names, wchar_t **patterns, int count) {
	if (names != NULL) {
		for (int i = 0; i < count; i++) {
			if (names[i] != NULL) {
				free(names[i]);
			}
		}
		free(names);
	}
	if (patterns != NULL) {
		for (int i = 0; i < count; i++) {
			if (patterns[i] != NULL) {
				free(patterns[i]);
			}
		}
		free(patterns);
	}
	if (specs != NULL) {
		free(specs);
	}
}

/* expected format: "Label|*.ext\\nLabel2|*.foo;*.bar". falls back to All Files */
static bool build_filter_specs(const char *filters_utf8, COMDLG_FILTERSPEC **out_specs, int *out_count, wchar_t ***out_names, wchar_t ***out_patterns) {
	*out_specs = NULL;
	*out_count = 0;
	*out_names = NULL;
	*out_patterns = NULL;

	if (filters_utf8 == NULL || filters_utf8[0] == '\0') {
		COMDLG_FILTERSPEC *spec = (COMDLG_FILTERSPEC *)malloc(sizeof(COMDLG_FILTERSPEC));
		wchar_t **names = (wchar_t **)malloc(sizeof(wchar_t *));
		wchar_t **patterns = (wchar_t **)malloc(sizeof(wchar_t *));
		if (spec == NULL || names == NULL || patterns == NULL) {
			if (spec != NULL) free(spec);
			if (names != NULL) free(names);
			if (patterns != NULL) free(patterns);
			return false;
		}
		names[0] = utf8_to_wide_alloc("All Files (*.*)");
		patterns[0] = utf8_to_wide_alloc("*.*");
		if (names[0] == NULL || patterns[0] == NULL) {
			free_filter_arrays(spec, names, patterns, 1);
			return false;
		}
		spec[0].pszName = names[0];
		spec[0].pszSpec = patterns[0];
		*out_specs = spec;
		*out_count = 1;
		*out_names = names;
		*out_patterns = patterns;
		return true;
	}

	char *work = cstr_dup_alloc(filters_utf8);
	if (work == NULL) {
		return false;
	}

	int cap = 8;
	int count = 0;
	COMDLG_FILTERSPEC *specs = (COMDLG_FILTERSPEC *)malloc((size_t)cap * sizeof(COMDLG_FILTERSPEC));
	wchar_t **names = (wchar_t **)malloc((size_t)cap * sizeof(wchar_t *));
	wchar_t **patterns = (wchar_t **)malloc((size_t)cap * sizeof(wchar_t *));
	if (specs == NULL || names == NULL || patterns == NULL) {
		free(work);
		if (specs != NULL) free(specs);
		if (names != NULL) free(names);
		if (patterns != NULL) free(patterns);
		return false;
	}

	char *line_cursor = work;
	for (char *line = split_next_token_inplace(&line_cursor, '\n'); line != NULL; line = split_next_token_inplace(&line_cursor, '\n')) {
		if (line[0] == '\0') {
			continue;
		}
		char *sep = strchr(line, '|');
		if (sep == NULL) {
			continue;
		}
		*sep = '\0';
		char *name_utf8 = line;
		char *pattern_utf8 = sep + 1;
		if (name_utf8[0] == '\0' || pattern_utf8[0] == '\0') {
			continue;
		}

		if (count == cap) {
			int next_cap = cap * 2;
			/* grow all parallel arrays together to keep COMDLG_FILTERSPEC pointers aligned */
			COMDLG_FILTERSPEC *next_specs = (COMDLG_FILTERSPEC *)malloc((size_t)next_cap * sizeof(COMDLG_FILTERSPEC));
			wchar_t **next_names = (wchar_t **)malloc((size_t)next_cap * sizeof(wchar_t *));
			wchar_t **next_patterns = (wchar_t **)malloc((size_t)next_cap * sizeof(wchar_t *));
			if (next_specs == NULL || next_names == NULL || next_patterns == NULL) {
				if (next_specs != NULL) free(next_specs);
				if (next_names != NULL) free(next_names);
				if (next_patterns != NULL) free(next_patterns);
				free(work);
				free_filter_arrays(specs, names, patterns, count);
				return false;
			}

			memcpy(next_specs, specs, (size_t)count * sizeof(COMDLG_FILTERSPEC));
			memcpy(next_names, names, (size_t)count * sizeof(wchar_t *));
			memcpy(next_patterns, patterns, (size_t)count * sizeof(wchar_t *));
			free(specs);
			free(names);
			free(patterns);
			specs = next_specs;
			names = next_names;
			patterns = next_patterns;
			cap = next_cap;
		}

		names[count] = utf8_to_wide_alloc(name_utf8);
		patterns[count] = utf8_to_wide_alloc(pattern_utf8);
		if (names[count] == NULL || patterns[count] == NULL) {
			free(work);
			free_filter_arrays(specs, names, patterns, count + 1);
			return false;
		}
		specs[count].pszName = names[count];
		specs[count].pszSpec = patterns[count];
		count++;
	}
	free(work);

	if (count == 0) {
		free_filter_arrays(specs, names, patterns, 0);
		return build_filter_specs(NULL, out_specs, out_count, out_names, out_patterns);
	}

	*out_specs = specs;
	*out_count = count;
	*out_names = names;
	*out_patterns = patterns;
	return true;
}

static void free_option_controls(dialog_option_control_t *controls, int count) {
	if (controls == NULL) {
		return;
	}
	for (int i = 0; i < count; i++) {
		if (controls[i].name != NULL) {
			free(controls[i].name);
		}
		if (controls[i].items != NULL) {
			for (int j = 0; j < controls[i].item_count; j++) {
				if (controls[i].items[j] != NULL) {
					free(controls[i].items[j]);
				}
			}
			free(controls[i].items);
		}
	}
	free(controls);
}

static void free_wide_string_array(wchar_t **items, int count) {
	if (items == NULL) {
		return;
	}
	for (int i = 0; i < count; i++) {
		if (items[i] != NULL) {
			free(items[i]);
		}
	}
	free(items);
}

/* expected format per line: "name|default|v1;v2;v3". Empty values -> checkbox */
static bool build_option_controls(const char *options_utf8, dialog_option_control_t **out_controls, int *out_count) {
	*out_controls = NULL;
	*out_count = 0;
	if (options_utf8 == NULL || options_utf8[0] == '\0') {
		return true;
	}

	char *work = cstr_dup_alloc(options_utf8);
	if (work == NULL) {
		return false;
	}

	int cap = 8;
	int count = 0;
	dialog_option_control_t *controls = (dialog_option_control_t *)calloc((size_t)cap, sizeof(dialog_option_control_t));
	if (controls == NULL) {
		free(work);
		return false;
	}
	DWORD next_id = 100;

	char *line_cursor = work;
	for (char *line = split_next_token_inplace(&line_cursor, '\n'); line != NULL; line = split_next_token_inplace(&line_cursor, '\n')) {
		if (line[0] == '\0') {
			continue;
		}
		char *sep1 = strchr(line, '|');
		if (sep1 == NULL) {
			continue;
		}
		*sep1 = '\0';
		char *sep2 = strchr(sep1 + 1, '|');
		if (sep2 == NULL) {
			continue;
		}
		*sep2 = '\0';

		char *name_utf8 = line;
		char *default_utf8 = sep1 + 1;
		char *values_utf8 = sep2 + 1;
		if (name_utf8[0] == '\0') {
			continue;
		}

		if (count == cap) {
			int next_cap = cap * 2;
			dialog_option_control_t *next_controls = (dialog_option_control_t *)realloc(controls, (size_t)next_cap * sizeof(dialog_option_control_t));
			if (next_controls == NULL) {
				free(work);
				free_option_controls(controls, count);
				return false;
			}
			memset(next_controls + cap, 0, (size_t)(next_cap - cap) * sizeof(dialog_option_control_t));
			controls = next_controls;
			cap = next_cap;
		}

		dialog_option_control_t *ctl = &controls[count];
		ctl->group_id = next_id++;
		ctl->control_id = next_id++;
		ctl->name = utf8_to_wide_alloc(name_utf8);
		if (ctl->name == NULL) {
			free(work);
			free_option_controls(controls, count + 1);
			return false;
		}
		ctl->default_value = atoi(default_utf8);

		if (values_utf8[0] == '\0') {
			ctl->is_checkbox = 1;
			ctl->items = NULL;
			ctl->item_count = 0;
		} else {
			ctl->is_checkbox = 0;
			int item_cap = 4;
			int item_count = 0;
			wchar_t **items = (wchar_t **)malloc((size_t)item_cap * sizeof(wchar_t *));
			if (items == NULL) {
				free(work);
				free_option_controls(controls, count + 1);
				return false;
			}

			char *values_copy = cstr_dup_alloc(values_utf8);
			if (values_copy == NULL) {
				free(items);
				free(work);
				free_option_controls(controls, count + 1);
				return false;
			}

			char *value_cursor = values_copy;
			for (char *value = split_next_token_inplace(&value_cursor, ';'); value != NULL; value = split_next_token_inplace(&value_cursor, ';')) {
				if (value[0] == '\0') {
					continue;
				}
				if (item_count == item_cap) {
					int next_item_cap = item_cap * 2;
					wchar_t **next_items = (wchar_t **)realloc(items, (size_t)next_item_cap * sizeof(wchar_t *));
					if (next_items == NULL) {
						free(values_copy);
						free_wide_string_array(items, item_count);
						free(work);
						free_option_controls(controls, count + 1);
						return false;
					}
					items = next_items;
					item_cap = next_item_cap;
				}
				items[item_count] = utf8_to_wide_alloc(value);
				if (items[item_count] == NULL) {
					free(values_copy);
					free_wide_string_array(items, item_count);
					free(work);
					free_option_controls(controls, count + 1);
					return false;
				}
				item_count++;
			}
			free(values_copy);

			if (item_count == 0) {
				free(items);
				ctl->is_checkbox = 1;
				ctl->items = NULL;
				ctl->item_count = 0;
			} else {
				ctl->items = items;
				ctl->item_count = item_count;
			}
		}
		count++;
	}

	free(work);
	*out_controls = controls;
	*out_count = count;
	return true;
}

/*
 * IFileDialogCustomize methods return HRESULTs; keep the first failure visible to Go.
 * Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nn-shobjidl_core-ifiledialogcustomize
 */
static HRESULT add_option_controls(IFileDialogCustomize *pfdc, dialog_option_control_t *controls, int count) {
	if (pfdc == NULL || controls == NULL || count == 0) {
		return S_OK;
	}
	for (int i = 0; i < count; i++) {
		dialog_option_control_t *ctl = &controls[i];
		if (ctl->is_checkbox) {
			HRESULT hr = pfdc->lpVtbl->StartVisualGroup(pfdc, ctl->group_id, L"");
			if (FAILED(hr)) return hr;
			hr = pfdc->lpVtbl->AddCheckButton(pfdc, ctl->control_id, ctl->name, ctl->default_value ? TRUE : FALSE);
			if (FAILED(hr)) {
				pfdc->lpVtbl->EndVisualGroup(pfdc);
				return hr;
			}
			hr = pfdc->lpVtbl->SetControlState(pfdc, ctl->control_id, CDCS_VISIBLE | CDCS_ENABLED);
			if (FAILED(hr)) {
				pfdc->lpVtbl->EndVisualGroup(pfdc);
				return hr;
			}
			hr = pfdc->lpVtbl->EndVisualGroup(pfdc);
			if (FAILED(hr)) return hr;
		} else {
			HRESULT hr = pfdc->lpVtbl->StartVisualGroup(pfdc, ctl->group_id, ctl->name);
			if (FAILED(hr)) return hr;
			hr = pfdc->lpVtbl->AddComboBox(pfdc, ctl->control_id);
			if (FAILED(hr)) {
				pfdc->lpVtbl->EndVisualGroup(pfdc);
				return hr;
			}
			for (int j = 0; j < ctl->item_count; j++) {
				hr = pfdc->lpVtbl->AddControlItem(pfdc, ctl->control_id, (DWORD)j, ctl->items[j]);
				if (FAILED(hr)) {
					pfdc->lpVtbl->EndVisualGroup(pfdc);
					return hr;
				}
			}
			int idx = ctl->default_value;
			if (idx < 0) idx = 0;
			if (idx >= ctl->item_count) idx = ctl->item_count - 1;
			hr = pfdc->lpVtbl->SetSelectedControlItem(pfdc, ctl->control_id, (DWORD)idx);
			if (FAILED(hr)) {
				pfdc->lpVtbl->EndVisualGroup(pfdc);
				return hr;
			}
			hr = pfdc->lpVtbl->SetControlState(pfdc, ctl->control_id, CDCS_VISIBLE | CDCS_ENABLED);
			if (FAILED(hr)) {
				pfdc->lpVtbl->EndVisualGroup(pfdc);
				return hr;
			}
			hr = pfdc->lpVtbl->EndVisualGroup(pfdc);
			if (FAILED(hr)) return hr;
		}
	}
	return S_OK;
}

/* serializes selections as "name|b|0/1" or "name|i|index", one control per line */
static char *collect_selected_options_utf8(IFileDialogCustomize *pfdc, dialog_option_control_t *controls, int count) {
	if (pfdc == NULL || controls == NULL || count == 0) {
		return NULL;
	}
	/* values are queried after Show() returns to capture final control state */
	char *buf = NULL;
	size_t len = 0;
	size_t cap = 0;
	bool first = true;

	for (int i = 0; i < count; i++) {
		dialog_option_control_t *ctl = &controls[i];
		char *name_utf8 = wide_to_utf8_alloc(ctl->name);
		if (name_utf8 == NULL) {
			continue;
		}

		if (!first) {
			if (!append_char(&buf, &len, &cap, '\n')) {
				free(name_utf8);
				free(buf);
				return NULL;
			}
		}
		first = false;

		if (!append_text(&buf, &len, &cap, name_utf8) || !append_char(&buf, &len, &cap, '|')) {
			free(name_utf8);
			free(buf);
			return NULL;
		}
		free(name_utf8);

		if (ctl->is_checkbox) {
			BOOL checked = FALSE;
			if (FAILED(pfdc->lpVtbl->GetCheckButtonState(pfdc, ctl->control_id, &checked))) {
				checked = ctl->default_value ? TRUE : FALSE;
			}
			if (!append_text(&buf, &len, &cap, checked ? "b|1" : "b|0")) {
				free(buf);
				return NULL;
			}
		} else {
			DWORD selected = (DWORD)ctl->default_value;
			if (FAILED(pfdc->lpVtbl->GetSelectedControlItem(pfdc, ctl->control_id, &selected))) {
				selected = (DWORD)ctl->default_value;
			}
			char tmp[32];
			snprintf(tmp, sizeof(tmp), "i|%lu", (unsigned long)selected);
			if (!append_text(&buf, &len, &cap, tmp)) {
				free(buf);
				return NULL;
			}
		}
	}

	return buf;
}

static HRESULT STDMETHODCALLTYPE dialog_event_query_interface(IFileDialogEvents *This, REFIID riid, void **ppvObject) {
	if (ppvObject == NULL) {
		return E_POINTER;
	}
	*ppvObject = NULL;
	if (riid == NULL) {
		return E_NOINTERFACE;
	}
	if (IsEqualIID(riid, &IID_IUnknown) || IsEqualIID(riid, &IID_IFileDialogEvents)) {
		*ppvObject = This;
		This->lpVtbl->AddRef(This);
		return S_OK;
	}
	return E_NOINTERFACE;
}

static ULONG STDMETHODCALLTYPE dialog_event_add_ref(IFileDialogEvents *This) {
	dialog_event_handler_t *self = (dialog_event_handler_t *)This;
	return (ULONG)InterlockedIncrement(&self->ref_count);
}

static ULONG STDMETHODCALLTYPE dialog_event_release(IFileDialogEvents *This) {
	dialog_event_handler_t *self = (dialog_event_handler_t *)This;
	LONG ref = InterlockedDecrement(&self->ref_count);
	if (ref < 0) {
		ref = 0;
	}
	/* handler storage is stack-owned by run_native_file_dialog; do not free from Release */
	return (ULONG)ref;
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_file_ok(IFileDialogEvents *This, IFileDialog *pfd) {
	(void)This;
	(void)pfd;
	return S_OK;
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_folder_changing(IFileDialogEvents *This, IFileDialog *pfd, IShellItem *psiFolder) {
	(void)pfd;
	dialog_event_handler_t *self = (dialog_event_handler_t *)This;
	if (self == NULL || self->root == NULL || self->root[0] == L'\0') {
		return S_OK;
	}
	if (psiFolder == NULL) {
		return HRESULT_FROM_WIN32(ERROR_ACCESS_DENIED);
	}

	PWSTR folder_path = NULL;
	HRESULT hr = psiFolder->lpVtbl->GetDisplayName(psiFolder, SIGDN_FILESYSPATH, &folder_path);
	if (FAILED(hr) || folder_path == NULL) {
		if (folder_path != NULL) {
			CoTaskMemFree(folder_path);
		}
		return FAILED(hr) ? hr : HRESULT_FROM_WIN32(ERROR_ACCESS_DENIED);
	}

	bool ok = path_is_within_root(self->root, folder_path);
	CoTaskMemFree(folder_path);

	/*
	 * returning a failure HRESULT denies navigation, keeping the picker constrained
	 * to the configured root. Final selected paths are validated again after Show()
	 * as a hard safety check.
	 * Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialogevents-onfolderchanging
	 */
	return ok ? S_OK : HRESULT_FROM_WIN32(ERROR_ACCESS_DENIED);
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_folder_change(IFileDialogEvents *This, IFileDialog *pfd) {
	(void)This;
	(void)pfd;
	return S_OK;
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_selection_change(IFileDialogEvents *This, IFileDialog *pfd) {
	(void)This;
	(void)pfd;
	return S_OK;
}

/*
 * if the selected file is locked or in use, let the
 * Windows file dialog decide what message/action to show
 */
static HRESULT STDMETHODCALLTYPE dialog_event_on_share_violation(IFileDialogEvents *This, IFileDialog *pfd, IShellItem *psi, FDE_SHAREVIOLATION_RESPONSE *pResponse) {
	(void)This;
	(void)pfd;
	(void)psi;
	/* windows is supposed to provide a valid pointer -> check anyway */
	if (pResponse == NULL) {
		return E_POINTER;
	}
	/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialogevents-onshareviolation */
	/* Enum docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/ne-shobjidl_core-fde_shareviolation_response */
	*pResponse = FDESVR_DEFAULT;
	return S_OK;
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_type_change(IFileDialogEvents *This, IFileDialog *pfd) {
	(void)This;
	(void)pfd;
	return S_OK;
}

static HRESULT STDMETHODCALLTYPE dialog_event_on_overwrite(IFileDialogEvents *This, IFileDialog *pfd, IShellItem *psi, FDE_OVERWRITE_RESPONSE *pResponse) {
	(void)This;
	(void)pfd;
	(void)psi;
	/* Windows is supposed to provide a valid pointer -> check anyway */
	if (pResponse == NULL) {
		return E_POINTER;
	}
	/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialogevents-onoverwrite */
	/* Enum docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/ne-shobjidl_core-fde_overwrite_response */
	*pResponse = FDEOR_DEFAULT; /* Use the normal built-in behavior -> standard overwrite confirmation */
	return S_OK;
}

static IFileDialogEventsVtbl g_dialog_event_vtbl = {
	dialog_event_query_interface,
	dialog_event_add_ref,
	dialog_event_release,
	dialog_event_on_file_ok,
	dialog_event_on_folder_changing,
	dialog_event_on_folder_change,
	dialog_event_on_selection_change,
	dialog_event_on_share_violation,
	dialog_event_on_type_change,
	dialog_event_on_overwrite,
};

static void dialog_event_handler_init(dialog_event_handler_t *handler, const wchar_t *root) {
	if (handler == NULL) {
		return;
	}
	handler->iface.lpVtbl = &g_dialog_event_vtbl;
	handler->ref_count = 1;
	handler->root = root;
}

static wchar_t *resolve_dialog_start_folder_alloc(const wchar_t *current_dir, const wchar_t *root) {
	/*
	 * Prefer starting at current_dir only when it is within root.
	 * If a root is configured and current_dir falls outside it, start from root.
	 */
	if (root != NULL && root[0] != L'\0') {
		if (current_dir != NULL && current_dir[0] != L'\0' && path_is_within_root(root, current_dir)) {
			return normalize_full_path_alloc(current_dir);
		}
		return normalize_full_path_alloc(root);
	}
	if (current_dir != NULL && current_dir[0] != L'\0') {
		return normalize_full_path_alloc(current_dir);
	}
	return NULL;
}

static dialog_result_t make_empty_result(void) {
	dialog_result_t res;
	res.status = DIALOG_STATUS_CANCEL;
	res.selected_filter_index = 0;
	res.hresult = 0;
	res.path_count = 0;
	res.paths_utf8 = NULL;
	res.selected_options_utf8 = NULL;
	res.error_utf8 = NULL;
	return res;
}

static dialog_result_t run_native_file_dialog(const dialog_request_t *req) {
	/* high-level flow: init COM -> configure dialog -> show -> collect results -> cleanup */
	dialog_result_t res = make_empty_result();
	if (req == NULL) {
		res.status = DIALOG_STATUS_ERROR;
		set_error_message(&res, "dialog request was null");
		return res;
	}
	if (req->mode < DIALOG_MODE_OPEN_FILE || req->mode > DIALOG_MODE_OPEN_FOLDER) {
		res.status = DIALOG_STATUS_ERROR;
		set_error_message(&res, "invalid native file dialog mode");
		return res;
	}

	HRESULT hr = CoInitializeEx(NULL, COINIT_APARTMENTTHREADED | COINIT_DISABLE_OLE1DDE);
	bool should_uninitialize = SUCCEEDED(hr);
	if (hr == RPC_E_CHANGED_MODE) {
		/* RPC_E_CHANGED_MODE means the thread has an incompatible COM apartment */
		/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/combaseapi/nf-combaseapi-coinitializeex */
		res.status = DIALOG_STATUS_ERROR;
		res.hresult = (int)hr;
		set_error_message(&res, "COM is already initialized with an incompatible apartment model");
		return res;
	} else if (FAILED(hr)) {
		res.status = DIALOG_STATUS_ERROR;
		res.hresult = (int)hr;
		set_error_message(&res, "failed to initialize COM for file dialog");
		return res;
	}

	IFileDialog *pfd = NULL;
	if (req->mode == DIALOG_MODE_SAVE_FILE) {
		hr = CoCreateInstance(&CLSID_FileSaveDialog, NULL, CLSCTX_INPROC_SERVER, &IID_IFileSaveDialog, (void **)&pfd);
	} else {
		hr = CoCreateInstance(&CLSID_FileOpenDialog, NULL, CLSCTX_INPROC_SERVER, &IID_IFileOpenDialog, (void **)&pfd);
	}
	if (FAILED(hr) || pfd == NULL) {
		res.status = DIALOG_STATUS_ERROR;
		res.hresult = (int)hr;
		set_error_message(&res, "failed to create native file dialog");
		if (should_uninitialize) {
			CoUninitialize();
		}
		return res;
	}

	wchar_t *title = utf8_to_wide_alloc(req->title_utf8);
	wchar_t *current_dir = utf8_to_wide_alloc(req->current_directory_utf8);
	wchar_t *filename = utf8_to_wide_alloc(req->filename_utf8);
	wchar_t *root = utf8_to_wide_alloc(req->root_utf8);
	wchar_t *root_normalized = normalize_full_path_alloc(root);
	/* use root when possible to make prefix checks stable across equivalent paths */
	const wchar_t *enforced_root = root_normalized;
	if (enforced_root == NULL) {
		enforced_root = root;
	}

	COMDLG_FILTERSPEC *filter_specs = NULL;
	wchar_t **filter_names = NULL;
	wchar_t **filter_patterns = NULL;
	int filter_count = 0;
	if (!build_filter_specs(req->filters_utf8, &filter_specs, &filter_count, &filter_names, &filter_patterns)) {
		res.status = DIALOG_STATUS_ERROR;
		set_error_message(&res, "failed to build dialog file filters");
		pfd->lpVtbl->Release(pfd);
		if (title) free(title);
		if (current_dir) free(current_dir);
		if (filename) free(filename);
		if (root) free(root);
		if (root_normalized) free(root_normalized);
		if (should_uninitialize) CoUninitialize();
		return res;
	}

	dialog_option_control_t *controls = NULL;
	int control_count = 0;
	HWND owner = (HWND)req->hwnd;
	if (!build_option_controls(req->options_utf8, &controls, &control_count)) {
		res.status = DIALOG_STATUS_ERROR;
		set_error_message(&res, "failed to build dialog option controls");
		pfd->lpVtbl->Release(pfd);
		if (title) free(title);
		if (current_dir) free(current_dir);
		if (filename) free(filename);
		if (root) free(root);
		if (root_normalized) free(root_normalized);
		free_filter_arrays(filter_specs, filter_names, filter_patterns, filter_count);
		if (should_uninitialize) CoUninitialize();
		return res;
	}

	DWORD opts = 0;
	hr = pfd->lpVtbl->GetOptions(pfd, &opts);
	if (FAILED(hr)) {
		res.status = DIALOG_STATUS_ERROR;
		res.hresult = (int)hr;
		set_error_message(&res, "failed to read native dialog options");
		goto cleanup;
	}
	opts |= FOS_FORCEFILESYSTEM;
	if (req->mode == DIALOG_MODE_OPEN_FILES) opts |= FOS_ALLOWMULTISELECT;
	if (req->mode == DIALOG_MODE_OPEN_FOLDER) opts |= FOS_PICKFOLDERS;
	if (req->show_hidden) opts |= FOS_FORCESHOWHIDDEN;
	if (req->mode == DIALOG_MODE_OPEN_FILE || req->mode == DIALOG_MODE_OPEN_FILES) opts |= FOS_FILEMUSTEXIST;
	if (req->mode == DIALOG_MODE_SAVE_FILE) opts |= FOS_OVERWRITEPROMPT;
	if ((current_dir != NULL && current_dir[0] != L'\0') || (root != NULL && root[0] != L'\0')) {
		opts |= FOS_PATHMUSTEXIST;
	}
	hr = pfd->lpVtbl->SetOptions(pfd, opts);
	if (FAILED(hr)) {
		res.status = DIALOG_STATUS_ERROR;
		res.hresult = (int)hr;
		set_error_message(&res, "failed to apply native dialog options");
		goto cleanup;
	}

	if (title != NULL) {
		hr = pfd->lpVtbl->SetTitle(pfd, title);
		if (FAILED(hr)) {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to set native dialog title");
			goto cleanup;
		}
	}
	if (filename != NULL && req->mode == DIALOG_MODE_SAVE_FILE) {
		hr = pfd->lpVtbl->SetFileName(pfd, filename);
		if (FAILED(hr)) {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to set native dialog filename");
			goto cleanup;
		}
	}
	if (filter_count > 0 && req->mode != DIALOG_MODE_OPEN_FOLDER) {
		hr = pfd->lpVtbl->SetFileTypes(pfd, (UINT)filter_count, filter_specs);
		if (FAILED(hr)) {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to configure native dialog file filters");
			goto cleanup;
		}
		hr = pfd->lpVtbl->SetFileTypeIndex(pfd, 1);
		if (FAILED(hr)) {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to select default native dialog filter");
			goto cleanup;
		}
	}

	dialog_event_handler_t event_handler;
	DWORD event_cookie = 0;
	bool event_advised = false;
	if (enforced_root != NULL && enforced_root[0] != L'\0') {
		dialog_event_handler_init(&event_handler, enforced_root);
		/*
		 * Advise wires dialog_event_on_folder_changing so navigation can be
		 * blocked immediately.
		 * We still enforce root again against the final selected path(s) below.
		 * Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialog-advise
		 */
		hr = pfd->lpVtbl->Advise(pfd, (IFileDialogEvents *)&event_handler, &event_cookie);
		if (SUCCEEDED(hr)) {
			event_advised = true;
		} else {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to attach native dialog root guard");
			goto cleanup;
		}
	}

	IShellItem *folder_item = NULL;
	wchar_t *folder_ref = resolve_dialog_start_folder_alloc(current_dir, enforced_root);
	if (folder_ref != NULL) {
		/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-shcreateitemfromparsingname */
		hr = SHCreateItemFromParsingName(folder_ref, NULL, &IID_IShellItem, (void **)&folder_item);
		if (SUCCEEDED(hr) && folder_item != NULL) {
			/* SetDefaultFolder respects MRU (most recently used) */
			/* SetFolder forces the constrained starting folder */
			/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialog-setdefaultfolder */
			/*hr = pfd->lpVtbl->SetDefaultFolder(pfd, folder_item);
			if (FAILED(hr) && enforced_root != NULL && enforced_root[0] != L'\0') {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "failed to set native dialog default root folder");
				folder_item->lpVtbl->Release(folder_item);
				free(folder_ref);
				goto cleanup;
			}*/
			/* Docs: https://learn.microsoft.com/en-us/windows/win32/api/shobjidl_core/nf-shobjidl_core-ifiledialog-setfolder */
			hr = pfd->lpVtbl->SetFolder(pfd, folder_item);
			if (FAILED(hr) && enforced_root != NULL && enforced_root[0] != L'\0') {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "failed to set native dialog root folder");
				folder_item->lpVtbl->Release(folder_item);
				free(folder_ref);
				goto cleanup;
			}
			folder_item->lpVtbl->Release(folder_item);
		} else if (enforced_root != NULL && enforced_root[0] != L'\0') {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)hr;
			set_error_message(&res, "failed to create native dialog root folder item");
			free(folder_ref);
			goto cleanup;
		}
		free(folder_ref);
	}

	IFileDialogCustomize *pfdc = NULL;
	if (control_count > 0) {
		hr = pfd->lpVtbl->QueryInterface(pfd, &IID_IFileDialogCustomize, (void **)&pfdc);
		if (SUCCEEDED(hr) && pfdc != NULL) {
			hr = add_option_controls(pfdc, controls, control_count);
		}
		if (FAILED(hr) || pfdc == NULL) {
			res.status = DIALOG_STATUS_ERROR;
			res.hresult = (int)(FAILED(hr) ? hr : E_NOINTERFACE);
			set_error_message(&res, "failed to add custom controls to native dialog");
			if (pfdc) {
				pfdc->lpVtbl->Release(pfdc);
				pfdc = NULL;
			}
			goto cleanup;
		}
	}

	hr = pfd->lpVtbl->Show(pfd, owner);
	res.hresult = (int)hr;
	if (hr == HRESULT_FROM_WIN32(ERROR_CANCELLED)) {
		res.status = DIALOG_STATUS_CANCEL;
	} else if (FAILED(hr)) {
		res.status = DIALOG_STATUS_ERROR;
		set_error_message(&res, "native file dialog failed to show");
	} else {
		res.status = DIALOG_STATUS_OK;
		UINT idx = 1;
		if (SUCCEEDED(pfd->lpVtbl->GetFileTypeIndex(pfd, &idx))) {
			if (idx > 0) {
				idx -= 1;
			}
			res.selected_filter_index = (int)idx;
		}

		/* Custom controls are read after Show() returns because that is when
		 * the final user choice is available. Go fills defaults if this is empty.
		 */
		if (pfdc != NULL && control_count > 0) {
			res.selected_options_utf8 = collect_selected_options_utf8(pfdc, controls, control_count);
		}

		if (req->mode == DIALOG_MODE_OPEN_FILES) {
			IShellItemArray *results = NULL;
			hr = ((IFileOpenDialog *)pfd)->lpVtbl->GetResults((IFileOpenDialog *)pfd, &results);
			if (FAILED(hr) || results == NULL) {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "native dialog accepted but failed to read selected files");
				goto cleanup;
			}
			DWORD count = 0;
			hr = results->lpVtbl->GetCount(results, &count);
			if (FAILED(hr)) {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "failed to count selected files");
				clear_result_paths(&res);
				results->lpVtbl->Release(results);
				goto cleanup;
			}
			for (DWORD i = 0; i < count; i++) {
				IShellItem *item = NULL;
				hr = results->lpVtbl->GetItemAt(results, i, &item);
				if (FAILED(hr) || item == NULL) {
					res.status = DIALOG_STATUS_ERROR;
					res.hresult = (int)hr;
					set_error_message(&res, "failed to read selected file item");
					clear_result_paths(&res);
					results->lpVtbl->Release(results);
					goto cleanup;
				}

				PWSTR file_path = NULL;
				hr = item->lpVtbl->GetDisplayName(item, SIGDN_FILESYSPATH, &file_path);
				if (FAILED(hr) || file_path == NULL) {
					res.status = DIALOG_STATUS_ERROR;
					res.hresult = (int)hr;
					set_error_message(&res, "failed to read selected file path");
					clear_result_paths(&res);
					if (file_path != NULL) {
						CoTaskMemFree(file_path);
					}
					item->lpVtbl->Release(item);
					results->lpVtbl->Release(results);
					goto cleanup;
				}

				/* keep this post-selection check even with events; callers rely on hard enforcement */
				if (!path_is_within_root(enforced_root, file_path)) {
					res.status = DIALOG_STATUS_ERROR;
					set_error_message(&res, "selected path is outside configured dialog root");
					clear_result_paths(&res);
					CoTaskMemFree(file_path);
					item->lpVtbl->Release(item);
					results->lpVtbl->Release(results);
					goto cleanup;
				}
				if (!push_path(&res, file_path)) {
					res.status = DIALOG_STATUS_ERROR;
					set_error_message(&res, "failed to store selected file path");
					clear_result_paths(&res);
					CoTaskMemFree(file_path);
					item->lpVtbl->Release(item);
					results->lpVtbl->Release(results);
					goto cleanup;
				}

				CoTaskMemFree(file_path);
				item->lpVtbl->Release(item);
			}
			results->lpVtbl->Release(results);
			if (res.path_count == 0) {
				res.status = DIALOG_STATUS_ERROR;
				set_error_message(&res, "native dialog accepted but returned no file paths");
				goto cleanup;
			}
		} else {
			IShellItem *item = NULL;
			hr = pfd->lpVtbl->GetResult(pfd, &item);
			if (FAILED(hr) || item == NULL) {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "native dialog accepted but failed to read selected item");
				goto cleanup;
			}
			PWSTR file_path = NULL;
			hr = item->lpVtbl->GetDisplayName(item, SIGDN_FILESYSPATH, &file_path);
			if (FAILED(hr) || file_path == NULL) {
				res.status = DIALOG_STATUS_ERROR;
				res.hresult = (int)hr;
				set_error_message(&res, "failed to read selected file path");
				if (file_path != NULL) {
					CoTaskMemFree(file_path);
				}
				item->lpVtbl->Release(item);
				goto cleanup;
			}
			if (!path_is_within_root(enforced_root, file_path)) {
				res.status = DIALOG_STATUS_ERROR;
				set_error_message(&res, "selected path is outside configured dialog root");
				CoTaskMemFree(file_path);
				item->lpVtbl->Release(item);
				goto cleanup;
			}
			if (!push_path(&res, file_path)) {
				res.status = DIALOG_STATUS_ERROR;
				set_error_message(&res, "failed to store selected file path");
				CoTaskMemFree(file_path);
				item->lpVtbl->Release(item);
				goto cleanup;
			}
			CoTaskMemFree(file_path);
			item->lpVtbl->Release(item);
		}
	}

cleanup:
	if (owner != NULL) {
		SetForegroundWindow(owner);
	}
	if (event_advised) {
		pfd->lpVtbl->Unadvise(pfd, event_cookie);
	}
	if (pfdc != NULL) pfdc->lpVtbl->Release(pfdc);
	pfd->lpVtbl->Release(pfd);
	if (title) free(title);
	if (current_dir) free(current_dir);
	if (filename) free(filename);
	if (root) free(root);
	if (root_normalized) free(root_normalized);
	free_filter_arrays(filter_specs, filter_names, filter_patterns, filter_count);
	free_option_controls(controls, control_count);
	if (should_uninitialize) CoUninitialize();
	return res;
}

static void free_native_file_dialog_result(dialog_result_t *result) {
	if (result == NULL) {
		return;
	}
	clear_result_paths(result);
	if (result->selected_options_utf8 != NULL) {
		free(result->selected_options_utf8);
		result->selected_options_utf8 = NULL;
	}
	if (result->error_utf8 != NULL) {
		free(result->error_utf8);
		result->error_utf8 = NULL;
	}
}

#endif
