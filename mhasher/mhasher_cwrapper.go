// +build !nometri

package mhasher

/*
#cgo CFLAGS: -Iinclude

#cgo LDFLAGS: -Lso -lmhasher

#include "mhasher.external.h"

#include "stdlib.h"

*/
import "C" //must flow above
import (
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"time"
	"unsafe"

	ethCommon "github.com/arcology/3rd-party/eth/common"
	"github.com/arcology/common-lib/encoding"
)

const (
	HashType_160 = 20
	HashType_256 = 32
)

func getRoundBys(size int) []byte {
	t := time.Now()
	h := md5.New()
	io.WriteString(h, "crazyof4335435.me")
	io.WriteString(h, t.String())
	seed := h.Sum(nil)

	seedSize := len(seed)
	roundCount := size / seedSize
	if size%seedSize > 0 {
		roundCount = roundCount + 1
	}
	roundKey := make([]byte, roundCount*seedSize)
	bz := 0
	for i := 0; i < roundCount; i++ {
		bz += copy(roundKey[bz:], seed)
	}
	return roundKey[:size]
}

func UniqueKeysByMap(keys [][]byte) [][]byte {
	mp := make(map[string]int, len(keys))
	for i := range keys {
		mp[string(keys[i])] = i
	}
	outBys := make([][]byte, len(mp))
	i := 0
	for k, _ := range mp {
		outBys[i] = []byte(k)
		i = i + 1
	}
	return outBys
}

func UniqueKeys(keys [][]byte) ([][]byte, error) {
	if len(keys) == 0 {
		return [][]byte{}, nil
	}

	chars := encoding.Byteset(keys).Flatten()
	c_char := (*C.char)(unsafe.Pointer(&chars[0]))

	count := len(keys)
	c_count := C.uint64_t(count)

	step := len(keys[0])

	aIndex := make([]uint64, 1)
	a_index := (*C.uint64_t)(unsafe.Pointer(&aIndex[0]))

	ahash := make([]byte, len(chars))
	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))

	_, err := C.UniqueHash256(c_char, c_count, a_char, a_index)

	outBytes := make([][]byte, int(aIndex[0]))
	if aIndex[0] == 0 {
		return [][]byte{}, errors.New("return data is empty")
	}

	for i := 0; i < int(aIndex[0]); i++ {
		outBytes[i] = ahash[step*i : step*(i+1)]
	}

	return outBytes, err
}

func SortByHash(hashes []ethCommon.Hash) ([]uint64, error) {
	if len(hashes) == 0 {
		return []uint64{}, nil
	}
	chars := ethCommon.Hashes(hashes).Flatten()
	c_char := (*C.char)(unsafe.Pointer(&chars[0]))

	count := len(hashes)
	c_count := C.uint64_t(count)

	rIndex := make([]uint64, count)
	r_index := (*C.uint64_t)(unsafe.Pointer(&rIndex[0]))

	var err error
	_, err = C.SortHash256(c_char, c_count, r_index)

	return rIndex, err
}

func BinaryMhasherFromRaw(srcStr []byte, length int, HashType int) ([]byte, error) {
	ahash := make([]byte, HashType)
	if len(srcStr) == 0 {
		return ahash, nil
	}
	c_char := (*C.char)(unsafe.Pointer(&srcStr[0]))
	clenth := C.ulong(length)

	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))

	var err error
	if HashType == HashType_160 {
		_, err = C.ChecksumRIPEMD160(c_char, clenth, a_char)
	} else {
		_, err = C.ChecksumKecaak256(c_char, clenth, a_char)
	}

	return ahash, err
}

//is used in monaco/core/types/tx.go
func GetHash(src []byte, HashType int) ([]byte, error) {
	length := C.ulong(len(src))
	c_char := (*C.char)(unsafe.Pointer(&src[0]))

	ahash := make([]byte, HashType)
	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))
	var err error
	if HashType == HashType_160 {
		_, err = C.SingleHashRIPEMD160(c_char, length, a_char)
	} else {
		_, err = C.SingleHashKeccak256(c_char, length, a_char)
	}
	return ahash, err
}

//is used in monaco/core/types/tx.go
func Roothash(ls [][]byte, HashType int) ([]byte, error) {
	srcLenth := len(ls)
	var buffer bytes.Buffer
	for i := 0; i < srcLenth; i++ {

		buffer.Write(ls[i])
	}
	srcStr := buffer.Bytes()

	num := C.ulong(srcLenth)

	c_char := (*C.char)(unsafe.Pointer(&srcStr[0]))

	ahash := make([]byte, HashType)
	a_char := (*C.char)(unsafe.Pointer(&ahash[0]))

	var err error
	if HashType == HashType_160 {

		_, err = C.BinaryMhasherRIPEMD160(c_char, num, a_char)
	} else {
		_, err = C.BinaryMhasherKeccak256(c_char, num, a_char)
	}
	return ahash, err
}
