#!/usr/bin/env bash
set -euf

VERSION=""
MODE=""
DRY_RUN=false

usage() {
	cat <<EOF
Usage: $(basename "$0") [OPTIONS] [VERSION]

Release a new version by creating and pushing a signed git tag.

Options:
  -M, --major     Bump major version (e.g., 1.2.3 -> 2.0.0)
  -m, --minor     Bump minor version (e.g., 1.2.3 -> 1.3.0)
  -p, --patch     Bump patch version (e.g., 1.2.3 -> 1.2.4)
  -n, --dry-run   Show what would happen without making changes
  -h, --help      Show this help message

Examples:
  $(basename "$0")              # Interactive mode
  $(basename "$0") --patch      # Bump patch version
  $(basename "$0") --minor -n   # Preview minor bump
  $(basename "$0") 1.2.3        # Release specific version
EOF
	exit 0
}

get_current_version() {
	local current
	current=$(git describe --tags "$(git rev-list --tags --max-count=1)" 2>/dev/null || true)
	if [[ -z ${current} ]]; then
		current=0.0.0
	fi
	echo "${current#v}"
}

# Calculate all bumped versions in a single Python call
calc_versions() {
	local current=$1
	uv run --with semver python3 -c "
import semver, sys
v = semver.VersionInfo.parse(sys.argv[1])
print(f'{v.bump_major()}|{v.bump_minor()}|{v.bump_patch()}')
" "${current}"
}

bump_version() {
	local current major minor patch
	current=$(get_current_version)
	echo "Current version is ${current}"

	IFS='|' read -r major minor patch <<<"$(calc_versions "${current}")"

	echo "If we bump we get, Major: ${major} Minor: ${minor} Patch: ${patch}"
	read -r -p "To which version you would like to bump [M]ajor, Mi[n]or, [P]atch or Manua[l]: " ANSWER
	case ${ANSWER,,} in
	m) VERSION="${major}" ;;
	n) VERSION="${minor}" ;;
	p) VERSION="${patch}" ;;
	l) read -r -p "Enter version: " -e VERSION ;;
	*)
		echo "no or bad reply??"
		exit 1
		;;
	esac
}

apply_mode() {
	local current
	current=$(get_current_version)
	VERSION=$(uv run --with semver python3 -c "
import semver, sys
v = semver.VersionInfo.parse(sys.argv[1])
print(getattr(v, 'bump_${MODE}')())
" "${current}")
}

# Parse arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	-h | --help) usage ;;
	-M | --major)
		MODE="major"
		shift
		;;
	-m | --minor)
		MODE="minor"
		shift
		;;
	-p | --patch)
		MODE="patch"
		shift
		;;
	-n | --dry-run)
		DRY_RUN=true
		shift
		;;
	-*)
		echo "Unknown option: $1"
		usage
		;;
	*)
		VERSION="$1"
		shift
		;;
	esac
done

[[ $(git rev-parse --abbrev-ref HEAD) != main ]] && {
	echo "you need to be on the main branch"
	exit 1
}

# Determine version
if [[ -n ${MODE} ]]; then
	apply_mode
elif [[ -z ${VERSION} ]]; then
	bump_version
fi

[[ -z ${VERSION} ]] && {
	echo "no version specified"
	exit 1
}

if [[ ${VERSION} != v* ]]; then
	VERSION="v${VERSION}"
fi

if ${DRY_RUN}; then
	echo "[dry-run] Would release version ${VERSION}"
	echo "[dry-run] git tag -s ${VERSION} -m \"Releasing version ${VERSION}\""
	echo "[dry-run] git push --tags origin ${VERSION}"
	exit 0
fi

echo "Releasing version ${VERSION}"
git tag -s ${VERSION} -m "Releasing version ${VERSION}"
git push --tags origin ${VERSION}
git pull origin main
git push origin main
