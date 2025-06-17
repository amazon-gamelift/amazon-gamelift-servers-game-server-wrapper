/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type msg struct {
	ctx  context.Context
	msg  string
	args []any
}

type mockLogger struct {
	debug        []msg
	debugContext []msg
	info         []msg
	infoContext  []msg
	warn         []msg
	warnContext  []msg
	error        []msg
	errorContext []msg
}

func newMsg(m string, args []any) msg {
	return msg{
		msg:  m,
		args: args,
	}
}

func newMockLogger() *mockLogger {
	m := &mockLogger{
		debug:        make([]msg, 0),
		debugContext: make([]msg, 0),
		info:         make([]msg, 0),
		infoContext:  make([]msg, 0),
		warn:         make([]msg, 0),
		warnContext:  make([]msg, 0),
		error:        make([]msg, 0),
		errorContext: make([]msg, 0),
	}
	return m
}

func (m *mockLogger) Debug(msg string, args ...any) {
	m.debug = append(m.debug, newMsg(msg, args))
}

func (m *mockLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	m.debugContext = append(m.debugContext, newMsg(msg, args))
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.info = append(m.info, newMsg(msg, args))
}

func (m *mockLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	m.infoContext = append(m.infoContext, newMsg(msg, args))
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.warn = append(m.warn, newMsg(msg, args))
}

func (m *mockLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	m.warnContext = append(m.warnContext, newMsg(msg, args))
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.error = append(m.error, newMsg(msg, args))
}

func (m *mockLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	m.errorContext = append(m.errorContext, newMsg(msg, args))
}

func (m *mockLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	switch level {
	case slog.LevelDebug:
		m.DebugContext(ctx, msg, args)
	case slog.LevelInfo:
		m.InfoContext(ctx, msg, args)
	case slog.LevelWarn:
		m.WarnContext(ctx, msg, args)
	case slog.LevelError:
		m.ErrorContext(ctx, msg, args)
	}
}

func (m *mockLogger) With(args ...any) Logger {
	return newMockLogger()
}

func Test_BufferedLogger_WritesAsExpected(t *testing.T) {
	t.Skip()
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			logger := newMockLogger()

			dir, err := os.MkdirTemp("", "")
			assert.NoError(t, err)

			tmp, err := os.CreateTemp("", "")
			assert.NoError(t, err)

			genLine := func(i int) string {
				return fmt.Sprintf("test log line %d", i)
			}

			for i := 0; i < 100; i++ {
				fmt.Fprintln(tmp, genLine(i))
			}

			bl, err := NewBufferedLogger(ctx, logger, "test", "")
			assert.NoError(t, err)
			bl = bl.WithLevel(level)

			cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("while read -r line; do echo $line; sleep 0.01; done < %s", tmp.Name()))
			cmd.Stderr = bl
			cmd.Stdout = bl

			assert.NoError(t, cmd.Start())
			assert.NoError(t, cmd.Wait())

			err = bl.Close()
			if file := bl.File(); file != nil {
				assert.NoError(t, err)
				rdr := bufio.NewScanner(file)

				current := 0
				for rdr.Scan() {
					l := rdr.Text()
					expected := genLine(current)
					if l != expected {
						t.Errorf("expected %s but got %s", expected, l)
					}
					current++
				}
				assert.NoError(t, os.Remove(file.Name()))
			}

			assert.NoError(t, err)

			if len(logger.debug) > 0 {
				t.Errorf("non contextual debug log sent")
			}

			if len(logger.info) > 0 {
				t.Errorf("non contextual info log sent")
			}

			if len(logger.warn) > 0 {
				t.Errorf("non contextual warn log sent")
			}

			if len(logger.error) > 0 {
				t.Errorf("non contextual error log sent")
			}

			var msgArr []msg
			switch level {
			case slog.LevelDebug:
				msgArr = logger.debugContext
			case slog.LevelInfo:
				msgArr = logger.infoContext
			case slog.LevelWarn:
				msgArr = logger.warnContext
			case slog.LevelError:
				msgArr = logger.errorContext
			}

			if len(logger.debugContext) > 0 && level != slog.LevelDebug {
				t.Errorf("debug log unexpectedly sent")
			}

			if len(logger.infoContext) > 0 && level != slog.LevelInfo {
				t.Errorf("info log unexpectedly sent")
			}

			if len(logger.warnContext) > 0 && level != slog.LevelWarn {
				t.Errorf("warn log unexpectedly sent")
			}

			if len(logger.errorContext) > 0 && level != slog.LevelError {
				t.Errorf("error log unexpectedly sent")
			}

			current := 0
			for _, l := range msgArr {
				expected := genLine(current)
				if l.msg != expected {
					t.Errorf("expected %s but got %s", expected, l)
				}
				current++
			}

			assert.NoError(t, os.Remove(dir))
			assert.NoError(t, os.Remove(tmp.Name()))
		})
	}
}
