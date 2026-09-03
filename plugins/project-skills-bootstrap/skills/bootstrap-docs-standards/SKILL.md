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
