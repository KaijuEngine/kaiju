#if defined(_WIN32) || defined(_WIN64)
#define _CRT_SECURE_NO_DEPRECATE
#endif

#include "strings.h"
#include "memory.h"

#ifdef TESTING
#include <assert.h>
#endif

#if !defined(MAX)
#define MAX(a, b) (a) > (b) ? (a) : (b)
#endif

#if !defined(MIN)
#define MIN(a, b) (a) < (b) ? (a) : (b)
#endif

#ifndef STR_NO_16
/******************************************************************************/
/******************************************************************************/
/* WCHAR and char16_t functions                                               */
/******************************************************************************/
/******************************************************************************/
size_t strlen16(const char16_t* str) {
	if (str == NULL)
		return 0;
	char16_t* s = (char16_t*)str;
	for (; *s; ++s);
	return s - str;
}

void wchartou8(const wchar_t* str, char** outStr) {
	mbstate_t state;
	memset(&state, 0, sizeof state);
	const wchar_t* wStr = (wchar_t*)str;
	size_t len = 1 + wcsrtombs(NULL, &wStr, 0, &state);
	*outStr = (char*)malloc(len + 1);
	wcsrtombs(*outStr, &wStr, len, &state);
}

void u8towchar(const char* str, wchar_t** outStr) {
	size_t len = mbstowcs(NULL, str, 0);
	wchar_t* wStr = (wchar_t*)malloc((len + 1) * sizeof(wchar_t));
	mbstowcs(wStr, str, len);
	*(wStr + len) = '\0';
	*outStr = (wchar_t*)wStr;	// void* here to avoid compiler warning
}

void str16tou8(const char16_t* str, char** outStr) {
	wchartou8((wchar_t*)str, outStr);
}

void u8tostr16(const char* str, char16_t** outStr) {
	u8towchar(str, (wchar_t**)outStr);
}

void wstrsub(wchar_t* str, wchar_t find, wchar_t replace) {
	const size_t len = wcslen(str);
	for (size_t i = 0; i < len; ++i)
		if (str[i] == find)
			str[i] = replace;
}
#endif

/******************************************************************************/
/******************************************************************************/
/* UTF-8 functions                                                            */
/******************************************************************************/
/******************************************************************************/
static inline size_t utf8_count_size(const char* const str, size_t charCount) {
	size_t len = 0;
	unsigned char c = str[0];
	size_t i = 0;
	for (; c != 0 && len < charCount; ++len) {
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		c = str[i];
	}
	return i;
}

int utf8csize(const char* const str) {
	unsigned char c = str[0];
	int v0 = (c & 0x80) >> 7;
	int v1 = (c & 0x40) >> 6;
	int v2 = (c & 0x20) >> 5;
	int v3 = (c & 0x10) >> 4;
	return 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
}

size_t utf8len(const char* const str) {
	size_t len = 0;
	unsigned char c = str[0];
	for (size_t i = 0; c != 0; ++len) {
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		c = str[i];
	}
	return len;
}

size_t utf8len_s(const char* const str, size_t maxLen) {
	size_t len = 0;
	unsigned char c = str[0];
	for (size_t i = 0; c != 0 && i < maxLen; ++len) {
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		c = str[i];
	}
	return len;
}

bool utf8valid(const char* const str) {
	if (str == NULL)
		return false;
	const char* c = str;
	bool valid = true;
	for (size_t i = 0; c[0] != 0 && valid;) {
		valid = (c[0] & 0x80) == 0
		        || ((c[0] & 0xE0) == 0xC0 && (c[1] & 0xC0) == 0x80)
		        || ((c[0] & 0xF0) == 0xE0 && (c[1] & 0xC0) == 0x80 && (c[2] & 0xC0) == 0x80)
		        || ((c[0] & 0xF8) == 0xF0 && (c[1] & 0xC0) == 0x80 && (c[2] & 0xC0) == 0x80 && (c[3] & 0xC0) == 0x80);
		int v0 = (c[0] & 0x80) >> 7;
		int v1 = (c[0] & 0x40) >> 6;
		int v2 = (c[0] & 0x20) >> 5;
		int v3 = (c[0] & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		c = str + i;
	}
	return valid;
}

bool utf8valid_s(const char* const str, size_t maxLen) {
	if (str == NULL)
		return false;
	unsigned char c = '\0';
	bool valid = true;
	for (size_t i = 0; c != 0 && valid;) {
		valid = (c & 0x80) == 0
		        || (c & 0xE0) == 0xC0
		        || (c & 0xF0) == 0xE0
		        || (c & 0xF8) == 0xF0;
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		c = str[i];
	}
	return true;
}

uint32_t utf8toui(const char* str) {
	unsigned char c = str[0];
	int v0 = (c & 0x80) >> 7;
	int v1 = (c & 0x40) >> 6;
	int v2 = (c & 0x20) >> 5;
	int v3 = (c & 0x10) >> 4;
	int i = 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
	char tmp[5] = { 0, 0, 0, 0, 0 };
	memcpy(tmp, str, i);
	uint32_t val = 0;
	strcpy((char*)&val, tmp);
	return val;
}

uint8_t utf8letter(const char* const str, char out[4]) {
	memset(out, 0, 4);
	if (utf8valid(str)) {
		unsigned char c = str[0];
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		int i = 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		memcpy(out, str, i);
		return i;
	} else
		return 0;
}

void uitoutf8(uint32_t input, char out[4]) {
	strcpy(out, (char*)&input);
}

void utf8each(const char* str, size_t limit,
	void(*forEach)(const char*,uint8_t,size_t,void*), void* state)
{
	unsigned char c = str[0];
	for (size_t i = 0, len = 0; c != 0 && len < limit; ++len) {
		size_t start = i;
		int v0 = (c & 0x80) >> 7;
		int v1 = (c & 0x40) >> 6;
		int v2 = (c & 0x20) >> 5;
		int v3 = (c & 0x10) >> 4;
		i += 1 + v0 * v1 + v0 * v1 * v2 + v0 * v1 * v2 * v3;
		uint8_t l = (uint8_t)(i - start);
		forEach(str + i - l, l, len, state);
		c = str[i];
	}
}

/******************************************************************************/
/******************************************************************************/
/* String modifications                                                       */
/******************************************************************************/
/******************************************************************************/
void trim(char* str) {
	int32_t len = (int32_t)strlen(str);
	int32_t i;
	for (i = 0; i < len; ++i)
		if (is_ascii(str[i]) && !isspace(str[i]))
			break;
	len -= i;
	memmove(str, str + i, len);
	for (i = len - 1; i >= 0; --i)
		if (is_utf8head(str[i]) && (!is_ascii(str[i]) || !isspace(str[i])))
			break;
	str[i + 1] = '\0';
}

void substr(const char* str, int32_t start, int32_t len, char** outStr) {
	*outStr = NULL;
	if (len > 0) {
		start = MAX(0, start);
		size_t byteStart = utf8_count_size(str, start);
		const size_t memLen = utf8_count_size(str + byteStart,
			MIN((int32_t)utf8len(str) - start, len));
		*outStr = (char*)malloc(memLen + 1);
		memcpy(*outStr, str + byteStart, memLen);
		(*outStr)[memLen] = '\0';
	}
}

void strjoin(const char* lhs,
	const char* rhs, const char* glue, char** outStr)
{
	const size_t aLen = lhs != NULL ? strlen(lhs) : 0;
	const size_t bLen = rhs != NULL ? strlen(rhs) : 0;
	if (glue == NULL) {
		const size_t len = aLen + bLen + 1;
		char* joined = (char*)malloc(len);
		memcpy(joined, lhs, aLen);
		memcpy(joined + aLen, rhs, bLen);
		joined[len - 1] = '\0';
		*outStr = joined;
	} else {
		const size_t gLen = strlen(glue);
		const size_t len = aLen + bLen + gLen + 1;
		char* joined = (char*)malloc(len);
		memcpy(joined, lhs, aLen);
		memcpy(joined + aLen, glue, gLen);
		memcpy(joined + aLen + gLen, rhs, bLen);
		joined[len - 1] = '\0';
		*outStr = joined;
	}
}

void strclone(const char* str, char** outStr) {
	char* s = (char*)malloc(strlen(str) + 1);
	strcpy(s, str);
	*outStr = s;
}

void strcloneclr(const char* str, char** outStr) {
	char* s = (char*)malloc(strlen(str) + 1);
	strcpy(s, str);
	free(*outStr);
	*outStr = s;
}

void strsub(char* str, char find, char replace) {
	const size_t len = strlen(str);
	for (size_t i = 0; i < len; ++i)
		if (str[i] == find)
			str[i] = replace;
}

/******************************************************************************/
/******************************************************************************/
/* String queries                                                             */
/******************************************************************************/
/******************************************************************************/
int32_t stridxof(const char* str, const char* search, int32_t offset) {
	int32_t nLen = (int32_t)strlen(search);
	int32_t hLen = (int32_t)strlen(str) - nLen;
	for (int32_t h = offset; h <= hLen; ++h) {
		for (int32_t n = 0; n < nLen; ++n) {
			if (str[h + n] != search[n])
				break;
			else if (n == nLen - 1)
				return h;
		}
	}
	return -1;
}

int32_t stridxofi(const char* str, const char* search, int32_t offset) {
	int32_t nLen = (int32_t)strlen(search);
	int32_t hLen = (int32_t)strlen(str) - nLen;
	for (int32_t h = offset; h <= hLen; ++h) {
		for (int32_t n = 0; n < nLen; ++n) {
			if (tolower(str[h + n]) != tolower(search[n]))
				break;
			else if (n == nLen - 1)
				return h;
		}
	}
	return -1;
}

int32_t strpos(const char* str, char search) {
	int idx = -1;
	for (int i = 0; str[i] != '\0' && idx == -1; ++i) {
		if (str[i] == search)
			idx = i;
	}
	return idx;
}

int32_t stridxoflast(const char* str, const char* search, int32_t offset) {
	int32_t nLen = (int32_t)strlen(search);
	int32_t hLen = (int32_t)strlen(str) - nLen;
	for (int32_t h = hLen - offset; h >= 0; --h) {
		for (int32_t n = 0; n < nLen; ++n) {
			if (str[h + n] != search[n])
				break;
			else if (n == nLen - 1)
				return h;
		}
	}
	return -1;
}

int strcmpend(const char* str, const char* search) {
	const size_t l = strlen(str);
	const size_t r = strlen(search);
	if (r > l)
		return strcmp(str, search);
	return strcmp(str + (l - r), search);
}

int32_t strcount(const char* str, unsigned char needle) {
	int32_t count = 0;
	for (size_t i = 0; i < strlen(str); ++i)
		// C specification states equality returns (1) for true or (0) for false
		count += str[i] == needle;
	return count;
}

bool strisnum(const char* str) {
	return strisdec(str) || strisbin(str);
}

bool strisdec(const char* str) {
	const size_t len = strlen(str);
	size_t i = str[0] == '-' ? 1 : 0;
	if ((len == i + 1 && str[i] == '0'))
		return true;
	if (str[i] == '0' && str[i + 1] != '.')
		return false;
	for (int32_t d = 0; i < len; ++i) {
		if (str[i] == '.') {
			if (++d > 1)
				return false;
		}
		else if (!isdigit(str[i]))
			return false;
	}
	return true;
}

bool strishex(const char* str) {
	const size_t len = strlen(str);
	if (len < 3 || str[0] != '0' || tolower(str[1]) != 'x')
		return false;
	for (size_t i = 2; i < len; i++) {
		const char c = tolower(str[i]);
		if (!isdigit(c) && (c < 'a' || c > 'f'))
			return false;
	}
	return true;
}

int ftostr(float value, int maxDecimals, char** outStr) {
	char str[128];
	int len = snprintf(str, sizeof(str), "%.*f", maxDecimals, value);
	for (int i = len - 1; i >= 0; --i, --len) {
		if (str[i] == '.') {
			str[i] = '\0';
			break;
		} else if (str[i] != '0')
			break;
		str[i] = '\0';
	}
	strclone(str, outStr);
	return len;
}

int ftostr_s(char* outStr, size_t outStrMaxLen, float value, int maxDecimals) {
	int len = snprintf(outStr, outStrMaxLen, "%.*f", maxDecimals, value);
	for (int i = len - 1; i > 0; --i, --len) {
		if (outStr[i] == '.') {
			outStr[i] = '\0';
			break;
		} else if (outStr[i] != '0')
			break;
		outStr[i] = '\0';
	}
	return len;
}

bool strisbin(const char* str) {
	const size_t len = strlen(str);
	if (len < 3 || str[0] != '0' || tolower(str[1]) != 'b')
		return false;
	for (size_t i = 2; i < len; i++)
		if (str[i] != '0' && str[i] != '1')
			return false;
	return true;
}

bool streqi(const char* a, const char* b) {
	size_t lenA = strlen(a);
	size_t lenB = strlen(b);
	bool same = lenA == lenB;
	for (size_t i = 0; i < lenA && same; ++i)
		same = tolower((int)a[i]) == tolower((int)b[i]);
	return same;
}

int32_t strsplit(const char* a, unsigned char delimiter, char*** out) {
	*out = NULL;
	if (a == NULL)
		return 0;
	int32_t len = (int32_t)strlen(a);
	if (len == 0)
		return 0;
	int32_t count = strcount(a, delimiter) + 1;
	char** parts = (char**)malloc(sizeof(char*) * count);
	int32_t start = 0;
	int32_t pIdx = 0;
	for (int32_t i = 0; i < len; ++i) {
		if (a[i] == delimiter) {
			substr(a, start, i - start, &(parts[pIdx]));
			start = ++i;
			pIdx++;
		}
	}
	substr(a, start, len - start, &(parts[pIdx]));
	*out = parts;
	return count;
}

void strsplice(const char* str, int32_t start, int32_t len, char** outStr) {
	size_t sLen = strlen(str);
	char* newStr = malloc((sLen - len) + 1);
	*outStr = newStr;
	for (size_t i = 0, idx = 0; i < sLen; ++i, ++idx) {
		if (i == (size_t)start)
			i += len;
		else
			newStr[idx] = str[i];
	}
	newStr[sLen - len] = '\0';
}

void strfree(char* str) { free(str); }

int strdiff(const char* search, const char* against) {
	// Levenshtein distance
	int m = (int)strlen(search);
	int n = (int)strlen(against);
	if (m == 0)
		return n;
	else if (n == 0)
		return m;
	else if (m == n && streq(search, against))
		return 0;
	int* costs = malloc(sizeof(int) * (n + 1));
	for (int k = 0; k <= n; ++k)
		costs[k] = k;
	for (int i = 0; i < m; ++i) {
		costs[0] = i + 1;
		int corner = i;
		for (int j = 0; j < n; ++j) {
			int upper = costs[j + 1];
			if (search[i] == against[j])
				costs[j + 1] = corner;
			else {
				int t = upper < corner ? upper : corner;
				costs[j + 1] = (costs[j] < t ? costs[j] : t) + 1;
			}
			corner = upper;
		}
	}
	int r = costs[n];
	free(costs);
	return r;
}

int strdiffi(const char* search, const char* against) {
	// Levenshtein distance
	int m = (int)strlen(search);
	int n = (int)strlen(against);
	if (m == 0)
		return n;
	else if (n == 0)
		return m;
	else if (m == n && streqi(search, against))
		return 0;
	int* costs = malloc(sizeof(int) * (n + 1));
	for (int k = 0; k <= n; ++k)
		costs[k] = k;
	for (int i = 0; i < m; ++i) {
		costs[0] = i + 1;
		int corner = i;
		for (int j = 0; j < n; ++j) {
			int upper = costs[j + 1];
			if (tolower(search[i]) == tolower(against[j]))
				costs[j + 1] = corner;
			else {
				int t = upper < corner ? upper : corner;
				costs[j + 1] = (costs[j] < t ? costs[j] : t) + 1;
			}
			corner = upper;
		}
	}
	int r = costs[n];
	free(costs);
	return r;
}

double strasdouble(const char* str) {
	if (str[0] == '0' && tolower(str[1]) == 'x') {
		// Convert the hexidecimal string to a double
		unsigned long int v = strtoul(str, NULL, 16);
		return (double)v;
	} else if (str[0] == '0' && tolower(str[1]) == 'b') {
		// Convert the binary string to a double
		unsigned long int v = strtoul(str, NULL, 2);
		return (double)v;
	} else if (strchr(str, '.') == NULL) {
		// Convert the integer string to a double
		unsigned long int v = strtoul(str, NULL, 10);
		return (double)v;
	} else
		return strtod(str, NULL);
}