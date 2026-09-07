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

## UPDATE.md 기록

위 스킬들을 모두 실행한 뒤에 `.agents/skills/UPDATE.md`를 확인한다. 각 부트스트랩 스킬이
자기 항목을 기록하므로 대개 이미 채워져 있으니, 빠진 항목만 보태고 파일이 없으면 아래
내용으로 만든다. `fluent-korean`은 `.agents/skills/` 아래에 사본을 두지 않는 외부
플러그인이므로 이 목록에 넣지 않는다.

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
| `docs-standards`   | `.agents/skills/docs-standards/`   | `project-skills-bootstrap:bootstrap-docs-standards`   |
| `coding-standards` | `.agents/skills/coding-standards/` | `project-skills-bootstrap:bootstrap-coding-standards` |
| `work-cycle`       | `.agents/skills/work-cycle/`       | `project-skills-bootstrap:bootstrap-work-cycle`       |
| `scratch-dir`      | `.agents/skills/scratch-dir/`      | `project-skills-bootstrap:bootstrap-scratch-dir`      |
```

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
