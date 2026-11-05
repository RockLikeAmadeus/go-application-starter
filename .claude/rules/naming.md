---
paths:
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
  - "api/openapi.yaml"
---

# Naming types and values

Name a type after **what its values are**, not after the situation in which they are used.

- **The "is a" test:** for each value the type can hold, "`<value>` is a `<Type>`" must read as a true sentence. `BlockNew` is an `OverLimitBehavior` ✓. `BlockNew` is an `OverLimit` ✗: `BlockNew` is something done *when* over the limit, not an over-limit. If the sentence fails, the name describes the situation, so add the noun for what the values actually are (`…Behavior`, `…Policy`, `…Mode`, `…Kind`, `…Status`, `…Reason`). The exception is a type whose package-qualified name reads as English where it's used, in the `FromX` / `ToY` style: `workspace.FromContext{}` ("the workspace, from the context") and `workspace.Local{}` stay as they are, rather than `workspace.ContextResolver{}`.
- **When in doubt, prefer the name that makes the calling code read like English.** Fluent code is the aim: read the call site aloud (`workspace.FromContext{}`, `entitlements.CheckCreate(limitName, count)`, `if feature.IsOn`) and pick the name that makes it a natural sentence, even if it reads less well where it's declared.
- **A field or variable names its role, and its type names its kind.** Context can carry the situation there: `Limit.WhenOver` of type `OverLimitBehavior` reads as "when over: block-new". Don't repeat the context in the field name (`Limit.OverLimitBehavior`), and don't leave it out of the type name (`OverLimit`), because a type is often read with no field next to it (in signatures, maps, tests).
- **Check that a variable reads as a sentence where it is used,** not only where it is declared. Being short isn't a reason to give it a name that describes the wrong thing.
- **Name a boolean as a question it answers:** a variable, field, function or method that is or returns a `bool` starts with a word like `is`, `has`, `should`, `can`, `was` or `needs`: `isOver`, `hasAny`, `shouldRun`, `isValid()`, never `over`, `any`, `run` or `valid()`. The exceptions: Go's `ok` in `v, ok := m[k]`; a function named as a verb question that already reads as yes or no (`keepsDataInAFolder(cfg)`, `matches(sign, name)`, `exists(path)`); and test-table fields such as `want` and `wantErr`.
- **Prefer the longer, descriptive name,** within reason. `validationErr` beats `verr`, `feature` beats `f`, and `item` beats `it`. The wider the scope and the more visible the symbol, the more this matters: a one-line loop body or a method receiver can stay short (`r`, `i`), but a package-level name, a field, a parameter or a variable used across a function should say what it holds. This also matters more the more potentially ambiguous an abbreviation is (is `conString` the `connectionString` or the `convertedString`?) Don't lengthen a name so far that it repeats its context (see `WhenOver` above). Go's conventional short names are exceptions: `ctx`, `err`, `t`, `w` and `r` in an HTTP handler, `ok`, method receivers, and test-table fields such as `want`, `got` and `wantErr`.

When renaming, use the language server (`gopls rename`, or VS Code's Rename Symbol), then `git grep -nw <OldName>` for the docs, comments, Gherkin and YAML that it doesn't update.
