package log

import (
	"log/slog"
	"strings"
)

type Logger struct {
	prefixs []string
}

func (l *Logger) SetPrefix(prefix ...string) {
	l.prefixs = prefix
}
func (l *Logger) AddPrefix(prefix ...string) {
	l.prefixs = append(l.prefixs, prefix...)
}

func (l Logger) getPrefix() string {
	if len(l.prefixs) == 0 {
		return ""
	}
	return strings.Join(l.prefixs, " ") + " "
}

func (l Logger) Debug(msg string, args ...any) {
	slog.Debug(l.getPrefix()+msg, args...)
}
func (l Logger) Info(msg string, args ...any) {
	slog.Info(l.getPrefix()+msg, args...)
}

func (l Logger) Warn(msg string, args ...any) {
	slog.Warn(l.getPrefix()+msg, args...)
}
func (l Logger) Error(msg string, args ...any) {
	slog.Error(l.getPrefix()+msg, args...)
}

var DefaultLog Logger

func WithPrefix(prefix ...string) *Logger {
	return &Logger{
		prefixs: prefix,
	}
}

func Debug(msg string, args ...any) {
	DefaultLog.Debug(msg, args...)
}
func Info(msg string, args ...any) {
	DefaultLog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	DefaultLog.Warn(msg, args...)
}
func Error(msg string, args ...any) {
	DefaultLog.Error(msg, args...)
}

func init() {
	DefaultLog = Logger{}
}
