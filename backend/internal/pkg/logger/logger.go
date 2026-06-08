package logger

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileConfig struct {
	BaseDir    string
	Purpose    string
	MaxSizeMB  int
	MaxBackups int
}

func DefaultFileConfig() FileConfig {
	return FileConfig{
		BaseDir:    "logs",
		Purpose:    "api",
		MaxSizeMB:  10,
		MaxBackups: 30,
	}
}

type dailyRotatingWriter struct {
	baseDir     string
	purpose     string
	maxSizeMB   int
	maxBackups  int
	mu          sync.Mutex
	currentDate string
	current     *lumberjack.Logger
}

func (w *dailyRotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if w.currentDate != today {
		if w.current != nil {
			_ = w.current.Close()
		}
		dir := filepath.Join(w.baseDir, today)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, err
		}
		w.current = &lumberjack.Logger{
			Filename:   filepath.Join(dir, w.purpose+".log"),
			MaxSize:    w.maxSizeMB,
			MaxBackups: w.maxBackups,
			MaxAge:     0,
			LocalTime:  true,
		}
		w.currentDate = today
	}
	return w.current.Write(p)
}

func New(env string) (*zap.Logger, error) {
	cfg := DefaultFileConfig()
	return newImpl(env, &cfg)
}

func NewWithConfig(env string, fileCfg FileConfig) (*zap.Logger, error) {
	return newImpl(env, &fileCfg)
}

func NewNoFile(env string) (*zap.Logger, error) {
	return newImpl(env, nil)
}

func newImpl(env string, fileCfg *FileConfig) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	enc := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	if env == "production" {
		enc = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	}
	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stderr), cfg.Level)
	if fileCfg != nil && fileCfg.BaseDir != "" && fileCfg.Purpose != "" {
		maxSize := fileCfg.MaxSizeMB
		if maxSize <= 0 {
			maxSize = 10
		}
		daily := &dailyRotatingWriter{
			baseDir:    fileCfg.BaseDir,
			purpose:    fileCfg.Purpose,
			maxSizeMB:  maxSize,
			maxBackups: fileCfg.MaxBackups,
		}
		core = zapcore.NewTee(core, zapcore.NewCore(enc, zapcore.AddSync(daily), cfg.Level))
	}
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
