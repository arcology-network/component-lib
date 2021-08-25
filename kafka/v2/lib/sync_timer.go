package lib

import (
	"time"
)

//
type SyncTimer struct {
	// timeOutFun TimerDo
}

type TimerDo func()

//add a object for wait return
func (t *SyncTimer) StartTimer(timeout time.Duration, do TimerDo) {
	go func() {
		tick := time.NewTicker(timeout)
		for {
			select {
			//wait uintl timeout
			case <-tick.C:

				do()

			}
		}
	}()
}
