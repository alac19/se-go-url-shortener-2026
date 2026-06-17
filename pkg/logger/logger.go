// Package logger provides structured log initialization based on slog.
package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Init 初始化全局 slog 日志器。
func Init(level string, filePath string) error {
	// 解析日志级别
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var writers []io.Writer
	writers = append(writers, os.Stderr) // 控制台

	if filePath != "" {
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			return err
		}

		writers = append(writers, file)
	}

	multiWriter := io.MultiWriter(writers...)
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	return nil
}
