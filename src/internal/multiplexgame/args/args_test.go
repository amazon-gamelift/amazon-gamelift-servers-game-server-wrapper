/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package args

import (
	"testing"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	pkgConf "github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/game"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Works(t *testing.T) {
	type Scenario struct {
		Name          string
		Config        *Config
		GameStart     *game.StartArgs
		ExpectedArgs  []string
		ExpectedError error
	}

	ip := "10.0.0.1"

	scenarios := []Scenario{
		{
			Name: "basic",
			Config: &Config{
				BuildDetail: config.BuildDetail{
					DefaultArgs: []pkgConf.CliArg{
						{
							Name:     "+map",
							Value:    "blood_gulch",
							Position: 100,
						},
						{
							Name:     "+environment",
							Value:    "prod",
							Position: 101,
						},
						{
							Name:     "+ip",
							Value:    "{{.IpAddress}}",
							Position: 2,
						},
						{
							Name:     "--port",
							Value:    "{{.GamePort}}",
							Position: 3,
						},
						{
							Name:     "-gameProperties",
							Value:    "{{.GameProperties}}",
							Position: 4,
						},
					},
				},
			},
			GameStart: &game.StartArgs{
				HostingStart: &events.HostingStart{
					IpAddress: ip,
					GamePort:  123,
					CliArgs: []pkgConf.CliArg{
						{
							Name:     "-first",
							Position: 1,
						},
						{
							Name:  "+map",
							Value: "de_dust2",
						},
					},
					GameProperties: "{\"meta1\":\"alpha\",\"meta2\":\"beta\",\"meta3\":\"charlie\"}",
				},
			},
			ExpectedArgs: []string{
				"-first",
				"+ip",
				ip,
				"--port",
				"123",
				"-gameProperties",
				"{\"meta1\":\"alpha\",\"meta2\":\"beta\",\"meta3\":\"charlie\"}",
				"+map",
				"de_dust2",
				"+environment",
				"prod",
			},
			ExpectedError: nil,
		},
		{
			Name: "position_clash",
			Config: &Config{
				BuildDetail: config.BuildDetail{
					DefaultArgs: []pkgConf.CliArg{
						{
							Name:     "+map",
							Value:    "blood_gulch",
							Position: 100,
						},
						{
							Name:     "+environment",
							Value:    "prod",
							Position: 101,
						},
						{
							Name:     "+ip",
							Value:    "{{.IpAddress}}",
							Position: 2,
						},
						{
							Name:     "--port",
							Value:    "{{.GamePort}}",
							Position: 3,
						},
					},
				},
			},
			GameStart: &game.StartArgs{
				HostingStart: &events.HostingStart{
					IpAddress: ip,
					GamePort:  123,
					CliArgs: []pkgConf.CliArg{
						{
							Name:     "-first",
							Position: 2,
						},
						{
							Name:  "+map",
							Value: "de_dust2",
						},
					},
				},
			},
			ExpectedArgs:  nil,
			ExpectedError: errors.New(""),
		},
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			norm := &normaliser{
				cfg: s.Config,
			}

			err := norm.Init()
			assert.NoError(t, err)

			srv := &generator{
				normaliser: norm,
			}

			args, err := srv.Get(s.GameStart)
			if s.ExpectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}

			assert.Equal(t, len(args), len(s.ExpectedArgs))

			for i := 0; i < len(args); i++ {
				exp := s.ExpectedArgs[i]
				act := args[i]

				assert.Equal(t, exp, act)
			}
		})
	}

}
