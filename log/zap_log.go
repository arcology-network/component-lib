package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HPISTechnologies/common-lib/common"
	"github.com/HPISTechnologies/common-lib/types"
	"github.com/HPISTechnologies/common-lib/version"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *ZapLog

var levels map[string]zapcore.Level

const (
	LogType_Msg   = "msg"   //message
	LogType_Act   = "act"   //action
	LogType_Inlog = "inlog" //inetrlogic

	LogLevel_Debug      = "debug"
	LogLevel_Info       = "info"
	LogLevel_CheckPoint = "checkpoint"
	LogLevel_Error      = "error"
	LogLevel_Panic      = "panic"

	LogLevel_Executing_Debug = "executing_debug"
)

type ILogHeader interface {
	GetHeader() (uint64, uint64, uint64)
}

type LogWraper struct {
	Logger         *ZapLog
	WorkThreadName string
}

func (logw *LogWraper) Log(level string, msgHeader ILogHeader, info string, fields ...zapcore.Field) uint64 {
	height, round, msgid := msgHeader.GetHeader()
	return logw.Logger.AddLog(0, level, "custom", logw.WorkThreadName, info, LogType_Inlog, height, round, msgid, 0, fields...)
}

func init() {
	levels = map[string]zapcore.Level{}
	levels["debug"] = zap.DebugLevel
	levels["info"] = zap.InfoLevel
	levels["warn"] = zap.WarnLevel
	levels["error"] = zap.ErrorLevel
	levels["panic"] = zap.PanicLevel

	Logger = &ZapLog{}
	Logger.Log, _ = zap.NewDevelopment()
}

func InitLogSystem(logfile, logcfg, svcname string, nodeid int, nodename string) {
	//console,file,redis

	InitCfg(logcfg)

	lookup := map[string]int{}
	notcareds := strings.Split(Conf.LocalFile.Ignored, ",")
	for i, v := range notcareds {
		lookup[v] = i
	}

	Logger = &ZapLog{
		Log:          NewLogger(logfile, svcname, nodename, nodeid),
		sourceLookup: lookup,
	}
	fmt.Printf("--------------------------log init completed------------------------------------------\n")
}

type ZapLog struct {
	Log          *zap.Logger
	sourceLookup map[string]int
	logsChan     chan *types.ExecutingLogsMessage
}

func (log *ZapLog) SetChan(logsChan chan *types.ExecutingLogsMessage) {
	log.logsChan = logsChan

}

func (log *ZapLog) GetLogId() uint64 {
	return common.GenerateUUID()
}

func (log *ZapLog) GetLogger(workThreadName string) *LogWraper {
	return &LogWraper{
		WorkThreadName: workThreadName,
		Logger:         log,
	}
}

func (log *ZapLog) needLog(surce string) bool {
	//log.sourceLookup
	sources := strings.Split(surce, ",")
	for _, sou := range sources {
		if _, ok := log.sourceLookup[sou]; ok {
			return false
		}
	}
	return true
}

//add log
func (log *ZapLog) AddLog(logid uint64, level, source, workthreadname, info, logType string, height, round, refid uint64, dataSize int, fields ...zapcore.Field) uint64 {
	if logid == 0 {
		logid = log.GetLogId()
	}
	if level == LogLevel_Executing_Debug {
		logs := types.ExecutingLogs{}
		err := logs.UnMarshal(info)
		if err == nil {
			msg := types.ExecutingLogsMessage{
				Logs:   logs,
				Height: height,
				Round:  round,
				Msgid:  logid,
			}
			log.logsChan <- &msg
		}
	}
	if log.needLog(source) == false {
		return 0
	}

	fields = append(fields, zap.String("WorkThreadName", workthreadname))
	fields = append(fields, zap.String("Source", source))
	fields = append(fields, zap.Uint64("LogId", logid))
	fields = append(fields, zap.String("LogType", logType))
	fields = append(fields, zap.Uint64("Height", height))
	fields = append(fields, zap.Uint64("Round", round))
	fields = append(fields, zap.Uint64("RefId", refid))
	fields = append(fields, zap.Int("DataSize", dataSize))

	switch level {
	case LogLevel_Debug:
		log.Log.Debug(info, fields...)
	case LogLevel_Executing_Debug:
		log.Log.Debug(info, fields...)
	case LogLevel_Info:
		log.Log.Info(info, fields...)
	case LogLevel_CheckPoint:
		log.Log.Warn(info, fields...)
	case LogLevel_Error:
		log.Log.Error(info, fields...)
	case LogLevel_Panic:
		log.Log.Panic(info, fields...)
	default:
		log.Log.Panic(info, fields...)
	}

	return logid
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000"))
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:  "T",
		LevelKey: "L",
		NameKey:  "N",
		//CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		//EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
func NewLogger(logfile, svcname, nodename string, nodeid int) *zap.Logger {
	jsonEnc := zapcore.NewJSONEncoder(NewEncoderConfig())
	// set log level
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(levels[Conf.Level])

	allCore := []zapcore.Core{}
	if Conf.Console.SystemOut {
		allCore = append(allCore, zapcore.NewCore(jsonEnc, zapcore.AddSync(os.Stdout), atomicLevel))
	}

	if Conf.LocalFile.SaveFile {
		hook := lumberjack.Logger{
			Filename:   logfile,                   // "./logs/spikeProxy1.log", // log path
			MaxSize:    Conf.LocalFile.MaxSize,    // max size of avery logifle,uint M
			MaxBackups: Conf.LocalFile.MaxBackups, // max backups
			MaxAge:     Conf.LocalFile.MaxAge,     // max days for saving file
			Compress:   Conf.LocalFile.Compress,   // copress or not
		}
		allCore = append(allCore, zapcore.NewCore(jsonEnc, zapcore.AddSync(&hook), atomicLevel))
	}

	core := zapcore.NewTee(allCore...)
	// develop mode ,tracing
	caller := zap.AddCaller()
	// linenumber of file
	development := zap.Development()
	// set global filed
	filed := zap.Fields(
		zap.String("ServiceName", svcname),
		zap.String("ClusterName", nodename),
		zap.Int("ClusterId", nodeid),
		zap.String("SystemVersion", version.MonacoVersion))
	//
	return zap.New(core, caller, development, filed)
}
