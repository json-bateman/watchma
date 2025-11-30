#!/bin/bash

# Exit on error
set -e

DOCKER_USERNAME="jsonbateman"
IMAGE_NAME="watchma"
VERSION=${1:-latest}

echo "Building multi-platform Docker image for $DOCKER_USERNAME/$IMAGE_NAME:$VERSION"

# Create and use a new builder instance (only needed once)
if ! docker buildx inspect multiplatform-builder > /dev/null 2>&1; then
    echo "Creating new buildx builder instance..."
    docker buildx create --name multiplatform-builder --use
    docker buildx inspect --bootstrap
else
    echo "Using existing buildx builder instance..."
    docker buildx use multiplatform-builder
fi

# Build and push multi-platform image
echo "Building and pushing for linux/amd64 and linux/arm64..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t $DOCKER_USERNAME/$IMAGE_NAME:$VERSION \
    -t $DOCKER_USERNAME/$IMAGE_NAME:latest \
    --push \
    .

echo "✅ Successfully built and pushed multi-platform images!"
echo "   - $DOCKER_USERNAME/$IMAGE_NAME:$VERSION"
echo "   - $DOCKER_USERNAME/$IMAGE_NAME:latest"
echo ""
echo "Images support:"
echo "   - linux/amd64 (Intel/AMD x86_64)"
echo "   - linux/arm64 (ARM 64-bit, e.g., Raspberry Pi 4, Apple Silicon)"
