#!/bin/bash
set -euo pipefail

DOCKER_USERNAME="${1:-}"
IMAGE_NAME="${2:-}"

if [[ -z "$DOCKER_USERNAME" || -z "$IMAGE_NAME" ]]; then
    echo "Usage: $0 <docker-username> <image-name>"
    exit 1
fi

GIT_SHA=$(git rev-parse --short HEAD)
IMAGE_BASE="${DOCKER_USERNAME}/${IMAGE_NAME}"

echo "Building ${IMAGE_BASE}:${GIT_SHA}..."
docker build -t "${IMAGE_BASE}:${GIT_SHA}" -t "${IMAGE_BASE}:latest" .

echo "Pushing images..."
docker push "${IMAGE_BASE}:${GIT_SHA}"
docker push "${IMAGE_BASE}:latest"

echo "Done: ${IMAGE_BASE}:${GIT_SHA}, ${IMAGE_BASE}:latest"
