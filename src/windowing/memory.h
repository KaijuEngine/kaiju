#ifndef CLIB_MEMORY_H
#define CLIB_MEMORY_H

#include <string.h>

#ifndef SWAP
#define SWAP(TYPE,A,B) {TYPE t=A; A=B; B=t;}
#endif

#ifndef MEMSETZERO
#define MEMSETZERO(a) memset(&(a), 0, sizeof(a));
#endif

#endif