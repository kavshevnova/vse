package main

// Задача: обёртка над FileLogger с потокобезопасным последовательным логированием.

import (
	"os"
	"sync"
)

type Logger interface {
	Log(message string) error
	Close() error
}

type FileLogger struct {
	file *os.File
}

func NewFileLogger(fileName string) (*FileLogger, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return &FileLogger{f}, nil
}

func (f *FileLogger) Log(message string) error {
	_, err := f.file.WriteString(message + "\n")
	return err
}

func (f *FileLogger) Close() error {
	return f.file.Close()
}

// ======= КОД ВЫШЕ НЕЛЬЗЯ МЕНЯТЬ =========

// Минусы FileLogger:
// 1. Не потокобезопасен — одновременные вызовы Log могут перемежаться
// 2. Нет буферизации — каждый Log делает syscall
// 3. Нет обработки ошибок Close при defer
// 4. Нет возможности сменить выход без изменения кода

// SequentialLogger — потокобезопасная обёртка: гарантирует последовательное логирование.
type SequentialLogger struct {
	wrppedLogger Logger
	mu           sync.Mutex
}

func NewSequentialLogger(wrppedLogger Logger) SequentialLogger {
	return SequentialLogger{wrppedLogger: wrppedLogger}
}

func (sl SequentialLogger) Log(message string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.wrppedLogger.Log(message)
}

func (sl SequentialLogger) Close() error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.wrppedLogger.Close()
}

func main() {
	fl, err := NewFileLogger("/tmp/test.log")
	if err != nil {
		panic(err)
	}
	logger := NewSequentialLogger(fl)
	logger.Log("hello")
	logger.Log("world")
	logger.Close()
}
