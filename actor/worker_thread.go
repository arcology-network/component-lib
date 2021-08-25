package actor

import (
	"errors"

	"github.com/HPISTechnologies/component-lib/log"
	"github.com/HPISTechnologies/component-lib/streamer"
	"go.uber.org/zap/zapcore"
)

type IWorker interface {
	Init(workThreadName string, broker *streamer.StatefulStreamer)
	ChangeEnvironment(message *Message)
	OnStart()
	OnMessageArrived(msgs []*Message) error
}

type WorkerThreadLogger struct {
	LatestMessage *Message
	Logger        *log.LogWraper
}

func (workerThdLogger *WorkerThreadLogger) Log(level string, info string, fields ...zapcore.Field) uint64 {
	return workerThdLogger.Logger.Log(level, workerThdLogger.LatestMessage, info, fields...)
}

func (workerThdLogger *WorkerThreadLogger) GetLogger(refid uint64) *WorkerThreadLogger {
	newLog := *workerThdLogger
	newLog.LatestMessage.Msgid = refid
	return &newLog
}

type WorkerThread struct {
	Name          string
	Groupid       string
	Concurrency   int
	MsgBroker     *MessageWrapper
	LatestMessage *Message
	Log           *log.LogWraper
}

//use when function is called
func (workerThd *WorkerThread) GetLogger(refid uint64) *WorkerThreadLogger {
	innerMessage := *workerThd.LatestMessage
	return &WorkerThreadLogger{
		LatestMessage: &innerMessage,
		Logger:        workerThd.Log,
	}
}

//use where is in worklthread
func (workerThd *WorkerThread) AddLog(level string, info string, fields ...zapcore.Field) uint64 {
	return workerThd.Log.Log(level, workerThd.LatestMessage, info, fields...)
}

func (workerThd *WorkerThread) Init(workThreadName string, broker *streamer.StatefulStreamer) {
	workerThd.Name = workThreadName
	latestMessage := NewMessage()

	workerThd.MsgBroker = &MessageWrapper{
		MsgBroker:      broker,
		LatestMessage:  latestMessage,
		WorkThreadName: workThreadName,
	}

	workerThd.LatestMessage = latestMessage
	workerThd.Log = log.Logger.GetLogger(workThreadName)
}

func (workerThd *WorkerThread) Set(concurrency int, groupid string) {
	workerThd.Concurrency = concurrency
	workerThd.Groupid = groupid
}

func (workerThd *WorkerThread) ChangeEnvironment(message *Message) {
	if message.Height > workerThd.LatestMessage.Height ||
		(message.Height == workerThd.LatestMessage.Height && message.Round > workerThd.LatestMessage.Round) {
		workerThd.LatestMessage = message
		workerThd.MsgBroker.LatestMessage = message
	}
}

func (workerThd *WorkerThread) IsNil(param interface{}, paramName string) (bool, error) {
	if param == nil {
		workerThd.AddLog(log.LogLevel_Error, workerThd.Name+" received params err:"+paramName+" is nil")
		return true, errors.New(workerThd.Name + " received params err:" + paramName + " is nil")
	}
	return false, nil
}
