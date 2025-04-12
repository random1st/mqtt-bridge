package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap/zapcore"
)

func resetLogger() {
	logger = nil
	once = sync.Once{}
}

func TestInitLoggerLevels(t *testing.T) {
	levels := []struct {
		name     string
		levelStr string
		debugOn  bool
		infoOn   bool
		warnOn   bool
		errorOn  bool
	}{
		{"debug", "debug", true, true, true, true},
		{"info", "info", false, true, true, true},
		{"warn", "warn", false, false, true, true},
		{"error", "error", false, false, false, true},
		{"unknown", "random", false, true, true, true},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			resetLogger()
			InitLogger(tc.levelStr)
			log := L()
			if log == nil {
				t.Fatalf("Logger is nil after InitLogger(%q)", tc.levelStr)
			}

			if log.Core().Enabled(zapcore.DebugLevel) != tc.debugOn {
				t.Errorf("Debug level enabled: got %v, want %v", log.Core().Enabled(zapcore.DebugLevel), tc.debugOn)
			}
			if log.Core().Enabled(zapcore.InfoLevel) != tc.infoOn {
				t.Errorf("Info level enabled: got %v, want %v", log.Core().Enabled(zapcore.InfoLevel), tc.infoOn)
			}
			if log.Core().Enabled(zapcore.WarnLevel) != tc.warnOn {
				t.Errorf("Warn level enabled: got %v, want %v", log.Core().Enabled(zapcore.WarnLevel), tc.warnOn)
			}
			if log.Core().Enabled(zapcore.ErrorLevel) != tc.errorOn {
				t.Errorf("Error level enabled: got %v, want %v", log.Core().Enabled(zapcore.ErrorLevel), tc.errorOn)
			}
		})
	}
}

func TestInitLoggerOnce(t *testing.T) {
	resetLogger()
	InitLogger("debug")
	first := L()
	InitLogger("error")
	second := L()

	if first != second {
		t.Error("Expected the same logger instance after second InitLogger call, got different")
	}
	if !first.Core().Enabled(zapcore.DebugLevel) {
		t.Errorf("Expected debug level to remain from first init, but it's not enabled")
	}
}

func TestSync(t *testing.T) {
	resetLogger()
	InitLogger("info")

	Sync()
}
