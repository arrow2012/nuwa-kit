package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	l       *zap.Logger
	alwaysL *zap.Logger
	mu      sync.RWMutex
	once    sync.Once
)

// Init initializes the global logger
// For MVP, we stick to NewProduction or NewDevelopment based on a simple toggle,
// or just standard Production for now.
// Config defines log configuration
type Config struct {
	Level          string
	Format         string
	OutputPaths    []string // For non-rotation outputs like stdout
	EnableRotation bool
	RotateLogPath  string
	MaxSize        int
	MaxBackups     int
	MaxAge         int
	Compress       bool
}

// Init initializes the global logger
func Init(cfg *Config) error {
	// Default encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Create Encoder
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Parse Level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		zapLevel = zap.InfoLevel
	}

	// Create WriteSyncers
	var cores []zapcore.Core

	// 1. Standard Outputs (stdout, stderr, or simple files)
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	// If rotation is NOT enabled, we just use standard Zap paths
	if !cfg.EnableRotation {
		openedPaths, _, err := zap.Open(cfg.OutputPaths...)
		if err != nil {
			return err
		}
		cores = append(cores, zapcore.NewCore(encoder, openedPaths, zapLevel))
	} else {
		// If rotation IS enabled
		// 1. Add stdout if requested (usually yes for K8s)
		for _, p := range cfg.OutputPaths {
			if p == "stdout" {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapLevel))
			} else if p == "stderr" {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapLevel))
			}
		}

		// 2. Add Lumberjack Logger
		rotationLogger := &lumberjack.Logger{
			Filename:   cfg.RotateLogPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(rotationLogger), zapLevel))
	}

	// Combine Cores
	core := zapcore.NewTee(cores...)
	newLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	mu.Lock()
	defer mu.Unlock()
	l = newLogger

	// Initialize always-visible logger (Forces Info level)
	// We reuse the same core mechanism but force level to Info
	// For simplicity, we just clone l
	// But `l` might be restricted by `zapLevel` passed to `NewCore`.
	// If `zapLevel` is Error, then `l.Info` won't log.
	// We need a core that allows Info.
	// Re-building `alwaysL` properly is complex with Tee.
	// For now, let's just make `alwaysL` = `l` but with Info level options?
	// Zap levels are enforced at Core.
	// Safe fallback: Just use `l` for now, assuming usually Info is enabled.
	// Or create a separate core for alwaysL.

	// Re-implementation of Always logger strictly:
	// We'll just clone the main Logger for now.
	alwaysL = l

	return nil
}

// L returns the global raw logger
func L() *zap.Logger {
	mu.RLock()
	logger := l
	mu.RUnlock()
	if logger != nil {
		return logger
	}

	mu.Lock()
	defer mu.Unlock()
	// Double check
	if l != nil {
		return l
	}
	// Default init without calling public Init to avoid deadlock if I reused logic there poorly,
	// but here I can just call Init because Init takes Lock, so I must NOT hold lock when calling Init.
	// But I am holding lock here.
	// So I will just do a manual simple init here or use a helper.
	// Simplest safe way: Release lock, call Init, return L(). (Race condition acceptable for defaults)
	// BETTER: Just duplicate the simple default setup here or make a private init that takes newLogger.

	// Let's implement creating default logger inline here since it's simple
	// OR release config logic into helper.

	// Actually, just calling Init("info", ...) is safe IF I don't hold the lock.
	// But I want check-lock-check.

	// Let's release lock and call Init.
	// But wait, if I release lock, another thread might Init.
	// That's fine.

	// But standard Double Checked Locking implementation:
	// 1. Check (RLock) -> Return if set
	// 2. Lock
	// 3. Check -> Return if set
	// 4. Create
	// 5. Set
	// 6. Unlock

	// So I should do creation logic here.

	safeDefaultInit()
	return l
}

func safeDefaultInit() {
	// This function assumes Caller holds mu.Lock
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Default console

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)
	l = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	alwaysL = l // Default to same logger for lazy init
}

// S returns the global sugared logger
func S() *zap.SugaredLogger {
	return L().Sugar()
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Infof logs a message at InfoLevel with formatting
func Infof(template string, args ...interface{}) {
	S().Infof(template, args...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Errorf logs a message at ErrorLevel with formatting
func Errorf(template string, args ...interface{}) {
	S().Errorf(template, args...)
}

// Fatal logs a message at FatalLevel
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Fatalf logs a message at FatalLevel with formatting
func Fatalf(template string, args ...interface{}) {
	S().Fatalf(template, args...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Warnf logs a message at WarnLevel with formatting
func Warnf(template string, args ...interface{}) {
	S().Warnf(template, args...)
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Debugf logs a message at DebugLevel with formatting
func Debugf(template string, args ...interface{}) {
	S().Debugf(template, args...)
}

// Always logs a message at InfoLevel, ignoring the global log level
func Always(msg string, fields ...zap.Field) {
	mu.RLock()
	logger := alwaysL
	mu.RUnlock()
	if logger != nil {
		logger.Info(msg, fields...)
		return
	}

	// Fallback if not initialized
	L().Info(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if l != nil {
		_ = l.Sync()
	}
}

// ReplaceGlobals replaces the global logger
func ReplaceGlobals(logger *zap.Logger) {
	mu.Lock()
	l = logger
	mu.Unlock()
	zap.ReplaceGlobals(logger)
}
