---
paths:
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
  - "api/openapi.yaml"
---

# Naming types and values

Name a type after **what its values are**, not after the situation in which they are used.

- **The "is a" test:** for each value the type can hold, "`<value>` is a `<Type>`" must read as a true sentence. `BlockNew` is an `OverLimitBehavior` ✓. `BlockNew` is an `OverLimit` ✗: `BlockNew` is something done *when* over the limit, not an over-limit. If the sentence fails, the name describes the situation, so add the noun for what the values actually are (`…Behavior`, `…Policy`, `…Mode`, `…Kind`, `…Status`, `…Reason`).
- **A field or variable names its role, and its type names its kind.** Context can carry the situation there: `Limit.WhenOver` of type `OverLimitBehavior` reads as "when over: block-new". Don't repeat the context in the field name (`Limit.OverLimitBehavior`), and don't leave it out of the type name (`OverLimit`), because a type is often read with no field next to it (in signatures, maps, tests).
- **Check that a variable reads as a sentence where it is used,** not only where it is declared. Being short isn't a reason to give it a name that describes the wrong thing.

When renaming, use the language server (`gopls rename`, or VS Code's Rename Symbol), then `git grep -nw <OldName>` for the docs, comments, Gherkin and YAML that it doesn't update.
