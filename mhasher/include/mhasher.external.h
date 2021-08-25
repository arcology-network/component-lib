#pragma once

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

	void SortHash256(char* bytes, uint64_t count, uint64_t* indices);

	void QuickUniqueHash256(char* bytes, uint64_t inCount, char* uniqueHashes, uint64_t* outCount);
	void UniqueHash256(char* bytes, uint64_t inCount, char* uniqueHashes, uint64_t* outCount);

	/*--- Compute hashes from multiple raw input data sets and then compute the root hash  -----*/
	void ChecksumKecaak256(char* bytes, uint64_t length, char* rootHash);
	void ChecksumRIPEMD160(char* bytes, uint64_t length, char* rootHash);
	void ChecksumSHA3256  (char* bytes, uint64_t length, char* rootHash);
			  

	/*--- Build the binary tree from hashes  -----*/
	void BinaryMhasherKeccak256(char* bytes, uint64_t count, char* hash);
	void BinaryMhasherRIPEMD160(char* bytes, uint64_t count, char* hash);
	void BinaryMhasherSHA3256  (char* bytes, uint64_t count, char* hash);

	/*--- Build the Sexdec tree from hashes  -----*/
	void SexdecMhasherKeccak256(char* bytes, uint64_t count, char* hash);

	/*---- Compute hashes from single raw input -----*/
	void SingleHashKeccak256(char* bytes, uint64_t length, char* hash);
	void SingleHashRIPEMD160(char* bytes, uint64_t length, char* hash);
	void SingleHashSHA3256(char* bytes, uint64_t length, char* hash);

	/*-----Compute hashes from multiple raw input---- */
	void MultipleHashesKecaak256(char* bytes, uint64_t* counts, uint64_t length, char* concatenatedHashes);
	void MultipleHashesRIPEMD160(char* bytes, uint64_t* counts, uint64_t length, char* concatenatedHashes);
	void MultipleHashesSHA3256(char* bytes, uint64_t* counts, uint64_t length, char* concatenatedHashes);


	///*============== 2D input  =============== */
	void MultipleHashesKecaak2562D(char** bytes, uint64_t* counts, uint64_t length, char** hashes);
	void MultipleHashesRIPEMD1602D(char** bytes, uint64_t* counts, uint64_t length, char** hashes);
	void MultipleHashesSHA32562D(char** bytes, uint64_t* counts, uint64_t length, char** hashes);
	
	/*2D array*/
	void ChecksumKecaak2562D(char** bytes, uint64_t* lengthVec, uint64_t count, char* rootHash);
	void ChecksumRIPEMD1602D(char** bytes, uint64_t* lengthVec, uint64_t count, char* rootHash);
	void ChecksumSHA32562D(char** bytes, uint64_t* lengthVec, uint64_t count, char* rootHash);

	
	/*-----Product related Info---- */
	void GetVersion(char* ver);
	void GetProduct(char* ver);

#ifdef __cplusplus
}
#endif
