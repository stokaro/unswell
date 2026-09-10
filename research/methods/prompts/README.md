# Pilot prompts

These files hold the verbatim prompt conditions of
[`unswell-llm-patterns-v1`](../llm-patterns-v1.md), fixed on September 10,
2026, before any generation. A change to a prompt is a protocol amendment. A
changed prompt gets a new file with a new version suffix, and the old file
stays in place for the runs that used it.

Each prompt receives the task material in a separate message and asks for one
document of a stated length. Neither prompt names a construction under study,
asks for typical AI text, or asks the model to avoid any pattern. The tooling
picks the opening line for the operation from the table below; the shared
body follows it.

| Operation | Opening line |
| --- | --- |
| `generate` | "Write the documentation described below." |
| `polish` | "Edit the text below for clarity and correctness. Keep its facts, structure, and length." |
| `expand` | "Expand the text below with the additional facts provided. Keep its existing statements." |
| `condense` | "Shorten the text below to the requested length. Keep every stated requirement." |
