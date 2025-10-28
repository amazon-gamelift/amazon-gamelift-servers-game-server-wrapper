#!/bin/bash

# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

GAME_EXECUTABLE="/local/game/amazon-gamelift-servers-game-server-wrapper"

echo "Starting game server: $GAME_EXECUTABLE $GAME_ARGUMENTS"
if [ -n "$GAME_ARGUMENTS" ]; then
    exec "$GAME_EXECUTABLE" $GAME_ARGUMENTS
else
    exec "$GAME_EXECUTABLE"
fi
