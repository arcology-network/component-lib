package aggregator

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/arcology-network/3rd-party/eth/common"
	ethCommon "github.com/arcology-network/3rd-party/eth/common"
	"github.com/arcology-network/common-lib/types"
)

func Test_Inclusive(t *testing.T) {
	size := 2
	hashlist := make([]*common.Hash, size)
	successList := make([]bool, size)

	for i := range hashlist {
		hash := common.BytesToHash([]byte{byte(i), byte(i + 1), byte(i + 2)})
		hashlist[i] = &hash
		successList[i] = true
	}

	inclusive := types.InclusiveList{
		HashList:   hashlist,
		Successful: successList,
		Mode:       types.InclusiveMode_Results,
	}

	ag := NewAggregator()

	for i := range hashlist {
		fmt.Printf("hashlist[%v]=%x\n", i, hashlist[i])
		ag.OnDataReceived(*hashlist[i], "ok")
	}
	ag.OnDataReceived(common.BytesToHash([]byte{byte(20)}), "ok")

	ag.OnDataReceived(common.BytesToHash([]byte{byte(21)}), "ok")
	ag.OnDataReceived(common.BytesToHash([]byte{byte(22)}), "ok")
	ag.OnDataReceived(common.BytesToHash([]byte{byte(23)}), "ok")

	datas := ag.selector.pool.data
	for k, v := range datas {
		fmt.Printf("k=%x,v=%v\n", k, v)
	}

	result, remain := ag.OnListReceived(&inclusive)
	if result == nil {
		fmt.Printf("nil\n")
	}
	fmt.Printf("result=%v  remain=%v\n", result, remain)

	remain = ag.OnClearInfoReceived()
	fmt.Printf(" remain=%v\n", remain)
}
func Test_mapTest(t *testing.T) {
	hashes := make(map[ethCommon.Hash]int, 500000)
	startTime := time.Now()
	for i := 0; i < 500000; i++ {
		hash := ethCommon.BigToHash(big.NewInt(int64(i)))
		hashes[hash] = i
	}
	fmt.Printf("init maptime=%v\n", time.Now().Sub(startTime))

	startTime2 := time.Now()
	hashes2 := map[ethCommon.Hash]int{}
	for i := 0; i < 500000; i++ {
		hash := ethCommon.BigToHash(big.NewInt(int64(i)))
		hashes2[hash] = i
	}
	fmt.Printf("init map 2 time=%v\n", time.Now().Sub(startTime2))
}

func Test_arrayTest(t *testing.T) {
	hashes := make([]int, 0, 500000)

	startTime := time.Now()
	for i := 0; i < 500000; i++ {
		hashes = append(hashes, i)
	}
	fmt.Printf("hashs size=%v,time=%v\n", len(hashes), time.Now().Sub(startTime))

	hashess := make([]int, 500000)
	startTime1 := time.Now()
	for i := 0; i < 500000; i++ {
		hashess[i] = i
	}
	fmt.Printf("hashss size=%v,time=%v\n", len(hashess), time.Now().Sub(startTime1))

}
func Test_mapprofermance(t *testing.T) {
	missingList := map[ethCommon.Hash]int{}
	hashes := make([]ethCommon.Hash, 500000)
	for i := range hashes {
		hash := ethCommon.BigToHash(big.NewInt(int64(i)))
		hashes[i] = hash
	}
	startTime := time.Now()
	for i := range hashes {
		missingList[hashes[i]] = i
	}
	fmt.Printf("map time=%v\n", time.Now().Sub(startTime))
}
func Test_profermance(t *testing.T) {
	ag := NewAggregator()

	hashes := make([]*ethCommon.Hash, 500000)
	for i := range hashes {
		hash := ethCommon.BigToHash(big.NewInt(int64(i)))
		hashes[i] = &hash
		ag.OnDataReceived(hash, hash)
	}

	reapinglist := &types.ReapingList{
		List: hashes,
	}

	startTime := time.Now()
	ag.OnListReceived(reapinglist)
	fmt.Printf("time=%v\n", time.Now().Sub(startTime))
}
