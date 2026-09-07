---
name: bootstrap
description: >-
  docs-standards, coding-standards, work-cycle, fluent-korean, scratch-dir 스킬을
  현재 프로젝트에 한 번에 설치하거나 업데이트하고 싶을 때 부른다. anti-claudeism과
  git-workflow는 이 묶음에 들어 있지 않으므로 각각 따로 부른다.
---

## 이 묶음이 설치하는 스킬

아래 스킬들을 순서대로 실행한다.

- `project-skills-bootstrap:bootstrap-docs-standards`
- `project-skills-bootstrap:bootstrap-coding-standards`
- `project-skills-bootstrap:bootstrap-work-cycle`
- `project-skills-bootstrap:bootstrap-fluent-korean`
- `project-skills-bootstrap:bootstrap-scratch-dir`

## 이 묶음에서 제외한 스킬

사용자의 명시적인 요청이 있을 때만 설치한다.

- `project-skills-bootstrap:bootstrap-anti-claudeism`
- `project-skills-bootstrap:bootstrap-git-workflow`

## 사용자 가이드

- `docs-standards`: 리포지토리 안의 문서를 어떤 위치에 어떤 기준으로 작성할지 정하고, 주석·이슈·PR/MR 본문과 리뷰에도 적용한다.
- `coding-standards`: 코드 작성과 리뷰에서 따를 공통 기준을 정한다.
- `work-cycle`: 문서를 작성하거나 구현한 뒤 `docs-standards`와 `coding-standards`를 기준으로 리뷰 라운드를 세 번 연속 지적이 없을 때까지 반복한다.
- `fluent-korean`: 한국어 출력의 문장 구성과 표현 규칙을 정한다.
- `scratch-dir`: 커밋하지 않는 임시 작업 파일의 위치와 이름을 정한다.
- `anti-claudeism`: Opus와 Sonnet 5의 한국어 문체 결함 교정 기준을 정하고, 낱말 단위의 결함을 Stop 훅으로 탐지한다.
- `git-workflow`: 브랜치 생성과 GitHub 또는 GitLab의 PR/MR 작업 절차를 정한다.
