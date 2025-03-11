/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package args

import (
	"bytes"
	"cmp"
	"slices"
	"text/template"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	pkgConfig "github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/pkg/errors"
)

// Generator defines the interface for generating command-line arguments.
// It converts game start arguments into a slice of strings suitable for process execution.
type Generator interface {
	Get(session *game.StartArgs) ([]string, error)
}

// Normaliser defines the interface for normalizing and validating command-line arguments.
type Normaliser interface {
	Init() error
	Normalise(session *game.StartArgs) ([]pkgConfig.CliArg, error)
}

type normaliser struct {
	cfg *Config
}

// Init validates the default arguments configuration.
// It ensures no duplicate arguments and validates argument template syntax.
//
// Returns:
//   - error: If validation fails
func (normaliser *normaliser) Init() error {
	posMap := make(map[int]*pkgConfig.CliArg)
	argMap := make(map[string]*pkgConfig.CliArg)

	for _, a := range normaliser.cfg.DefaultArgs {
		if argMap[a.Name] == nil {
			argMap[a.Name] = &a
		} else {
			return errors.Errorf("duplicate default arg: %s", a.Name)
		}

		if posMap[a.Position] == nil {
			posMap[a.Position] = &a
		} else {
			return errors.Errorf("duplicate default position: %d", a.Position)
		}
	}

	return nil
}

// Normalise processes and combines default arguments with session-specific arguments.
// Parameters:
//   - session: Game session start arguments
//
// Returns:
//   - []pkgConfig.CliArg: Normalized arguments
//   - error: If normalization fails
func (normaliser *normaliser) Normalise(session *game.StartArgs) ([]pkgConfig.CliArg, error) {
	args := make([]pkgConfig.CliArg, 0)

	argMap := make(map[string]*pkgConfig.CliArg)

	for _, a := range normaliser.cfg.DefaultArgs {
		argMap[a.Name] = &a
	}

	for _, a := range session.CliArgs {
		defaultArg := argMap[a.Name]
		if defaultArg != nil {
			if a.Position == 0 {
				a.Position = defaultArg.Position
			}
		}

		argMap[a.Name] = &a
	}

	posMap := make(map[int]*pkgConfig.CliArg)
	for _, v := range argMap {
		existing := posMap[v.Position]
		if existing == nil {
			posMap[v.Position] = v
		} else {
			return nil, errors.Errorf("duplicate default position: %d for '%s' and '%s'", v.Position, v.Name, existing.Name)
		}
		args = append(args, *v)
	}

	slices.SortFunc(args, func(a, b pkgConfig.CliArg) int {
		return cmp.Compare(a.Position, b.Position)
	})

	return args, nil
}

type generator struct {
	normaliser Normaliser
}

type StartArgs struct {
	*game.StartArgs
}

// Get generates the final command-line arguments for a game session.
//
// Parameters:
//   - gsa: Game session start arguments
//
// Returns:
//   - []string: Final processed command-line arguments
//   - error: Any error during generation
func (generator *generator) Get(gsa *game.StartArgs) ([]string, error) {
	session := &StartArgs{
		StartArgs: gsa,
	}

	args, err := generator.normaliser.Normalise(gsa)
	if err != nil {
		return nil, err
	}

	cmdArgs := make([]string, 0)
	for _, arg := range args {
		if len(arg.Value) > 0 {
			val := arg.Value
			t, err := template.New(arg.Name).Parse(val)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse arg template for %s", arg.Name)
			}

			var b bytes.Buffer
			if err := t.Execute(&b, session); err != nil {
				return nil, errors.Wrapf(err, "failed to execute arg template for %s", arg.Name)
			}

			value := b.String()

			cmdArgs = append(cmdArgs, arg.Name, value)
		} else {
			cmdArgs = append(cmdArgs, arg.Name)
		}
	}

	return cmdArgs, nil
}

type Config struct {
	config.BuildDetail
}

// New creates a new argument generator instance.
// It initializes and validates the generator configuration.
//
// Parameters:
//   - cfg: Configuration for the generator
//
// Returns:
//   - Generator: New generator instance
//   - error: Any error during initialization
func New(cfg *Config) (Generator, error) {
	generator := &generator{
		normaliser: &normaliser{
			cfg: cfg,
		},
	}

	if err := generator.normaliser.Init(); err != nil {
		return nil, errors.Wrapf(err, "arg provider validation failed")
	}

	return generator, nil
}
