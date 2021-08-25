package actor

import (
	"github.com/HPISTechnologies/component-lib/log"
	"github.com/HPISTechnologies/component-lib/streamer"
)

type MessageWrapper struct {
	MsgBroker      *streamer.StatefulStreamer
	LatestMessage  *Message
	WorkThreadName string
}

func (wrapper *MessageWrapper) SendMessage(msg *Message, from byte) {
	msgId := log.Logger.GetLogId()

	msg.From = from
	msg.Msgid = msgId

	wrapper.MsgBroker.Send(msg.Name, msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		msg.Name,
		wrapper.WorkThreadName,
		"send msg "+msg.Name,
		"msg",
		msg.Height,
		msg.Round,
		msg.Msgid,
		0)
}

func (wrapper *MessageWrapper) Send(name string, data interface{}) {
	msgId := log.Logger.GetLogId()
	msg := Message{
		Msgid:  msgId,
		Name:   name,
		Height: wrapper.LatestMessage.Height,
		Round:  wrapper.LatestMessage.Round,
		Data:   data,
	}
	wrapper.MsgBroker.Send(name, &msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		name,
		wrapper.WorkThreadName,
		"send msg "+name,
		"msg",
		wrapper.LatestMessage.Height,
		wrapper.LatestMessage.Round,
		wrapper.LatestMessage.Msgid,
		0)
}
