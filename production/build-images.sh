#!/bin/bash

# Multi-platform build and push script for Kuerzen services
# Usage: ./build-and-push.sh <docker-hub-username>

set -e

# Check if Docker Hub username is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <docker-hub-username>"
    echo "Example: $0 myusername"
    exit 1
fi

DOCKER_HUB_USER="$1"
PLATFORMS="linux/amd64,linux/arm64"

echo "🔨 Building multi-platform images for platforms: $PLATFORMS"
echo "📦 Docker Hub user: $DOCKER_HUB_USER"
echo ""

# Create buildx builder if it doesn't exist
if ! docker buildx ls | grep -q "multi-platform"; then
    echo "🔧 Creating buildx builder..."
    docker buildx create --use --name multi-platform --bootstrap
else
    echo "🔧 Using existing buildx builder..."
    docker buildx use multi-platform
fi

# Function to build and push an image
build_and_push() {
    local service_name=$1
    local context_path=$2
    local dockerfile_path=$3
    local image_name="kuerzen-${service_name}"
    local image_tag="${DOCKER_HUB_USER}/${image_name}:latest"

    echo "🏗️  Building $image_name..."
    docker buildx build \
        --platform $PLATFORMS \
        --file $dockerfile_path \
        --tag $image_tag \
        --push \
        $context_path

    echo "✅ Successfully built and pushed $image_tag"
    echo ""
}

# Navigate to production directory
cd "$(dirname "$0")"

echo "📍 Working directory: $(pwd)"
echo ""

# Build and push all services
echo "🚀 Starting multi-platform builds..."
echo ""

# Analytics service
build_and_push "analytics" ".." "../analytics/Dockerfile"

# Shortener service
build_and_push "shortener" ".." "../shortener/Dockerfile"

# Redirector service
build_and_push "redirector" ".." "../redirector/Dockerfile"

# API Gateway service
build_and_push "api-gateway" "../api-gateway" "../api-gateway/Dockerfile"

echo "🎉 All images built and pushed successfully!"
echo ""
echo "📋 Images pushed:"
echo "   - $DOCKER_HUB_USER/kuerzen-analytics:latest"
echo "   - $DOCKER_HUB_USER/kuerzen-shortener:latest"
echo "   - $DOCKER_HUB_USER/kuerzen-redirector:latest"
echo "   - $DOCKER_HUB_USER/kuerzen-api-gateway:latest"
echo ""
