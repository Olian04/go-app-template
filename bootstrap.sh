#!/usr/bin/env bash

set -euo pipefail

GUM=(go tool gum)
OLD_MODULE="github.com/Olian04/go-app-template"
OLD_CMD="echo"

remote_module_path() {
	local remote top
	top="$(git rev-parse --show-toplevel 2>/dev/null || true)"
	if [[ "${top}" != "$(pwd)" ]]; then
		return 1
	fi
	remote="$(git config --get remote.origin.url 2>/dev/null || true)"
	if [[ -z "${remote}" ]]; then
		return 1
	fi
	remote="${remote%.git}"
	case "${remote}" in
		git@*:*)
			remote="${remote#git@}"
			remote="${remote/:/\/}"
			;;
		ssh://git@*)
			remote="${remote#ssh://git@}"
			;;
		https://*)
			remote="${remote#https://}"
			;;
		http://*)
			remote="${remote#http://}"
			;;
	esac
	printf '%s\n' "${remote}"
}

module_basename() {
	local module_path="$1"
	local name="${module_path##*/}"
	name="${name%.git}"
	printf '%s\n' "${name,,}"
}

latest_go_major_minor() {
	local version
	version="${GO_VERSION:-$(go env GOVERSION)}"
	version="${version#go}"
	if [[ "${version}" =~ ^([0-9]+)\.([0-9]+) ]]; then
		printf '%s.%s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}"
		return 0
	fi
	printf 'unable to parse Go version from %q\n' "${version}" >&2
	return 1
}

replace_all() {
	local from="$1"
	local to="$2"
	python3 - "$from" "$to" <<'PY'
import sys
from pathlib import Path

root = Path.cwd()
old, new = sys.argv[1], sys.argv[2]
skip_dirs = {".git", "dist"}
for path in root.rglob("*"):
    if any(part in skip_dirs for part in path.parts):
        continue
    if path.name == "bootstrap.sh":
        continue
    if not path.is_file():
        continue
    try:
        data = path.read_text()
    except UnicodeDecodeError:
        continue
    updated = data.replace(old, new)
    if updated != data:
        path.write_text(updated)
PY
}

update_command_metadata() {
	local old="$1"
	local new="$2"
	python3 - "$old" "$new" <<'PY'
import sys
from pathlib import Path

old, new = sys.argv[1], sys.argv[2]
files = [
    ".gitignore",
    "Makefile",
    "README.md",
    "docs/AGENT_CONTEXT.md",
    "goreleaser/.goreleaser.yaml",
    "goreleaser/Dockerfile",
]

replacements = {
    f"cmd/{old}": f"cmd/{new}",
    f"./dist/{old}": f"./dist/{new}",
    f"./{old}": f"./{new}",
    f"COPY {old} ": f"COPY {new} ",
    f"/usr/local/bin/{old}": f"/usr/local/bin/{new}",
    f"project_name: {old}-template": f"project_name: {new}",
    f"id: {old}": f"id: {new}",
    f"binary: {old}": f"binary: {new}",
}

for name in files:
    path = Path(name)
    if not path.exists():
        continue
    data = path.read_text()
    updated = data
    for source, target in replacements.items():
        updated = updated.replace(source, target)
    if updated != data:
        path.write_text(updated)
PY
}

main() {
	local root module_path cmd_name go_version self
	self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
	root="$(dirname "${self}")"
	cd "${root}"

	module_path="${MODULE_PATH:-$(remote_module_path || true)}"
	if [[ -z "${module_path}" ]]; then
		"${GUM[@]}" log --level error "Could not infer module path from git remote.origin.url" "hint" "set MODULE_PATH or configure git remote"
		exit 1
	fi

	cmd_name="${CMD_NAME:-$(module_basename "${module_path}")}"
	go_version="$(latest_go_major_minor)"

	"${GUM[@]}" style --bold --border rounded --padding "1 2" "Bootstrapping ${module_path}"
	"${GUM[@]}" log --structured --level info "Resolved template settings" module "${module_path}" cmd "${cmd_name}" go "${go_version}"

	"${GUM[@]}" log --level info "Updating module path"
	replace_all "${OLD_MODULE}" "${module_path}"

	if [[ "${cmd_name}" != "${OLD_CMD}" && -d "cmd/${OLD_CMD}" ]]; then
		"${GUM[@]}" spin --spinner dot --title "Renaming command" -- mv "cmd/${OLD_CMD}" "cmd/${cmd_name}"
	fi

	"${GUM[@]}" log --level info "Updating command references"
	update_command_metadata "${OLD_CMD}" "${cmd_name}"
	"${GUM[@]}" spin --spinner dot --title "Updating Go version" -- go mod edit -go="${go_version}"
	"${GUM[@]}" spin --spinner dot --title "Tidying module" -- go mod tidy

	"${GUM[@]}" log --structured --level info "Template configured" module "${module_path}" cmd "${cmd_name}" go "${go_version}"
	if [[ "${BOOTSTRAP_KEEP:-0}" != "1" ]]; then
		"${GUM[@]}" log --level info "Removing bootstrap script" path "${self}"
		rm -- "${self}"
	fi
}

main "$@"
