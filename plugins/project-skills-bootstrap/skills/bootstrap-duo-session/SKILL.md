---
name: bootstrap-duo-session
description: >-
  duo-session-agy 스킬을 프로젝트에 설치하거나 업데이트하며, 이미 설치된 duo-session 스킬도 업데이트한다. Claude Code 전용이다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래의 템플릿을 읽는다.
`duo-session-agy`를 프로젝트에 설치하거나 업데이트하며, `duo-session`이 이미 설치되어 있으면 함께 업데이트한다.

- Codex에서 불렀다면 설치하지 않고, 이 스킬이 Claude Code 전용이라고 알린다.
- 두 스킬의 설치본은 `.agents/skills/`에 쓰지 않고, 심볼릭 링크 없이 `.claude/skills/<스킬 이름>/SKILL.md`에 쓴다.
  - Codex는 `.agents/skills/`에서 스킬을 읽는다.
- 새로 설치할 때는 `duo-session-agy`만 설치하며, `duo-session`은 설치하지 않는다.

## 사전 설치 확인

설치 또는 업데이트하려는 스킬이 필요로 하는 스킬만 골라 프로젝트에 설치되어 있는지 확인하고, 없으면 해당 부트스트랩 스킬로 먼저 설치한다.

| 스킬               | 이 스킬을 필요로 하는 스킬       | 설치 여부를 확인할 곳                      | 부트스트랩 스킬                                       |
| ------------------ | -------------------------------- | ------------------------------------------ | ----------------------------------------------------- |
| `docs-standards`   | `duo-session-agy`, `duo-session` | `.agents/skills/docs-standards/SKILL.md`   | `project-skills-bootstrap:bootstrap-docs-standards`   |
| `coding-standards` | `duo-session-agy`, `duo-session` | `.agents/skills/coding-standards/SKILL.md` | `project-skills-bootstrap:bootstrap-coding-standards` |
| `anti-claudeism`   | `duo-session-agy`, `duo-session` | `.agents/skills/anti-claudeism/SKILL.md`   | `project-skills-bootstrap:bootstrap-anti-claudeism`   |
| `git-workflow`     | `duo-session-agy`, `duo-session` | `.agents/skills/git-workflow/SKILL.md`     | `project-skills-bootstrap:bootstrap-git-workflow`     |
| `scratch-dir`      | `duo-session-agy`                | `.agents/skills/scratch-dir/SKILL.md`      | `project-skills-bootstrap:bootstrap-scratch-dir`      |
| `work-cycle`       | `duo-session`                    | `.agents/skills/work-cycle/SKILL.md`       | `project-skills-bootstrap:bootstrap-work-cycle`       |

## 설치

`.claude/skills/duo-session-agy/SKILL.md`가 없을 때 실행한다.

- `bootstrap` 스킬이 업데이트 절차로 이 스킬을 불렀다면 설치하지 않는다.

1. 베이스 디렉터리 아래 `duo-session-agy.md`를 읽어 `.claude/skills/duo-session-agy/SKILL.md`에 쓴다.
2. 베이스 디렉터리 아래 `duo-session-agy-peer.md`를 읽어 `.agents/agents/duo-session-agy-peer/agent.md`에 쓴다.
3. `duo-relay`가 `PATH`에 없으면 `user-tools-bootstrap` 플러그인의 `bootstrap-statusline`이 바이너리를 설치하는 방식과 같게 설치한다.
   - 설치할 디렉터리는 사용자와 상의해 정하며, `PATH`에 들어 있는 곳이어야 한다.
   - `GOBIN=<디렉터리의 절대 경로> go install github.com/nil-park/claude-workflow-kit/duo-relay@latest`를 실행한다.
   - 이 명령에는 `go`가 필요하며, `go`가 없으면 사용자에게 설치를 요청한다.
4. `.claude/settings.json`의 `permissions.allow`에 `Bash(duo-relay:*)`가 없으면 추가한다.
5. `.agents/rules/`에 상시 로드 규칙이 있는지 확인한다.
   - 규칙이 있으면 `peer`도 그 규칙을 함께 읽어 들인다고 사용자에게 알린다.
   - `work-cycle` 및 `git-workflow` 규칙이 `peer`의 역할과 겹친다는 점도 함께 알린다.
6. `~/.gemini/antigravity-cli/settings.json`에서 `toolPermission`이 `always-proceed`이고 `allowNonWorkspaceAccess`가 `true`인지 확인한다.
   - 둘 중 하나라도 다르면 사용자에게 설정을 요청한다.
   - 이 설정 파일은 직접 고치지 않는다.
7. 완료 후 설치 결과를 보고한다.

## 업데이트

이미 설치되어 있는 스킬마다 아래 절차를 실행한다.

### `duo-session-agy`

`.claude/skills/duo-session-agy/SKILL.md`가 이미 있을 때 실행한다.

1. 설치된 각 파일을 베이스 디렉터리 아래의 템플릿과 비교해 차이를 사용자에게 보고한다.
   - `.claude/skills/duo-session-agy/SKILL.md`: 템플릿 `duo-session-agy.md`
   - `.agents/agents/duo-session-agy-peer/agent.md`: 템플릿 `duo-session-agy-peer.md`
2. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 각 파일을 수정한다.
3. 현재 `duo-relay`가 있는 디렉터리에 같은 `go install` 명령으로 최신 커밋의 `duo-relay`를 다시 빌드한다.
4. 설치 4~6단계를 다시 실행한다.

### `duo-session`

`.claude/skills/duo-session/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `duo-session.md`(템플릿)를 읽는다.
2. `.claude/skills/duo-session/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- 베이스 디렉터리를 기준으로 `../bootstrap/UPDATE.md`(템플릿)를 읽는다.
- `.agents/skills/UPDATE.md`가 없으면 템플릿의 내용으로 만든다.
- 파일이 있는데 내용이 템플릿과 다르면 템플릿에 맞춘다. 프로젝트가 덧붙인 내용은 유지한다.
