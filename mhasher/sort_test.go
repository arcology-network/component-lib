package mhasher

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSortBytes(t *testing.T) {
	datas := [][]byte{[]byte{1, 2, 4}, []byte{5, 6, 7, 8}, []byte{9, 10, 12}}

	result, err := SortBytes(datas)
	if err != nil {
		t.Error("sort err:" + err.Error())
	}
	if !reflect.DeepEqual(result[0], datas[0]) || !reflect.DeepEqual(result[1], datas[2]) || !reflect.DeepEqual(result[2], datas[1]) {
		t.Error("Wrong order !")
	}
	fmt.Printf("result=%v\n", result)
}

func TestUniqueBytes(t *testing.T) {
	datas := [][]byte{[]byte{1, 2, 4}, []byte{5, 6, 7, 8}, []byte{9, 10, 12}, []byte{5, 6, 7, 8}}

	result, err := UniqueBytes(datas)
	if err != nil {
		t.Error("Unique err:" + err.Error())
	}
	if !reflect.DeepEqual(len(result), 3) || !reflect.DeepEqual(result[0], datas[0]) || !reflect.DeepEqual(result[1], datas[1]) || !reflect.DeepEqual(result[2], datas[2]) {
		t.Error("Wrong Unique !")
	}
	fmt.Printf("result=%v\n", result)
}

func TestRemoveBytes(t *testing.T) {
	datas := [][]byte{[]byte{1, 2, 4}, []byte{5, 6, 7, 8}, []byte{9, 10, 12}}
	toRemove := [][]byte{[]byte{1, 2, 4}, []byte{5, 6, 7, 8}}
	result, err := RemoveBytes(datas, toRemove)
	if err != nil {
		t.Error("Unique err:" + err.Error())
	}
	fmt.Printf("result=%v\n", result)
	if !reflect.DeepEqual(len(result), 1) || !reflect.DeepEqual(result[0], datas[2]) {
		t.Error("Wrong Remove !")
	}

}
