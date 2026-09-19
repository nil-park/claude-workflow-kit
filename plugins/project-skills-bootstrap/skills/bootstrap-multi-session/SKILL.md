---
name: bootstrap-multi-session
description: >-
  multi-session 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다. Claude Code 전용이다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`multi-session.md`를 읽어 프로젝트에 설치하거나 업데이트한다.

- Codex에서 불렀다면 설치하지 않고, 이 스킬이 Claude Code 전용이라고 알린다.
- 설치본은 `.agents/skills/`에 쓰지 않고, 심볼릭 링크 없이 `.claude/skills/multi-session/SKILL.md`에 쓴다.
  - Codex는 `.agents/skills/`에서 스킬을 읽는다.

## 사전 설치 확인

`multi-session`은 본문에서 아래 스킬을 이름으로 참조한다. 설치 또는 업데이트를 시작하기 전에 각
스킬이 프로젝트에 설치되어 있는지 확인하고, 없으면 해당 부트스트랩 스킬로 먼저 설치한다.

| 스킬               | 설치 여부를 확인할 곳                      | 부트스트랩 스킬                                       |
| ------------------ | ------------------------------------------ | ----------------------------------------------------- |
| `docs-standards`   | `.agents/skills/docs-standards/SKILL.md`   | `project-skills-bootstrap:bootstrap-docs-standards`   |
| `coding-standards` | `.agents/skills/coding-standards/SKILL.md` | `project-skills-bootstrap:bootstrap-coding-standards` |
| `work-cycle`       | `.agents/skills/work-cycle/SKILL.md`       | `project-skills-bootstrap:bootstrap-work-cycle`       |
| `git-workflow`     | `.agents/skills/git-workflow/SKILL.md`     | `project-skills-bootstrap:bootstrap-git-workflow`     |

## 설치

`.claude/skills/multi-session/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `multi-session.md`를 읽는다.
2. `.claude/skills/multi-session/SKILL.md`에 내용을 쓴다.
3. 완료 후 설치 결과를 보고한다.

## 업데이트

`.claude/skills/multi-session/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `multi-session.md`(템플릿)를 읽는다.
2. `.claude/skills/multi-session/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- 베이스 디렉터리를 기준으로 `../bootstrap/UPDATE.md`(템플릿)를 읽는다.
- `.agents/skills/UPDATE.md`가 없으면 템플릿의 내용으로 만든다.
- 파일이 있는데 내용이 템플릿과 다르면 템플릿에 맞춘다. 프로젝트가 덧붙인 내용은 유지한다.
