/******************************************************************************/
/* strings.h                                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

#ifndef CLIB_STRING_H
#define CLIB_STRING_H

#include <stdio.h>
#include <ctype.h>
#include <string.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdbool.h>


// On Apple, `uchar.h` may be unavailable depending on toolchain.
#ifdef __APPLE__
#define STR_NO_16 1
#endif

#ifndef STR_NO_16
#include <uchar.h>
#include <wchar.h>
#endif	// STR_NO_16

/**
 * Determines if the given byte of a UTF-8 character is an ASCII character. Only
 * ASCII characters have the the bit pattern 0b0xxxxxxx
 * @param c The character byte to check
 * @return True if the given character byte is ASCII
 */
static inline bool is_ascii(char c) { return (c & 0x80) == 0; }

/**
 * The input character is a part of a UTF-8 character byte array if it starts
 * with the bit sequence 0b10xxxxxx (regular ASCII is never a part)
 * @param c The character byte to check
 * @return True if it is part of a UTF-8 character but is not the first byte
 */
static inline bool is_utf8part(char c) { return (c & 0xC0) == 0x80; }

/**
 * The input character is the head of a UTF-8 character byte array if it does
 * not start with the bit sequence 0b10xxxxxx (works for ASCII)
 * @param c The character byte to check
 * @return True if it is the first byte of a UTF-8 character
 */
static inline bool is_utf8head(char c) { return !is_utf8part(c); }

#ifndef STR_NO_16
size_t strlen16(const char16_t* str);

/**
 * Used to convert a wchar_t string to a UTF-8 string (char*)
 * @param str The string to be converted into UTF-8
 * @param outStr The initialized memory output UTF-8 string
 */
void wchartou8(const wchar_t* str, char** outStr);

/**
 * Used to convert a UTF-8 string to a wchar_t string
 * @param str The string to be converted into wchar_t
 * @param outStr The initialized memory output wchar_t string
 */
void u8towchar(const char* str, wchar_t** outStr);

/**
 * Used to convert a char16_t string to a UTF-8 string (char*)
 * @param str The string to be converted into UTF-8
 * @param outStr The initialized memory output UTF-8 string
 */
void str16tou8(const char16_t* str, char** outStr);

/**
 * Used to convert a UTF-8 string to a char16_t string
 * @param str The string to be converted into char16_t
 * @param outStr The initialized memory output char16_t string
 */
void u8tostr16(const char* str, char16_t** outStr);

void wstrsub(wchar_t* str, wchar_t find, wchar_t replace);
#endif	// STR_NO_16

static inline void utf8_from_unicode(uint32_t unicode, char utf8[4]) {
	if (unicode < 0x80)
		*utf8++ = (char)unicode;
	else if (unicode < 0x800) {
		*utf8++ = (char)(192 + unicode / 64);
		*utf8++ = (char)(128 + unicode % 64);
	} else if (unicode - 0xd800u < 0x800)
		memset(utf8, 0, 4);
	else if (unicode < 0x10000) {
		*utf8++ = (char)(224 + unicode / 4096);
		*utf8++ = (char)(128 + unicode / 64 % 64);
		*utf8++ = (char)(128 + unicode % 64);
	} else if (unicode < 0x110000) {
		*utf8++ = (char)(240 + unicode / 262144);
		*utf8++ = (char)(128 + unicode / 4096 % 64);
		*utf8++ = (char)(128 + unicode / 64 % 64);
		*utf8++ = (char)(128 + unicode % 64);
	} else
		memset(utf8, 0, 4);
}

/**
 * Return the size in bytes for the char at the start of the given string
 * @param str The UTF-8 character to get the size for
 * @return The size in bytes of the UTF-8 character
 */
int utf8csize(const char* str);

/**
 * Count the number of characters that make up a UTF-8 string
 * @param[in] str The UTF-8 string buffer
 * @return The length of the string
*/
size_t utf8len(const char* str);

/**
 * Safely count the number of characters that make up a UTF-8 string
 * @param[in] str The UTF-8 string buffer
 * @param[in] maxLen The maximum length of the input buffer
 * @return The length of the string
*/
size_t utf8len_s(const char* str, size_t maxLen);

/**
 * Determine if the supplied string is a valid UTF-8 string
 * @param[in] str The UTF-8 string buffer
 * @return False if the string is invalid, otherwise true
*/
bool utf8valid(const char* str);

/**
 * Safely determine if the supplied string is a valid UTF-8 string
 * @param[in] str The UTF-8 string buffer
 * @param[in] maxLen The maximum length of the input buffer
 * @return False if the string is invalid, otherwise true
*/
bool utf8valid_s(const char* str, size_t maxLen);

/**
 * Convert the UTF-8 character provided into a uint32_t value
 * @param str The character to convert to an unsigned int
 * @return The value of the unsigned int character
 */
uint32_t utf8toui(const char* str);

/**
 * Get a single character from the UTF-8 string and return it as the out
 * @param str The input string starting position
 * @param out The output letter
 * @return The number of bytes that make up the letter
 */
uint8_t utf8letter(const char* str, char out[4]);

/**
 * Convert the integer code points to UTF-8 string
 * @param input Integer code points
 * @param out The output string
 */
void uitoutf8(uint32_t input, char out[4]);

/**
 * Go through each letter in a UTF-8 string and execute the given function
 * @param str The string to loop through
 * @param limit The number of characters to read before exiting
 * @param forEach The function to execute for each UTF-8 letter
 * @param state The state for the function to reference
 */
void utf8each(const char* str, size_t limit,
	void(*forEach)(const char* letter, uint8_t byteSize, size_t idx, void* state),
	void* state);

/**
 * Modifies the supplied string to trim the beginning and by removing any
 * occurrence of \n, \t, \r, or spaces using the pre-existing memory space
 * @param str The string to be modified by trimming "white space"
 */
void trim(char* str);

/**
 * Using a provided string, copy a section of that string into a new string
 * @param str The string to copy a section from
 * @param start The left offset to start copying from
 * @param length The number of bytes to copy from the source string
 * @param outStr The initialized memory output string
 */
void substr(const char* str, int32_t start, int32_t len, char** outStr);

/**
 * Join 2 strings together using the "glue" as the string that separates them
 * into a single string
 * @param lhs The left side of the string
 * @param rhs The right side of the string
 * @param glue The string to be copied in-between the lhs and rhs strings
 * @param outStr The initialized memory output string
 */
void strjoin(const char* lhs,
	const char* rhs, const char* glue, char** outStr);

/**
 * Allocates enough memory into address pointed to by outStr and copies the
 * contents of the string into the newly allocated memory.
 */
void strclone(const char* str, char** outStr);

/**
 * Allocates enough memory into address pointed to by outStr and copies the
 * contents of the string into the newly allocated memory. This will also free
 * the existing string if needed
 */
void strcloneclr(const char* str, char** outStr);

void strsub(char* str, char find, char replace);

/**
 * Find the index of a matching sub-string (search) within a supplied string
 * @param str The string used for searching
 * @param search The string to match within the search
 * @param offset The left offset to start searching from (0 for beginning)
 * @return The index in the string that matches the search or -1 if not found
 */
int32_t stridxof(const char* str, const char* search, int32_t offset);
int32_t stridxofi(const char* str, const char* search, int32_t offset);
int32_t strpos(const char* str, char search);

/**
 * Find the index of a matching sub-string (search) within a supplied string
 * @param str The string used for searching
 * @param search The string to match within the search
 * @param offset The right offset to start searching from (0 for end)
 * @return The index in the string that matches the search or -1 if not found
 */
int32_t stridxoflast(const char* str, const char* search, int32_t offset);

/**
 * Compares the end of the supplied string for a search string. If a match is
 * found then 0 will be returned, otherwise -1 or 1 if the supplied string is
 * less than or greater than respectively (same as stdlib strcmp function)
 * @param str The string to be searched within
 * @param search The string to compare the end of str with
 * @return 0 if matching end, -1 if str is <, otherwise 1 (same as strcmp)
 */
int strcmpend(const char* str, const char* search);

/**
 * Count the occurrences of the character in the supplied string
 * @param str The string to search within
 * @param needle The character to search for
 * @return The number of times the searched character occurs
 */
int32_t strcount(const char* str, unsigned char needle);

bool strisnum(const char* str);
/**
 * Determines if the supplied string represents a decimal number
 * @param str The string to check
 * @return True if the string is a decimal representation
 */
bool strisdec(const char* str);

/**
 * Determines if the supplied string represents a hexadecimal number. The string
 * should be prefixed with 0x
 * @param str The string to check
 * @return True if the string is a hexadecimal representation
 */
bool strishex(const char* str);

/**
 * Converts a float into a string, limiting decimals and removing trailing 0s
 * @param value The value to convert to a string
 * @param maxDecimals The maximum number of decimal places for the float
 * @param outStr The string generated from the float
 * @return The length of the string
 */
int ftostr(float value, int maxDecimals, char** outStr);

/**
 * Converts a float into a string, limiting decimals and removing trailing 0s,
 * this version of ftostr uses pre-allocated memory to do the assignment
 * @param outStr The string generated from the float
 * @param outStrMaxLen The maximum size of the memory pointed to by outStr
 * @param value The value to convert to a string
 * @param maxDecimals The maximum number of decimal places for the float
 * @return The length of the string
 */
int ftostr_s(char* outStr, size_t outStrMaxLen, float value, int maxDecimals);

/**
 * Determines if the supplied string represents a binary number. The string
 * should be prefixed with 0b
 * @param str The string to check
 * @return True if the string is a binary representation
 */
bool strisbin(const char* str);

/**
 * Returns: True if the strings match with ignored casing
 */
bool streqi(const char* a, const char* b);

int32_t strsplit(const char* a, unsigned char delimiter, char*** out);

void strsplice(const char* str, int32_t start, int32_t len, char** outStr);

void strfree(char* str);

/**
 * Check 2 strings and the the score difference between them. The higher the
 * the score, the less alike they are. If the score is 0 then they are the same
 * @return The difference score where higher is less similar
 */
int strdiff(const char* search, const char* against);
int strdiffi(const char* search, const char* against);

double strasdouble(const char* str);

static inline size_t strsize(const char* str) {
	return str == NULL ? 0 : strlen(str);
}

static inline void strtolower(char* str) {
	for (; *str != '\0'; str++)
		*str = tolower(*str);
}

static inline bool strempty(const char* str) {
	return str == NULL || str[0] == '\0';
}

static inline bool strblank(const char* str) {
	if (strempty(str))
		return true;
	else {
		bool blank = true;
		while (*str != '\0' && blank) {
			blank = isspace(*str++);
		}
		return blank;
	}
}

static inline bool streq(const char* a, const char* b) {
	return strcmp(a, b) == 0;
}

static inline bool strstartswith(const char* str, const char* start) {
	return strncmp(str, start, strlen(start)) == 0;
}

static inline bool strendswith(const char* str, const char* end) {
	return stridxoflast(str, end, 0) == (int)(strlen(str) - strlen(end));
}

#ifdef __linux__
static inline char* strtok_s(char* src, char* delim, char** tokState) {
	char* res = strtok(src, delim);
	*tokState = src;
	return res;
}
#endif

#endif