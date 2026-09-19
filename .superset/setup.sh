#!/usr/bin/env bash
set -euo pipefail

root_path="${SUPERSET_ROOT_PATH:?SUPERSET_ROOT_PATH is required}"
workspace_path="$(pwd -P)"
state_dir=".superset"
env_file="$state_dir/workspace.env"
registry_dir="$root_path/.superset"
registry_file="$registry_dir/port-allocations"
lock_dir="$registry_dir/port-allocations.lock"

copy_env_files() {
  while IFS= read -r -d '' relative_path; do
    mkdir -p "$(dirname "$relative_path")"
    cp -p "$root_path/$relative_path" "$relative_path"
  done < <(git -C "$root_path" ls-files -z --others --ignored --exclude-standard -- ':(glob)**/.env*')
}

acquire_lock() {
  until mkdir "$lock_dir" 2>/dev/null; do sleep 0.1; done
}

release_lock() { rmdir "$lock_dir"; }

port_is_available() {
  ! lsof -nP -iTCP:"$1" -sTCP:LISTEN >/dev/null 2>&1
}

allocate_port() {
  mkdir -p "$registry_dir" "$state_dir"
  acquire_lock
  trap release_lock RETURN
  touch "$registry_file"
  sed -i.bak "\\|^$workspace_path[[:space:]]|d" "$registry_file"
  rm -f "$registry_file.bak"

  for port in $(seq 18080 18999); do
    if ! awk -v candidate="$port" '$2 == candidate { found = 1 } END { exit found ? 0 : 1 }' "$registry_file" && port_is_available "$port"; then
      printf '%s\t%s\n' "$workspace_path" "$port" >> "$registry_file"
      printf 'export CONSENT__SERVER__PORT=%q\nexport CONSENT__SERVER__BASE_URL=%q\nexport CONSENT__SERVER__FRONTEND_URL=%q\n' \
        "$port" "http://localhost:$port" "http://localhost:$port" > "$env_file"
      return
    fi
  done

  echo "No free port in 18080-18999" >&2
  return 1
}

copy_env_files
allocate_port

for module in customer appointment notification consent-service core-contracts/notification core-contracts/appointment core-contracts/runtime; do
  (cd "$module" && go mod download)
done
