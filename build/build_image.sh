#!/usr/bin/env bash

./build/build.sh

buildah from --name puzzlerightserver-working-container scratch
buildah copy puzzlerightserver-working-container $HOME/go/bin/puzzlerightserver /bin/puzzlerightserver
buildah config --env SERVICE_PORT=50051 puzzlerightserver-working-container
buildah config --port 50051 puzzlerightserver-working-container
buildah config --entrypoint '["/bin/puzzlerightserver"]' puzzlerightserver-working-container
buildah commit puzzlerightserver-working-container puzzlerightserver
buildah rm puzzlerightserver-working-container

buildah push puzzlerightserver docker-daemon:puzzlerightserver:latest
