
#pragma once

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif
	void Sort(char* bytes, uint32_t* lengthVec, uint32_t count, uint32_t* indices);
	void Unique(char* bytes, uint32_t* lengthVec, uint32_t count, uint8_t* newIndices);
	void Remove(char* bytes, uint32_t* lengths, uint32_t count, char* toRemove, uint32_t* toRemoveLengths, uint32_t toRemoveCount, uint8_t* newIndices); //Remove only
#ifdef __cplusplus
}
#endif
