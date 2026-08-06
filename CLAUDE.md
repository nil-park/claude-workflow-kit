# CLAUDE.md

## Repository

- Remote: https://github.com/nil-park/claude-workflow-kit

## Required Reading

- [README.md](README.md) — big-picture overview; read first. Describes major features only, so don't assume a feature is missing just because a newly added one isn't listed there.
- https://code.claude.com/docs/en/skills.md — `SKILL.md` authoring: frontmatter, structure, supporting files. Read when editing or adding a skill under `plugins/*/skills/`.
- https://code.claude.com/docs/en/plugin-marketplaces.md — `marketplace.json`, plugin sources, hosting, and the install/update flow. Read when changing the catalog, adding a plugin, or touching how plugins are distributed.
- https://code.claude.com/docs/en/plugins-reference.md — authoritative schema for `plugin.json` and `marketplace.json`. Read when unsure about a manifest field, component path, or version resolution.

## Language

- CLAUDE.md, commit messages, and PR titles/descriptions: English.
- Issues and everything under `docs/`: Korean.
- Skill bodies: Korean. Plugin and marketplace manifests stay English.

## Git Convention

- Branch naming and PR conventions follow the `git-workflow` skill in this repo.

## Development Convention

- On the first prompt of a session, run `eza --tree --git-ignore -a --ignore-glob='.git' .` for an up-to-date, `.gitignore`-respecting tree of the project. The trailing `.` is required (without an explicit path this build prints nothing); `-a` shows dotfiles like `.claude-plugin/`, and `.git` is excluded.

## Formatting

- Markdown, JSON, and YAML are formatted with Prettier. Run `make format` (which runs `npx prettier`) after editing any of these files.
- Run `make format` once before committing to `main`, or as the last step of a PR, so formatting never lands in a separate follow-up.
- `.prettierignore` excludes build outputs and tooling dirs; new build/generated paths that emit these extensions should be added there.
- Prettier markdown gotcha: a wrapped prose line that begins with `+`, `-`, or `*` is reparsed as a list marker. Don't start a continuation line with those (e.g. write `및`/`and`, not a leading `+`), or the formatter will turn it into a nested bullet and change the meaning.
