#!/usr/bin/env bash
set -euo pipefail

root_path="${SUPERSET_ROOT_PATH:?SUPERSET_ROOT_PATH is required}"
workspace_path="$(pwd -P)"
state_dir=".superset"
pid_file="$state_dir/consent-service.pid"
registry_dir="$root_path/.superset"
registry_file="$registry_dir/port-allocations"
lock_dir="$registry_dir/port-allocations.lock"

if [[ -f "$pid_file" ]]; then
  pid="$(<"$pid_file")"
  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
  fi
  rm -f "$pid_file"
fi

if [[ -f "$registry_file" ]]; then
  until mkdir "$lock_dir" 2>/dev/null; do sleep 0.1; done
  trap 'rmdir "$lock_dir"' EXIT
  sed -i.bak "\\|^$workspace_path[[:space:]]|d" "$registry_file"
  rm -f "$registry_file.bak"
fi
