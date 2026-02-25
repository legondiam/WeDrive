package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.Logger
var S *zap.SugaredLogger

func Init() {
	writeSyncer := getLogwriter()

	consoleEncoder := getConsoleEncoder()
	fileEncoder := getFileEncoder()
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel) //控制台
	fileCore := zapcore.NewCore(fileEncoder, writeSyncer, zapcore.InfoLevel)                       //文件
	core := zapcore.NewTee(consoleCore, fileCore)
	Logger = zap.New(core, zap.AddCaller())
	S = Logger.Sugar()
}

// 日志输出
func getLogwriter() zapcore.WriteSyncer {
	lumberjacklogger := &lumberjack.Logger{
		Filename:   "./logs/app.log", // 日志文件路径
		MaxSize:    10,               // 每个日志文件保存的最大尺寸
		MaxBackups: 5,                // 日志文件最多保存多少个备份
		MaxAge:     30,               // 文件最多保存多少天
		Compress:   false,            // 是否压缩
	}
	return zapcore.AddSync(lumberjacklogger)
}

// 日志格式
func getConsoleEncoder() zapcore.Encoder {
	encoderconfig := zap.NewDevelopmentEncoderConfig()
	encoderconfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderconfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderconfig)
}
func getFileEncoder() zapcore.Encoder {
	encoderconfig := zap.NewProductionEncoderConfig()
	encoderconfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderconfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	return zapcore.NewJSONEncoder(encoderconfig)
}
