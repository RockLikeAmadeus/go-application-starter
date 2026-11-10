---
paths:
  - "README.md"
  - "SPEC.md"
  - "docs/*.md"
  - "docs/adr/**"
---

# Writing docs for developers

README.md, SPEC.md, the guides in `docs/`, and the ADRs in `docs/adr/` are written for developers. Many of those developers are new to the tools this template uses. Write these docs to be clear, even when that makes them longer.

This rule doesn't cover the working files in `docs/plans/` and `docs/wip/`, or CLAUDE.md and `.claude/`. Claude reads those on almost every task, so they stay short.

- **Don't assume the reader knows a tool, term, or pattern,** even one that seems widely known. The first time a doc mentions one (goose, WAL, a session lock, a readiness probe, a CI service container), say in a sentence what it is and what it does here. Link to its own documentation or website.
- **Write full sentences, with one idea in each.** Don't pack several facts into one sentence with parentheses, semicolons, or colons. A list is fine, but each item is still a sentence.
- **Say why, not only what.** Give the reason behind a rule, a setting, or a step, so the reader can tell when it applies and when it doesn't.
- **Show an example:** a command to run, part of a file, or what the output looks like.
- **Draw a diagram** wherever one makes something easier to follow: how the parts connect, what calls what, the order of steps, or what happens at start-up. Use [Mermaid](https://mermaid.js.org/) in a ```` ```mermaid ```` block, which GitHub draws as a picture, and keep each diagram to the few boxes that matter.
- **Use the Oxford comma,** as everywhere in the repository: "memory, jsonfile, and sqlite".

When a change touches a section of an existing doc, rewrite that section in this style. Leave the sections it doesn't touch for later.
