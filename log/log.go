package log

import (
	"context"
	"dash/config"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 单例模式
var (
	exportUseLogger        *zap.Logger
	exportUseSugaredLogger *zap.SugaredLogger
)

// NewLogger 根据配置文件返回一个日志记录器
func NewLogger(conf *config.Config) *zap.Logger {
	if _, err := os.Stat(conf.Dash.LogDir); err != nil {
		if os.IsNotExist(err) && !config.LogToConsole() {
			if err := os.MkdirAll(conf.Dash.LogDir, os.ModePerm); err != nil {
				panic("mkdir failed![%v]")
			}
		}
	}
	var core zapcore.Core
	if config.LogToConsole() {
		core = zapcore.NewCore(getDevEncoder(), os.Stdout, getLogLevel(conf.Log.Levels.App))
	} else {
		core = zapcore.NewCore(getProdEncoder(), getWriter(conf), zap.DebugLevel)
	}
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel))
	exportUseLogger = logger.WithOptions(zap.AddCallerSkip(1))
	exportUseSugaredLogger = exportUseLogger.Sugar()
	return logger
}

// getWriter 自定义Writer,分割日志
func getWriter(conf *config.Config) zapcore.WriteSyncer {
	rotatingLogger := &lumberjack.Logger{
		Filename: filepath.Join(conf.Dash.LogDir, conf.Log.FileName),
		MaxSize:  conf.Log.MaxSize,
		MaxAge:   conf.Log.MaxAge,
		Compress: conf.Log.Compress,
	}
	return zapcore.AddSync(rotatingLogger)
}

// getProdEncoder 自定义日志编码器
func getProdEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getDevEncoder 自定义日志编码器
func getDevEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		panic("log level error")
	}
}

func Debugf(template string, args ...interface{}) {
	exportUseSugaredLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	exportUseSugaredLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	exportUseSugaredLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	exportUseSugaredLogger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	exportUseSugaredLogger.Fatalf(template, args...)
}

func Debug(msg string, fields ...zap.Field) {
	exportUseLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	exportUseLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	exportUseLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	exportUseLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	exportUseLogger.Fatal(msg, fields...)
}

func CtxDebugf(ctx context.Context, template string, args ...interface{}) {
	exportUseSugaredLogger.Debugf(template, args...)
}

func CtxInfof(ctx context.Context, template string, args ...interface{}) {
	exportUseSugaredLogger.Infof(template, args...)
}

func CtxWarnf(ctx context.Context, template string, args ...interface{}) {
	exportUseSugaredLogger.Warnf(template, args...)
}

func CtxErrorf(ctx context.Context, template string, args ...interface{}) {
	exportUseSugaredLogger.Errorf(template, args...)
}

func CtxFatalf(ctx context.Context, template string, args ...interface{}) {
	exportUseSugaredLogger.Fatalf(template, args...)
}

func CtxDebug(ctx context.Context, msg string, fields ...zap.Field) {
	exportUseLogger.Debug(msg, fields...)
}

func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	exportUseLogger.Info(msg, fields...)
}

func CtxWarn(ctx context.Context, msg string, fields ...zap.Field) {
	exportUseLogger.Warn(msg, fields...)
}

func CtxError(ctx context.Context, msg string, fields ...zap.Field) {
	exportUseLogger.Error(msg, fields...)
}

func CtxFatal(ctx context.Context, msg string, fields ...zap.Field) {
	exportUseLogger.Fatal(msg, fields...)
}

func Sync() {
	_ = exportUseLogger.Sync()
	_ = exportUseSugaredLogger.Sync()
}
