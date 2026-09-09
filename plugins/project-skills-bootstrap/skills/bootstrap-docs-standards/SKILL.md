---
name: bootstrap-docs-standards
description: >-
  docs-standards 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`docs-standards.md`를 읽어 프로젝트에 설치하거나 업데이트한다.

## 설치

`.agents/skills/docs-standards/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `docs-standards.md`를 읽는다.
2. `.agents/skills/docs-standards/SKILL.md`에 내용을 쓴다.
3. `.claude/skills/docs-standards`에 `.agents/skills/docs-standards`을 가리키는
   심볼릭 링크를 만든다.
4. 완료 후 설치 결과를 보고한다.

## 업데이트

`.agents/skills/docs-standards/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `docs-standards.md`(템플릿)를 읽는다.
2. `.agents/skills/docs-standards/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- `.agents/skills/UPDATE.md`가 없으면 아래 내용으로 만든다.
- 파일은 있으나 목록에 `docs-standards` 항목이 없으면 추가한다.

```markdown
# 스킬 업데이트

## 업스트림

- 리포지토리: https://github.com/nil-park/claude-workflow-kit
- 경로: `plugins/project-skills-bootstrap/skills/`
- 업스트림의 `bootstrap-<스킬명>/SKILL.md`는 설치·업데이트 로더이고, 실제 템플릿은 같은 디렉터리에 있는 아래 표의 마크다운 파일이다.

## 대상 스킬

| 로컬 경로                 | 업스트림 템플릿                              |
| ------------------------- | -------------------------------------------- |
| `docs-standards/SKILL.md` | `bootstrap-docs-standards/docs-standards.md` |

## 업데이트 절차

- `project-skills-bootstrap:bootstrap-<스킬명>` 스킬을 실행하면, 이미 설치된 스킬은 업스트림 템플릿과 비교하는 업데이트 흐름으로 진행된다.
```
