---
name: bootstrap-anti-claudeism
description: >-
  anti-claudeism 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`anti-claudeism.md`와 `references/` 파일들을 읽어 프로젝트에 설치하거나 업데이트한다.

## 설치

`.agents/skills/anti-claudeism/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `anti-claudeism.md`를 읽어 `.agents/skills/anti-claudeism/SKILL.md`에 쓴다.
2. 베이스 디렉터리 아래 `references/word-level.md`를 읽어
   `.agents/skills/anti-claudeism/references/word-level.md`에 쓴다.
3. 베이스 디렉터리 아래 `references/sentence-level.md`를 읽어
   `.agents/skills/anti-claudeism/references/sentence-level.md`에 쓴다.
4. `.claude/skills/anti-claudeism`에 `.agents/skills/anti-claudeism`을 가리키는
   심볼릭 링크를 만든다.
5. 완료 후 설치 결과를 보고한다.

## 업데이트

`.agents/skills/anti-claudeism/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `anti-claudeism.md`, `references/word-level.md`,
   `references/sentence-level.md`(템플릿)를 읽는다.
2. `.agents/skills/anti-claudeism/` 아래 대응하는 파일들(현재 파일)을 읽는다.
3. 파일별로 차이를 비교해 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.
