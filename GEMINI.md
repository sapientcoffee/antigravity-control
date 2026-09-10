<!-- NOTE: This file is always loaded into context, no matter what you are trying to do! You should ensure that its contents are as universally applicable as possible.-->
<!-- Top Tips
An LLM will perform better on a task when its' context window is full of focused, relevant context including examples, related files, tool calls, and tool results compared to when its context window has a lot of irrelevant context.

* ~60 lines no more than 600 (channel your inner minimalist; less is more)
* If the info is somewhere in the codebase you probably don't need it to be here - Really good at figuring out what files/folders matter for tasks, what commands to run, the dependencies you have
* Use it to steer the model away from things its consistently doing wrong or quirks that keep happening
* Include bash commands that can't be guesses by the model
* Exclude information that changes frequently
* Include unique instructions or team etiquette (branch naming, PR conventions)
* Consider what the model can workout and what it can't (and will) work out by reviewing the codebase; reduce the risk of confusing it -->

# Project: Antigravity Control (agyctl)
<!-- agents operate best on rigid, operational guardrails and specific constraints rather than polite requests or general guidelines. Stick to concrete "Do X, Never do Y" statements. -->
This file describes common mistakes and confusion points that agents might encounter as they work in this project. If you ever encounter something in the project that surprises you please alert the developer working with you and indicate that this is the case in GEMINI.MD file to help prevent future agents from having the same issue

## Setup & Developer Environment
<!-- Gemini will work out dependencies from the codebase (e.g. package.json). Hardcoding in here is like having stale docs -->
- **CLI Framework:** Built using Go (Golang) and Cobra (`github.com/spf13/cobra`).
- **Build CLI:** `go build -o bin/agyctl ./cmd/agyctl`
- **Test:** `go test ./...`

## Deep Context (Progressive Disclosure)
<!-- The Gemini CLI executes a downward Breadth-First-Search (BFS) scan through your project, grabbing context files from subdirectories (up to a limit of 200 folders) and layering them over the root file. Ensure this root file remains strictly for global mandates, and rely heavily on nested GEMINI.md files in your sub-folders for component-specific instructions, as those will be appended closer to the active user prompt  -->
- **Architecture & Metaphor:** ` @./README.md`
- **Persona Pods:** ` @./personas/`
- **Control Skill Spec:** ` @./skills/control/SKILL.md`
- **Persona Skill Spec:** ` @./skills/persona/SKILL.md`

## Rules, Gotchas, & Anti-Patterns
<!-- For comparative data or strict rule matrices, structural analysis shows that formatting these rules into a Markdown table or using YAML/XML structures significantly improves the model's comprehension and token ingestion efficiency compared to plain prose or basic lists -->
<!-- Specify prefered process and specific instructions to be followed -->

| Category | Mandate | Anti-Pattern to Avoid |
| :--- | :--- | :--- |
| **CLI Implementation** | Build the CLI using Go (Golang) and Cobra (`github.com/spf13/cobra`). | Writing new CLI logic in Bash/Python scripts or using non-Cobra frameworks. |
| **Context Hygiene** | Keep active context minimalist (< 15 skills symlinked at any time). | Bulk loading skills or leaving inactive plugins enabled. |
| **Filesystem Safety** | Symlink skills from `~/.gemini/catalog/skills/` to `~/.gemini/config/skills/` atomically. | Destructive file deletion, hardcopies, or modifying catalog skills directly. |
| **Upstream Auditing** | Implement non-blocking checks and cached statuses for upstream plugin version audits. | Synchronous network calls blocking CLI startup or session hooks. |
| **Commits & Style** | Follow standard Go formatting (`gofmt`/`go vet`) and conventional commit messages. | Unformatted code or unstructured commit messages. |
