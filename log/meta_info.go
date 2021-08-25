package log

import (
	"io/ioutil"
	"sync"
)

var Metas *MetaInfos

type MetaInfoItem struct {
	ReceptivMsgs    []string
	DeliverableMsgs []string
}

type MetaInfos struct {
	Mis         map[string]*MetaInfoItem
	WorkThreads []string
	lock        sync.RWMutex
}

func InitMetaInfos() {
	Metas = &MetaInfos{
		Mis: map[string]*MetaInfoItem{},
	}
}

func (mi *MetaInfos) Add(workthread, msgname string, isSend bool) {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	mii, ok := mi.Mis[workthread]
	if !ok {
		mii = &MetaInfoItem{
			ReceptivMsgs:    []string{},
			DeliverableMsgs: []string{},
		}
		mi.Mis[workthread] = mii
		mi.WorkThreads = append(mi.WorkThreads, workthread)
	}
	if isSend {
		appendString(&mii.DeliverableMsgs, msgname)
	} else {
		appendString(&mii.ReceptivMsgs, msgname)
	}

}

func (mi *MetaInfos) MetaInfoToFile(servicename, configfilePath string) error {
	mi.lock.Lock()
	defer mi.lock.Unlock()

	outputString := ""
	for _, workthreadname := range mi.WorkThreads {
		if metainfoItem, ok := mi.Mis[workthreadname]; ok {
			for _, msgname := range metainfoItem.DeliverableMsgs {
				outputString += servicename + "\t" + workthreadname + "\t" + msgname + "\tsend" + "\n"
			}
			for _, msgname := range metainfoItem.ReceptivMsgs {
				outputString += servicename + "\t" + workthreadname + "\t" + msgname + "\treceive" + "\n"
			}
		}
	}

	return ioutil.WriteFile(configfilePath+"/"+servicename+".conf", []byte(outputString), 0666)

}

func appendString(list *[]string, one string) {
	for _, v := range *list {
		if v == one {
			return
		}
	}
	(*list) = append((*list), one)
}
