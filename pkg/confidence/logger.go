package confidence

// Logger is the logging interface used by the Confidence SDK.
// A custom implementation can be provided via ConfidenceBuilder.SetLogger().
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// noopLogger silently discards all log messages.
type noopLogger struct{}

func (l *noopLogger) Debug(_ string, _ ...any) {}
func (l *noopLogger) Info(_ string, _ ...any)  {}
func (l *noopLogger) Warn(_ string, _ ...any)  {}
