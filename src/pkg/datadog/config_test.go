/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datadog

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_UpdateSessionTag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service
	service := NewForTesting(configPath, map[string]string{"session_name": "{{.GameSessionName}}"}, logger)

	ctx := context.Background()
	sessionName := "test-session-123"

	// Test updating session tag
	hostingStart := &struct {
		GameSessionName string
	}{
		GameSessionName: sessionName,
	}
	err := service.UpdateTags(ctx, hostingStart)
	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Verify the config file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Check that the session tag was added
	configContent := string(data)
	assert.Contains(t, configContent, "session_name:test-session-123")
	assert.Contains(t, configContent, "tags:")
}

func TestService_UpdateSessionTag_ExistingConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create initial config file
	initialConfig := `tags:
  - env:test
  - service:game-server
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service
	service := NewForTesting(configPath, map[string]string{"session_name": "{{.GameSessionName}}"}, logger)

	ctx := context.Background()
	sessionName := "test-session-456"

	// Test updating session tag
	hostingStart := &struct {
		GameSessionName string
	}{
		GameSessionName: sessionName,
	}
	err = service.UpdateTags(ctx, hostingStart)
	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)
	assert.Contains(t, configContent, "session_name:test-session-456")
	assert.Contains(t, configContent, "env:test")
	assert.Contains(t, configContent, "service:game-server")
}

func TestService_UpdateSessionTag_DuplicateTag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service
	service := NewForTesting(configPath, map[string]string{"session_name": "{{.GameSessionName}}"}, logger)

	ctx := context.Background()
	sessionName := "test-session-789"

	// Add the same session tag twice
	hostingStart := &struct {
		GameSessionName string
	}{
		GameSessionName: sessionName,
	}
	err := service.UpdateTags(ctx, hostingStart)

	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	err = service.UpdateTags(ctx, hostingStart)

	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)

	// Count occurrences of the session tag
	count := 0
	lines := strings.Split(configContent, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "- session_name:test-session-789" {
			count++
		}
	}

	// Should only appear once
	assert.Equal(t, 1, count)
}

func TestService_UpdateTags_Templating(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service with multiple tag templates
	tagTemplates := map[string]string{
		"session_name": "{{.GameSessionName}}",
		"fleet_id":     "{{.FleetId}}",
		"provider":     "{{.Provider}}",
		"static_tag":   "static-value",
	}
	service := NewForTesting(configPath, tagTemplates, logger)

	ctx := context.Background()

	// Create a mock HostingStart event
	hostingStart := &struct {
		GameSessionName string
		FleetId         string
		Provider        string
	}{
		GameSessionName: "test-session-templating",
		FleetId:         "fleet-12345",
		Provider:        "gamelift",
	}

	// Test updating tags with templating
	err := service.UpdateTags(ctx, hostingStart)

	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)

	// Check that all templated tags were added
	assert.Contains(t, configContent, "session_name:test-session-templating")
	assert.Contains(t, configContent, "fleet_id:fleet-12345")
	assert.Contains(t, configContent, "provider:gamelift")
	assert.Contains(t, configContent, "static_tag:static-value")
}

func TestService_UpdateTags_ReplacesOldTags(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create initial config file with existing templated tags
	initialConfig := `tags:
  - env:test
  - session_name:old-session
  - fleet_id:old-fleet
  - service:game-server
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service with tag templates
	tagTemplates := map[string]string{
		"session_name": "{{.GameSessionName}}",
		"fleet_id":     "{{.FleetId}}",
	}
	service := NewForTesting(configPath, tagTemplates, logger)

	ctx := context.Background()

	// Create a mock HostingStart event for new session
	hostingStart := &struct {
		GameSessionName string
		FleetId         string
	}{
		GameSessionName: "new-session",
		FleetId:         "new-fleet",
	}

	// Test updating tags - should replace old templated tags
	err = service.UpdateTags(ctx, hostingStart)
	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)

	// Check that old templated tags were removed
	assert.NotContains(t, configContent, "session_name:old-session")
	assert.NotContains(t, configContent, "fleet_id:old-fleet")

	// Check that new templated tags were added
	assert.Contains(t, configContent, "session_name:new-session")
	assert.Contains(t, configContent, "fleet_id:new-fleet")

	// Check that non-templated tags were preserved
	assert.Contains(t, configContent, "env:test")
	assert.Contains(t, configContent, "service:game-server")
}

func TestService_UpdateTags_PreservesExternalTags(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create initial config file with external tags (not managed by wrapper)
	initialConfig := `tags:
  - service:game-server
  - environment:production
  - region:us-west-2
  - custom:my-value
  - external:tag
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service with only session_name template (not the other tags)
	tagTemplates := map[string]string{
		"session_name": "{{.GameSessionName}}",
	}
	service := NewForTesting(configPath, tagTemplates, logger)

	ctx := context.Background()

	// Create a mock HostingStart event
	hostingStart := &struct {
		GameSessionName string
	}{
		GameSessionName: "new-session",
	}

	// Test updating tags - should preserve external tags
	err = service.UpdateTags(ctx, hostingStart)
	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)

	// Check that external tags were preserved
	assert.Contains(t, configContent, "service:game-server")
	assert.Contains(t, configContent, "environment:production")
	assert.Contains(t, configContent, "region:us-west-2")
	assert.Contains(t, configContent, "custom:my-value")
	assert.Contains(t, configContent, "external:tag")

	// Check that new templated tag was added
	assert.Contains(t, configContent, "session_name:new-session")
}

func TestService_CheckWritePermissions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service
	service := NewForTesting(configPath, map[string]string{}, logger)

	// Test with non-existent file (should create directory and succeed)
	err := service.checkWritePermissions()
	assert.NoError(t, err)

	// Verify the directory was created
	_, err = os.Stat(filepath.Dir(configPath))
	assert.NoError(t, err)

	// Test with existing file (should succeed)
	err = os.WriteFile(configPath, []byte("test"), 0644)
	require.NoError(t, err)

	err = service.checkWritePermissions()
	assert.NoError(t, err)
}

func TestService_TestMode_LogsCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a logger that captures output
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create service in test mode
	service := NewForTesting(configPath, map[string]string{"session_name": "{{.GameSessionName}}"}, logger)

	ctx := context.Background()

	// Create a mock HostingStart event
	hostingStart := &struct {
		GameSessionName string
	}{
		GameSessionName: "test-session",
	}

	// Test updating tags - should log the command instead of executing it
	err := service.UpdateTags(ctx, hostingStart)
	require.NoError(t, err)

	// Check that the log contains the test mode message
	logContent := logOutput.String()
	assert.Contains(t, logContent, "Test mode: Would execute reload command")
	assert.Contains(t, logContent, "command=\"sudo systemctl restart datadog-agent\"")
}

func TestService_UpdateTags_PreservesFullConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "datadog.yaml")

	// Create a comprehensive datadog config with many fields
	initialConfig := `api_key: "test-api-key"
site: "datadoghq.com"
logs_enabled: true
apm_config:
  enabled: true
  env: "production"
process_config:
  enabled: true
  process_collection:
    enabled: true
tags:
  - service:game-server
  - environment:production
  - region:us-west-2
  - existing:tag
`
	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create service with tag templates
	tagTemplates := map[string]string{
		"session_name": "{{.GameSessionName}}",
		"fleet_id":     "{{.FleetId}}",
	}
	service := NewForTesting(configPath, tagTemplates, logger)

	ctx := context.Background()

	// Create a mock HostingStart event
	hostingStart := &struct {
		GameSessionName string
		FleetId         string
	}{
		GameSessionName: "test-session",
		FleetId:         "test-fleet",
	}

	// Test updating tags - should preserve all other config
	err = service.UpdateTags(ctx, hostingStart)
	// In test mode, the reload command is mocked, so no errors should occur
	require.NoError(t, err)

	// Read and verify the config content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configContent := string(data)

	// Check that all original config fields are preserved
	assert.Contains(t, configContent, "api_key: test-api-key")
	assert.Contains(t, configContent, "site: datadoghq.com")
	assert.Contains(t, configContent, "logs_enabled: true")
	assert.Contains(t, configContent, "apm_config:")
	assert.Contains(t, configContent, "enabled: true")
	assert.Contains(t, configContent, "env: production")
	assert.Contains(t, configContent, "process_config:")
	assert.Contains(t, configContent, "process_collection:")

	// Check that existing tags are preserved
	assert.Contains(t, configContent, "service:game-server")
	assert.Contains(t, configContent, "environment:production")
	assert.Contains(t, configContent, "region:us-west-2")
	assert.Contains(t, configContent, "existing:tag")

	// Check that new templated tags were added
	assert.Contains(t, configContent, "session_name:test-session")
	assert.Contains(t, configContent, "fleet_id:test-fleet")
}
