---
name: bootstrap-work-cycle
description: >-
  work-cycle 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`work-cycle.md`를 읽어 프로젝트에 설치하거나 업데이트한다.

## 사전 설치 확인

`work-cycle`은 리뷰 라운드에서 `docs-standards`와 `coding-standards`를 참조한다. 설치
또는 업데이트를 시작하기 전에 두 스킬이 프로젝트에 설치되어 있는지 확인하고, 없으면 먼저
설치한다.

1. `.agents/skills/docs-standards/SKILL.md`가 없으면 `project-skills-bootstrap:bootstrap-docs-standards`를 실행한다.
2. `.agents/skills/coding-standards/SKILL.md`가 없으면 `project-skills-bootstrap:bootstrap-coding-standards`를 실행한다.
3. 두 스킬이 모두 설치된 것을 확인한 뒤 아래 절차를 진행한다.

## 설치

`.agents/skills/work-cycle/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `work-cycle.md`를 읽는다.
2. `.agents/skills/work-cycle/SKILL.md`에 내용을 쓴다.
3. `.claude/skills/work-cycle`에 `.agents/skills/work-cycle`을 가리키는
   심볼릭 링크를 만든다.
4. 완료 후 설치 결과를 보고한다.

## 업데이트

`.agents/skills/work-cycle/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `work-cycle.md`(템플릿)를 읽는다.
2. `.agents/skills/work-cycle/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.
