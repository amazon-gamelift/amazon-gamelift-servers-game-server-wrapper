/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datadog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the datadog agent configuration
// Using map[string]interface{} to preserve all configuration fields
type Config map[string]interface{}

// Service handles datadog configuration updates
type Service struct {
	configPath   string
	tagTemplates map[string]string
	reloadCmd    []string
	testMode     bool
	logger       *slog.Logger
}

// New creates a new datadog configuration service
func New(configPath string, tagTemplates map[string]string, logger *slog.Logger) *Service {
	return &Service{
		configPath:   configPath,
		tagTemplates: tagTemplates,
		reloadCmd:    []string{"sudo", "systemctl", "restart", "datadog-agent"},
		testMode:     false,
		logger:       logger,
	}
}

// NewForTesting creates a new datadog configuration service in test mode
func NewForTesting(configPath string, tagTemplates map[string]string, logger *slog.Logger) *Service {
	return &Service{
		configPath:   configPath,
		tagTemplates: tagTemplates,
		reloadCmd:    []string{"sudo", "systemctl", "restart", "datadog-agent"},
		testMode:     true,
		logger:       logger,
	}
}

// UpdateTags processes tag templates and updates the datadog configuration
// This method replaces any existing templated tags with new ones
func (s *Service) UpdateTags(ctx context.Context, data interface{}) error {
	s.logger.DebugContext(ctx, "Updating datadog configuration with templated tags", "data", data)

	// Check if we have write permissions
	if err := s.checkWritePermissions(); err != nil {
		return errors.Wrap(err, "insufficient permissions to modify datadog configuration")
	}

	// Read current configuration
	config, err := s.readConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read datadog configuration")
	}

	// Initialize config if it's nil
	if config == nil {
		config = make(Config)
	}

	// Remove existing templated tags (tags that match our template patterns)
	if tags, ok := config["tags"].([]interface{}); ok {
		config["tags"] = s.removeTemplatedTagsFromInterface(tags)
	} else {
		// If tags field doesn't exist or is not a slice, initialize it
		config["tags"] = []interface{}{}
	}

	// Process each tag template
	for tagName, templateStr := range s.tagTemplates {
		renderedValue, err := s.renderTemplate(templateStr, data)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to render tag template", "tag", tagName, "template", templateStr, "error", err)
			continue
		}

		tag := fmt.Sprintf("%s:%s", tagName, renderedValue)
		if tags, ok := config["tags"].([]interface{}); ok {
			config["tags"] = append(tags, tag)
		} else {
			config["tags"] = []interface{}{tag}
		}
		s.logger.InfoContext(ctx, "Added templated tag to datadog configuration", "tag", tag)
	}

	// Write updated configuration
	if err := s.writeConfig(config); err != nil {
		return errors.Wrap(err, "failed to write datadog configuration")
	}

	// Reload datadog agent
	if err := s.reloadAgent(ctx); err != nil {
		return errors.Wrap(err, "failed to reload datadog agent")
	}

	s.logger.InfoContext(ctx, "Successfully updated datadog configuration with templated tags and reloaded agent")
	return nil
}

// readConfig reads the current datadog configuration from disk
func (s *Service) readConfig() (Config, error) {
	// Check if config file exists
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		return Config{
			"tags": []interface{}{},
		}, nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read datadog config file")
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to parse datadog config file")
	}

	return config, nil
}

// writeConfig writes the configuration to disk
func (s *Service) writeConfig(config Config) error {
	// Ensure directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create datadog config directory")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "failed to marshal datadog config")
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write datadog config file")
	}

	return nil
}

// reloadAgent restarts the datadog agent to pick up configuration changes
func (s *Service) reloadAgent(ctx context.Context) error {
	// In test mode, just log what would be executed
	if s.testMode {
		cmdStr := strings.Join(s.reloadCmd, " ")
		s.logger.InfoContext(ctx, "Test mode: Would execute reload command", "command", cmdStr)
		return nil
	}

	// Use configured command to restart the datadog-agent service
	cmd := exec.CommandContext(ctx, s.reloadCmd[0], s.reloadCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to restart datadog agent", "error", err, "output", string(output))
		return errors.Wrap(err, "datadog agent restart failed")
	}

	s.logger.DebugContext(ctx, "Datadog agent restarted successfully", "output", string(output))
	return nil
}

// renderTemplate renders a template string with the provided data
func (s *Service) renderTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("tag").Parse(templateStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return buf.String(), nil
}

// removeTemplatedTagsFromInterface removes all tags that match our template patterns from interface slice
func (s *Service) removeTemplatedTagsFromInterface(tags []interface{}) []interface{} {
	var result []interface{}
	for _, tagInterface := range tags {
		tag, ok := tagInterface.(string)
		if !ok {
			// If it's not a string, keep it as-is
			result = append(result, tagInterface)
			continue
		}

		// Check if this tag matches any of our template patterns
		shouldRemove := false
		for tagName := range s.tagTemplates {
			// Check if tag starts with our template tag name followed by ":"
			if strings.HasPrefix(tag, tagName+":") {
				shouldRemove = true
				break
			}
		}

		if !shouldRemove {
			result = append(result, tag)
		}
	}
	return result
}

// checkWritePermissions verifies that the wrapper has write permissions to the datadog config file
func (s *Service) checkWritePermissions() error {
	// Check if the file exists
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		// If file doesn't exist, check if we can create it in the directory
		dir := filepath.Dir(s.configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "failed to create datadog config directory")
		}
	}

	// Try to open the file for writing
	file, err := os.OpenFile(s.configPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "cannot open datadog config file for writing")
	}
	defer file.Close()

	return nil
}
