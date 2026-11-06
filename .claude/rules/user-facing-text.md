---
paths:
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
  - "api/openapi.yaml"
  - "features/**"
  - "SPEC.md"
---

# User-facing text

User-facing text is what a person using the product reads: warnings, error messages, labels, empty states, command output and help text. Log and developer-only text doesn't count.

**Always agree it with me first.** Before writing or changing user-facing text, ask me with the AskUserQuestion tool, even if you're happy with the wording:
- Ask about one piece of text at a time.
- Show the old wording, if there is any.
- Offer two or more complete suggestions to choose from, each with the exact text, and recommend one.
- Say if the wording raises a question of accuracy: check that what it tells the user is true of the code, and that any instruction in it would work.

**Style for the suggestions:**
- Formal, in complete sentences, even if that's more verbose: no contractions, "the application" rather than "the app", and "please" before an instruction ("cannot", "could not", "Please close the application").
- Where space is at a premium, such as a note beside a control or a line of key help, keep it short instead. Short text still gets agreed with me.
- The Oxford comma is mandatory: "memory", "jsonfile", or "sqlite". This includes lists that code joins.
- Use the same word for the same thing as other messages: "conflicted copy", "file sync service", "plain-text storage", "your current plan".
