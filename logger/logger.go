package logger

import (
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(path string) {
	fileLogger := &lumberjack.Logger{
		Filename:	path,
		MaxSize:	10,
		MaxBackups:	3,
		MaxAge:		28,
		Compress: 	true,
	}
	logger := slog.New(slog.NewJSONHandler(fileLogger, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

}