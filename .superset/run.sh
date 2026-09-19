#!/usr/bin/env bash
set -euo pipefail

state_dir=".superset"
env_file="$state_dir/workspace.env"
pid_file="$state_dir/consent-service.pid"

if [[ ! -f "$env_file" ]]; then
  echo "Missing $env_file. Re-run workspace setup." >&2
  exit 1
fi

# shellcheck source=/dev/null
source "$env_file"
mkdir -p "$state_dir"
echo "$$" > "$pid_file"
trap 'rm -f "$pid_file"' EXIT INT TERM

cd consent-service
go build -o ../.superset/consent-service ./cmd/server
exec ../.superset/consent-service
