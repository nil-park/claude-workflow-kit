---
name: bootstrap-multi-session
description: >-
  multi-session, multi-session-copilot, multi-session-gemini 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다. Claude Code 전용이다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래의 공통
템플릿과 변형 파일을 조립해, peer로 쓰는 CLI마다 스킬을 하나씩 설치하거나 업데이트한다.

- Codex에서 불렀다면 설치하지 않고, 이 스킬이 Claude Code 전용이라고 알린다.
- 설치본은 `.agents/skills/`에 쓰지 않고, 심볼릭 링크 없이 `.claude/skills/`에 쓴다.
  - Codex는 `.agents/skills/`에서 스킬을 읽는다.

| 변형 파일                 | 스킬 이름               | 설치본                                          |
| ------------------------- | ----------------------- | ----------------------------------------------- |
| `variants/claude-code.md` | `multi-session`         | `.claude/skills/multi-session/SKILL.md`         |
| `variants/copilot.md`     | `multi-session-copilot` | `.claude/skills/multi-session-copilot/SKILL.md` |
| `variants/gemini.md`      | `multi-session-gemini`  | `.claude/skills/multi-session-gemini/SKILL.md`  |

## 사전 설치 확인

세 스킬은 본문에서 아래 스킬을 이름으로 참조한다. 설치 또는 업데이트를 시작하기 전에 각
스킬이 프로젝트에 설치되어 있는지 확인하고, 없으면 해당 부트스트랩 스킬로 먼저 설치한다.

| 스킬               | 설치 여부를 확인할 곳                      | 부트스트랩 스킬                                       |
| ------------------ | ------------------------------------------ | ----------------------------------------------------- |
| `docs-standards`   | `.agents/skills/docs-standards/SKILL.md`   | `project-skills-bootstrap:bootstrap-docs-standards`   |
| `coding-standards` | `.agents/skills/coding-standards/SKILL.md` | `project-skills-bootstrap:bootstrap-coding-standards` |
| `work-cycle`       | `.agents/skills/work-cycle/SKILL.md`       | `project-skills-bootstrap:bootstrap-work-cycle`       |
| `anti-claudeism`   | `.agents/skills/anti-claudeism/SKILL.md`   | `project-skills-bootstrap:bootstrap-anti-claudeism`   |
| `git-workflow`     | `.agents/skills/git-workflow/SKILL.md`     | `project-skills-bootstrap:bootstrap-git-workflow`     |
| `scratch-dir`      | `.agents/skills/scratch-dir/SKILL.md`      | `project-skills-bootstrap:bootstrap-scratch-dir`      |

## 조립

변형마다 아래 순서로 설치본의 내용을 만든다.

1. 베이스 디렉터리 아래 `multi-session.md`(공통 템플릿)를 읽는다.
2. 베이스 디렉터리 아래 변형 파일을 읽는다.
   - `##` 절 제목이 자리표시자 이름이고, 그 아래 코드 블록의 내용이 값이다.
3. 공통 템플릿의 `{{이름}}`을 같은 이름의 값으로 바꾼다.
   - 자리표시자만 있는 줄은 그 줄 전체를 값으로 바꾼다. 값에 적힌 들여쓰기는 바꾸지 않는다.
   - 값이 비어 있으면 그 줄을 지운다.
4. frontmatter의 `name`을 표의 스킬 이름으로 바꾼다.
5. 바꾼 결과에 `{{`가 남아 있지 않은지 확인한다.

## 설치

세 변형 중 설치본이 없는 것을 설치한다.

- `bootstrap` 스킬이 업데이트 절차로 이 스킬을 불렀다면 설치하지 않는다.
- `multi-session`은 사용자에게 묻지 않고 설치한다.
- `multi-session-copilot`과 `multi-session-gemini`는 설치할지 사용자에게 묻고, 사용자가 고른 것만 설치한다.

설치할 변형마다 아래 순서로 실행한다.

1. 조립 절차로 내용을 만든다.
2. 표의 설치본 경로에 내용을 쓴다.
3. 완료 후 설치 결과를 보고한다.

## 업데이트

세 변형 중 설치본이 이미 있는 것마다 실행한다.

1. 조립 절차로 내용을 만든다(템플릿).
2. 설치본(현재 파일)을 읽는다.
3. 두 내용을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- 베이스 디렉터리를 기준으로 `../bootstrap/UPDATE.md`(템플릿)를 읽는다.
- `.agents/skills/UPDATE.md`가 없으면 템플릿의 내용으로 만든다.
- 파일이 있는데 내용이 템플릿과 다르면 템플릿에 맞춘다. 프로젝트가 덧붙인 내용은 유지한다.
