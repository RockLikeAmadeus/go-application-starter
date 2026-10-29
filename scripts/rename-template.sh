#!/usr/bin/env bash
# Re-brand a fresh copy of this template.
#
#   scripts/rename-template.sh github.com/you/your-app your-app
#
# Rewrites the Go module path everywhere and the app name in the Fly configs
# and web/package.json. Run once, right after creating the repo from the template.
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "usage: $0 <new-go-module-path> <new-app-name>" >&2
  exit 1
fi

old_module="github.com/RockLikeAmadeus/go-application-starter"
old_app="go-application-starter"
new_module="$1"
new_app="$2"

cd "$(git rev-parse --show-toplevel)"

files=$(git ls-files | grep -vE '(^|/)(pnpm-lock\.yaml|go\.sum)$' | xargs grep -lIF -e "$old_module" -e "$old_app" -e "starter-web" || true)
for f in $files; do
  perl -pi -e "s#\Q$old_module\E#$new_module#g; s#\Q$old_app\E#$new_app#g; s#\"starter-web\"#\"$new_app-web\"#g" "$f"
done

go mod tidy
echo "Renamed to $new_module ($new_app). Next:"
echo "  task gen && task check"
echo "  Then update SPEC.md and replace the sample Item domain."
