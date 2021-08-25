// +build !nometri

package mhasher

/*
#cgo CFLAGS: -Iinclude

#cgo LDFLAGS: -Lso -lmhasher

#include "bytes.external.h"

#include "stdlib.h"

*/
import "C" //must flow above
import (
	"unsafe"

	"github.com/HPISTechnologies/common-lib/codec"
)

func SortBytes(datas [][]byte) ([][]byte, error) {
	totalBytes, lengthVec, dataLength := orderData(datas)

	indices := make([]uint32, dataLength)

	_, err := C.Sort(
		(*C.char)(unsafe.Pointer(&totalBytes[0])),
		(*C.uint32_t)(unsafe.Pointer(&lengthVec[0])),
		C.uint32_t(dataLength),
		(*C.uint32_t)(unsafe.Pointer(&indices[0])),
	)
	if err != nil {
		return datas, err
	}
	results := make([][]byte, dataLength)
	for i, idx := range indices {
		results[i] = datas[int(idx)]
	}
	return results, nil
}

func UniqueBytes(datas [][]byte) ([][]byte, error) {
	totalBytes, lengthVec, dataLength := orderData(datas)

	indices := make([]uint8, dataLength)

	_, err := C.Unique(
		(*C.char)(unsafe.Pointer(&totalBytes[0])),
		(*C.uint32_t)(unsafe.Pointer(&lengthVec[0])),
		C.uint32_t(dataLength),
		(*C.uint8_t)(unsafe.Pointer(&indices[0])),
	)
	if err != nil {
		return datas, err
	}
	results := make([][]byte, 0, dataLength)
	for i, flag := range indices {
		if flag == uint8(255) {
			results = append(results, datas[i])
		}
	}
	return results, nil
}

func RemoveBytes(datas, toRemove [][]byte) ([][]byte, error) {
	totalBytes, lengthVec, dataLength := orderData(datas)
	removeTotalBytes, removeLengthVec, removeDataLength := orderData(toRemove)

	indices := make([]uint8, dataLength)

	_, err := C.Remove(
		(*C.char)(unsafe.Pointer(&totalBytes[0])),
		(*C.uint32_t)(unsafe.Pointer(&lengthVec[0])),
		C.uint32_t(dataLength),
		(*C.char)(unsafe.Pointer(&removeTotalBytes[0])),
		(*C.uint32_t)(unsafe.Pointer(&removeLengthVec[0])),
		C.uint32_t(removeDataLength),
		(*C.uint8_t)(unsafe.Pointer(&indices[0])),
	)
	if err != nil {
		return datas, err
	}
	results := make([][]byte, 0, dataLength)
	for i, flag := range indices {
		if flag == uint8(255) {
			results = append(results, datas[i])
		}
	}
	return results, nil
}

func orderData(datas [][]byte) ([]byte, []uint32, uint32) {
	dataLength := len(datas)
	if dataLength == 0 {
		return []byte{}, []uint32{}, 0
	}
	lengthVec := make([]uint32, dataLength)
	for i, data := range datas {
		lengthVec[i] = uint32(len(data))
	}
	totalBytes := codec.Byteset(datas).Flatten()
	return totalBytes, lengthVec, uint32(dataLength)
}
