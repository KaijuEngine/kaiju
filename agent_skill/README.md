# Kaiju Agent Skills

This folder contains [Agent Skills](https://docs.claude.com/en/docs/agents/skills) for
working on the [Kaiju Engine](https://kaijuengine.com). Each skill is a directory holding a
`SKILL.md` (with YAML frontmatter: `name` + `description`) plus optional `reference/` files
the agent loads on demand.

| Skill | Folder | What it does |
|-------|--------|--------------|
| `kaijuengine-game-dev` | [`kaijuengine-game-dev/`](kaijuengine-game-dev/) | Building games/tools on the engine: `GameInterface` bootstrap, `Host` runtime, entities & the custom `matrix` library, the Vulkan Drawing system, the HTML/CSS-like UI, build tags, testing. |
| `kaijuengine-aidriver` | [`kaijuengine-aidriver/`](kaijuengine-aidriver/) | Drive a *running* Kaiju game via its built-in AI Driver HTTP server (screenshot + inject mouse/keyboard) when built with the `ai_driver` tag. |

The skill format is Anthropic's. Tools that natively support skills (Claude Code, the Claude
apps) load them directly. Other agentic tools don't have a "skills" concept, but they all read
some form of rules/instructions file — so you point that file at these markdown files instead.

---

## Native skill support

### Claude Code (CLI / IDE extensions)

Skills are auto-discovered from two locations:

- **Personal** (all your projects): `~/.claude/skills/`
- **Project** (this repo only, shareable via git): `<repo>/.claude/skills/`

Copy or symlink each skill directory into one of those. Symlinking keeps them in sync with
this folder:

```bash
# Personal — available in every project
mkdir -p ~/.claude/skills
ln -s "$PWD/kaijuengine-game-dev"   ~/.claude/skills/kaijuengine-game-dev
ln -s "$PWD/kaijuengine-aidriver"   ~/.claude/skills/kaijuengine-aidriver

# …or project-scoped (checked in, shared with the team)
mkdir -p .claude/skills
ln -s "$PWD/kaijuengine-game-dev"   .claude/skills/kaijuengine-game-dev
ln -s "$PWD/kaijuengine-aidriver"   .claude/skills/kaijuengine-aidriver
```

Run `cd <skill>` from the repo root so `$PWD` resolves correctly, or use absolute paths.
Verify with `/skills` (or just start a session — skills are listed at startup). No restart of
an existing session picks them up; start a new one.

### Claude apps (claude.ai / Claude Desktop)

Enable skills under **Settings → Capabilities → Skills**, then upload each skill **folder**
(or a zip of it). `SKILL.md` must sit at the top level of the uploaded folder. Desktop also
honors skills placed in its config directory — see the in-app Skills panel for the path on
your OS.

### Anthropic API / Agent SDK

Skills are mounted from a directory passed to the agent runtime. Point your SDK/agent config
at this folder (or a copy of it) as the skills source; the loader reads each subdirectory's
`SKILL.md`. See the [Agent Skills docs](https://docs.claude.com/en/docs/agents/skills) for the
exact option name in your SDK version.

---

## Tools without native skills

These tools load a rules/instructions file automatically. They won't parse the `SKILL.md`
frontmatter, so the reliable pattern is: **add a short pointer in the tool's rules file that
tells the agent to read the skill markdown when relevant.** Keep the heavy content in the
skill files; keep the rules file thin so it doesn't bloat every prompt.

Drop a block like this into the relevant file:

```markdown
## Kaiju Engine skills
When working on Kaiju Engine code (anything importing `kaijuengine.com/...`), first read
`agent_skill/kaijuengine-game-dev/SKILL.md` and its `reference/` files.
To inspect or drive a running game built with the `ai_driver` tag, read
`agent_skill/kaijuengine-aidriver/SKILL.md`.
```

Where that block goes, per tool:

| Tool | File it reads | Notes |
|------|---------------|-------|
| **Codex CLI**, **Jules**, **Amp**, **Gemini CLI**, others on the [AGENTS.md](https://agents.md) convention | `AGENTS.md` (repo root) | Growing cross-tool standard. Safe default. |
| **Cursor** | `.cursor/rules/*.mdc` | One file per rule. Add frontmatter `description:` and set it to *Always* or *Agent-requested*; the body can `@`-reference the skill files. Cursor also reads root `AGENTS.md`. |
| **Windsurf** | `.windsurf/rules/*.md` | Per-rule activation mode (Always-on / Model-decision / Glob). |
| **Cline** | `.clinerules/` (dir) or `.clinerules` (file) | Markdown; all files in the dir are loaded. |
| **Roo Code** | `.roo/rules/` | Same idea as Cline. |
| **Aider** | `CONVENTIONS.md` | Add with `aider --read CONVENTIONS.md` (or set it in `.aider.conf.yml`). |
| **GitHub Copilot** | `.github/copilot-instructions.md` | Repo-wide custom instructions. |
| **Continue** | `config.yaml` rules / `.continue/rules/` | Reference the skill files from a rule block. |

Because the skill bodies are plain Markdown, you can also just paste their contents directly
into any of the above files if a tool doesn't reliably follow file pointers — at the cost of a
larger always-on prompt.

---

## Keeping skills up to date

These skills are distilled from the engine's `AGENTS.md` and are intentionally a snapshot. When
the skill text disagrees with the current engine source, **trust the source**. If you symlinked
(rather than copied) the skills into your tool's directory, a `git pull` on this repo updates
them everywhere automatically.
