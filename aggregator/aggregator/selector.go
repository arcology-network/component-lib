package aggregator

import (
	ethCommon "github.com/arcology-network/3rd-party/eth/common"
	"github.com/arcology-network/common-lib/common"
)

type Selector struct {
	clearance    []*ethCommon.Hash
	selectedList []*interface{}
	missingList  *MissingList
	pool         *DataPool
}

// NewDataPool returns a new DataPool structure.
func NewSelector() *Selector {
	return &Selector{
		clearance:    []*ethCommon.Hash{},
		selectedList: []*interface{}{},
		missingList:  NewMissingList(),
		pool:         NewDataPool(),
	}
}

//load data into missing when list received
func (s *Selector) GenerateMissing(hashs Hashable) int {
	s.missingList = NewMissingList()
	selectList, clearance := hashs.GetList()
	s.clearance = clearance
	s.selectedList = make([]*interface{}, len(selectList))

	worker := func(start, end, idx int, args ...interface{}) {
		selectingList := args[0].([]interface{})[0].([]*ethCommon.Hash)
		selectedList := args[0].([]interface{})[1].([]*interface{})
		missingist := args[0].([]interface{})[2].(*MissingList)
		for i := start; i < end; i++ {
			hash := selectingList[i]
			obj := s.pool.get(*hash)
			if obj != nil {
				selectedList[i] = &obj
			} else {
				missingist.Put(hash, i)
			}
		}
	}
	common.ParallelWorker(len(selectList), 4, worker, selectList, s.selectedList, s.missingList)
	s.missingList.CompleteGnereation()
	return s.missingList.Size()
}

//received a new data
func (s *Selector) OnDataReceived(h ethCommon.Hash, data interface{}) {
	s.pool.add(h, data)
	if s.missingList.IsGenerated() {
		if ok, idx := s.missingList.RemoveFromMissing(&h); ok {
			s.selectedList[idx] = &data
		}
	}
}

//clear from pool
func (s *Selector) Clear() {
	for _, h := range s.clearance {
		s.pool.remove(*h)
	}
}

//remain quantity in pool
func (s *Selector) Remaining() int {
	return s.pool.count()
}

//get selected data if gather terminated
func (s *Selector) GetSelected() (bool, *[]*interface{}) {
	if s.missingList.IsReceiveCompleted() {
		retobjs := s.selectedList
		s.missingList.ClearList()
		s.selectedList = []*interface{}{}
		return true, &retobjs
	} else {
		return false, nil
	}

}

//reap txs with no list
func (s *Selector) Reap(max int) (*[]*ethCommon.Hash, *[]*interface{}) {
	hashs := []*ethCommon.Hash{}
	objs := []*interface{}{}
	selected := 0
	f := func(hash ethCommon.Hash, obj interface{}) bool {
		hashs = append(hashs, &hash)
		objs = append(objs, &obj)
		selected++

		if max > 0 && selected >= max {
			return false
		}
		return true
	}
	s.pool.Range(f)
	return &hashs, &objs
}

//set clear list
func (s *Selector) SetClearance(hashs Hashable) {
	_, s.clearance = hashs.GetList()
}
