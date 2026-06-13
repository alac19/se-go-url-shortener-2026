package logger

import (
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		filepath string
		wantErr  bool
	}{
		{"debug 级别日志", "debug", "logs/test.log", false},
		{"info 级别日志", "info", "logs/test.log", false},
		{"warn 级别日志", "warn", "logs/test.log", false},
		{"error 级别日志", "error", "logs/test.log", false},
		{"Info 级别日志", "info", "logs/test.log", false},
		{"INFO 级别日志", "info", "logs/test.log", false},
		{"默认级别日志", "", "logs/test.log", false},
		{"info 级别日志, 无文件输出", "info", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Init(test.level, test.filepath)

			if (err != nil) != test.wantErr {
				t.Errorf("error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}
