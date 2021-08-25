package mhasher

import (
	"bytes"

	ethCommon "github.com/HPISTechnologies/3rd-party/eth/common"
)

// GetTxsHash get transactions roothash
func GetTxsHash(src2d [][]byte) ethCommon.Hash {
	if len(src2d) == 0 {
		return ethCommon.Hash{}
	}
	src := bytes.Join(src2d, []byte(""))
	totallen := len(src)
	roothash, err := BinaryMhasherFromRaw(src, totallen, HashType_256)
	if err != nil {
		return ethCommon.Hash{}
	}
	return ethCommon.BytesToHash(roothash)
}
