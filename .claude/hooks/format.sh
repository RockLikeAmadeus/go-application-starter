#!/usr/bin/env bash
# PostToolUse hook: format the file Claude just edited so diffs stay clean.
# Receives the tool call as JSON on stdin. Never blocks: failures are ignored.
set -u

file=$(jq -r '.tool_input.file_path // empty' 2>/dev/null) || exit 0
[ -n "$file" ] && [ -f "$file" ] || exit 0

case "$file" in
  *.gen.go | *.gen.ts) exit 0 ;;
  *.go)
    gofmt -w "$file" 2>/dev/null
    ;;
  "$CLAUDE_PROJECT_DIR"/web/*.ts | "$CLAUDE_PROJECT_DIR"/web/*.tsx | "$CLAUDE_PROJECT_DIR"/web/*.css | "$CLAUDE_PROJECT_DIR"/web/*.json)
    (cd "$CLAUDE_PROJECT_DIR/web" && pnpm exec prettier --write --log-level silent "$file") 2>/dev/null
    ;;
esac
exit 0
