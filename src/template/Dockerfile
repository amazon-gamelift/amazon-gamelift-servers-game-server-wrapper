# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Add the base image that you want to use over here,
# Make sure to use an image with the same architecture as the
# Instance type you are planning to use on your fleets.
FROM public.ecr.aws/amazonlinux/amazonlinux

# Add your game build directory in the 'GAME_BUILD_DIRECTORY' env variable below.
ENV GAME_BUILD_DIRECTORY="." \
    GAME_EXECUTABLE="amazon-gamelift-servers-game-server-wrapper" \
    HOME_DIR="/local/game"

# install dependencies as necessary
RUN yum install -y shadow-utils

RUN mkdir -p $HOME_DIR
COPY $GAME_BUILD_DIRECTORY/ $HOME_DIR

# Change directory to home
WORKDIR $HOME_DIR

RUN useradd -m gamescale && \
  echo "gamescale ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && \
  chown -R gamescale:gamescale $HOME_DIR

# Add permissions to game build
RUN chmod +x ./$GAME_EXECUTABLE

USER gamescale

# Check directory before starting the container
RUN ls -lhrt .

# Check path before starting the container
RUN echo $PATH

# Start the game build
ENTRYPOINT ["/bin/sh", "-c", "./$GAME_EXECUTABLE"]
