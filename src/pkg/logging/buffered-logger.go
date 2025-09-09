/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"bufio"
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/pkg/errors"
)

type BufferedLogger struct {
	buf     bytes.Buffer
	scanner *bufio.Scanner
	logger  Logger
	file    *os.File
	name    string

	level   slog.Level
	mutex   sync.Mutex
	closed  bool
	ctx     context.Context
	onClose func(context.Context) error
}

func NewBufferedLogger(ctx context.Context, logger Logger, name, logDirectory string) (*BufferedLogger, error) {
	bufferedLogger := &BufferedLogger{
		logger: logger,
		name:   name,
		level:  slog.LevelInfo,
		ctx:    context.WithValue(ctx, string(constants.ContextKeySource), name),
	}

	bufferedLogger.scanner = bufio.NewScanner(&bufferedLogger.buf)

	if len(logDirectory) != 0 {
		path := filepath.Join(logDirectory, name)

		lf, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create log file %s", path)
		}

		_, err = lf.Stat()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to stat log file %s", path)
		}

		bufferedLogger.file = lf
	}

	return bufferedLogger, nil
}

func (bufferedLogger *BufferedLogger) WithLevel(lvl slog.Level) *BufferedLogger {
	bufferedLogger.level = lvl
	return bufferedLogger
}

func (bufferedLogger *BufferedLogger) SetOnClosed(f func(ctx context.Context) error) {
	bufferedLogger.onClose = f
}

func (bufferedLogger *BufferedLogger) File() *os.File {
	return bufferedLogger.file
}

func (bufferedLogger *BufferedLogger) Write(p []byte) (int, error) {
	if bufferedLogger.file != nil {
		_, err := bufferedLogger.file.Write(p)
		if err != nil {
			return 0, err
		}
	}

	n, e := bufferedLogger.buf.Write(p)
	if bufferedLogger.scanner.Scan() {
		line := bufferedLogger.scanner.Text()
		bufferedLogger.logger.Log(bufferedLogger.ctx, bufferedLogger.level, line)
	}

	return n, e
}

func (bufferedLogger *BufferedLogger) Close() error {
	if bufferedLogger == nil {
		return nil
	}

	bufferedLogger.mutex.Lock()
	defer bufferedLogger.mutex.Unlock()
	if bufferedLogger.closed {
		return nil
	}

	if bufferedLogger.onClose != nil {
		bufferedLogger.onClose(bufferedLogger.ctx)
	}

	if bufferedLogger.file != nil {
		if err := bufferedLogger.file.Close(); err != nil {
			return err
		}
	}

	return nil
}
