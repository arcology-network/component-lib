package lib

import (
	"sync"
	"time"
)

//type Waitobjs map[int64]WaitObject
type Waitobjs struct {
	Wbs map[int64]*WaitObject
	mtx sync.Mutex
}

//start a wait objects
func StartWaitObjects() *Waitobjs {
	return &Waitobjs{
		Wbs: map[int64]*WaitObject{},
	}
}

//add a object for wait return
func (ws *Waitobjs) AddWaiter(msgid int64) *WaitObject {
	ws.mtx.Lock()
	defer ws.mtx.Unlock()

	ob := WaitObject{}
	ob.wg.Add(1)
	ob.data = nil
	ws.Wbs[msgid] = &ob
	return &ob
}

// set return data
func (ws *Waitobjs) Update(msgid int64, data interface{}) {
	if ob, ok := ws.Wbs[msgid]; ok {
		ob.mtx.Lock()
		defer ob.mtx.Unlock()

		ob.data = data
		ob.wg.Done()

	}
}

// return data
func (ws *Waitobjs) GetData(msgid int64) interface{} {
	ws.mtx.Lock()
	defer ws.mtx.Unlock()

	if ob, ok := ws.Wbs[msgid]; ok {

		retdata := ob.data
		delete(ws.Wbs, msgid)
		return retdata
	}
	return nil
	//return rd.data
}

// start waiting
func (ws *Waitobjs) Waitforever(msgid int64) {
	if rd, ok := ws.Wbs[msgid]; ok {
		rd.wg.Wait()
	}
}

// start waiting
func (ws *Waitobjs) WaitTo(msgid int64, timeout time.Duration) {
	if rd, ok := ws.Wbs[msgid]; ok {
		go rd.startTimer(timeout)
		rd.wg.Wait()
	}
}

// sync mode , save return data
type WaitObject struct {
	//code   int            //ret code 0-success 1-timeout  -1-default
	data interface{}    //result
	wg   sync.WaitGroup //sync mode,flag
	mtx  sync.Mutex
}

// set return data
func (rd *WaitObject) Update(data interface{}) {
	rd.mtx.Lock()
	defer rd.mtx.Unlock()

	rd.data = data
	rd.wg.Done()

}

// return data
func (rd *WaitObject) Get() interface{} {
	return rd.data
}

// start waiting
func (rd *WaitObject) Waitforever() {
	rd.wg.Add(1)
	rd.data = nil
	rd.wg.Wait()
}

// start waiting
func (rd *WaitObject) WaitTo(timeout time.Duration) {
	rd.wg.Add(1)
	rd.data = nil
	go rd.startTimer(timeout)

	rd.wg.Wait()
}

// start  a timer for async
func (rd *WaitObject) startTimer(timeout time.Duration) {
	reqTimer := time.NewTimer(timeout)
	for {
		select {
		case <-reqTimer.C:
			rd.timeout()
		}
	}
}

// timeout result
func (rd *WaitObject) timeout() {
	rd.mtx.Lock()
	defer rd.mtx.Unlock()

	rd.data = nil
	rd.wg.Done()

}
