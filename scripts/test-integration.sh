#!/usr/bin/env bash
set -euo pipefail

# Detect docker command (WSL2 uses docker.exe)
DOCKER="docker"
if ! command -v docker &>/dev/null && command -v docker.exe &>/dev/null; then
  DOCKER="docker.exe"
fi

COMPOSE="$DOCKER compose"
COMPOSE_FILE="docker-compose.integration.yml"
PORTABASE_URL="http://localhost:8887"
EMAIL="admin@test.local"
PASSWORD="TestPass123!"

cleanup() {
  echo "==> Tearing down..."
  $COMPOSE -f "$COMPOSE_FILE" down -v 2>/dev/null || true
}
trap cleanup EXIT

echo "==> Starting Portabase stack..."
$COMPOSE -f "$COMPOSE_FILE" up -d

echo "==> Waiting for Portabase to be healthy..."
for i in $(seq 1 60); do
  if curl -sf "$PORTABASE_URL/api/health" >/dev/null 2>&1; then
    echo "    Portabase is ready (${i}s)"
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "    ERROR: Portabase failed to start after 60s"
    $COMPOSE -f "$COMPOSE_FILE" logs portabase
    exit 1
  fi
  sleep 1
done

# Give it a moment for the default user to be seeded
sleep 3

echo "==> Creating API key..."
API_KEY=$(go run ./cmd/setup-apikey/main.go "$PORTABASE_URL" "$EMAIL" "$PASSWORD")
if [ -z "$API_KEY" ]; then
  echo "    ERROR: Failed to create API key"
  exit 1
fi
echo "    API key created: ${API_KEY:0:20}..."

echo "==> Running unit tests..."
go test ./internal/... -v

echo "==> Running integration tests (httptest fake)..."
INTEGRATION=1 go test ./tests/... -run TestIntegration -v

echo "==> Running e2e tests against real Portabase..."
E2E=1 PORTABASE_BASE_URL="$PORTABASE_URL" PORTABASE_API_TOKEN="$API_KEY" go test ./tests/... -run TestE2E -v -timeout 60s

echo "==> All tests passed!"
