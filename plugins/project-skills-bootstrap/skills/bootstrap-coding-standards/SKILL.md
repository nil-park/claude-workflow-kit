---
name: bootstrap-coding-standards
description: >-
  coding-standards 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`coding-standards.md`를 읽어 프로젝트에 설치하거나 업데이트한다.

## 설치

`.agents/skills/coding-standards/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `coding-standards.md`를 읽는다.
2. `.agents/skills/coding-standards/SKILL.md`에 내용을 쓴다.
3. `.claude/skills/coding-standards`에 `.agents/skills/coding-standards`을 가리키는
   심볼릭 링크를 만든다.
4. 완료 후 설치 결과를 보고한다.

## 업데이트

`.agents/skills/coding-standards/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `coding-standards.md`(템플릿)를 읽는다.
2. `.agents/skills/coding-standards/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- `.agents/skills/UPDATE.md`가 없으면 아래 내용으로 만든다.
- 파일은 있으나 목록에 `coding-standards` 항목이 없으면 추가한다.

```markdown
# 설치된 스킬의 출처와 업데이트 방법

이 파일이 있는 디렉터리의 스킬들은 아래 마켓플레이스가 제공하는
`project-skills-bootstrap` 플러그인이 복사해 넣은 사본이다. 원본이 갱신되더라도 사본은
저절로 따라가지 않으므로, 반영하려면 아래 절차를 직접 실행해야 한다.

- 마켓플레이스: `claude-workflow-kit` (<https://github.com/nil-park/claude-workflow-kit>)
- 플러그인: `project-skills-bootstrap`

원본의 변경을 사본에 반영하려면, 반영하려는 스킬에 대응하는 부트스트랩 스킬을 다시
부른다. 부트스트랩 스킬은 사본이 이미 있으면 업데이트 절차로 동작하여, 템플릿과 사본의
차이를 보고하고 무엇을 반영할지 사용자와 상의한다. 따라서 사본을 프로젝트에 맞게
고쳐 둔 부분이 있더라도 그대로 유지할 수 있다.

| 설치된 스킬        | 설치 위치                          | 다시 부를 부트스트랩 스킬                             |
| ------------------ | ---------------------------------- | ----------------------------------------------------- |
| `coding-standards` | `.agents/skills/coding-standards/` | `project-skills-bootstrap:bootstrap-coding-standards` |
```
