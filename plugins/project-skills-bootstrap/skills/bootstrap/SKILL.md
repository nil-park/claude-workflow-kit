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
# 스킬 업데이트

## 업스트림

- 리포지토리: https://github.com/nil-park/claude-workflow-kit
- 경로: `plugins/project-skills-bootstrap/skills/`
- 업스트림의 `bootstrap-<스킬명>/SKILL.md`는 설치·업데이트 로더이고, 실제 템플릿은 같은 디렉터리에 있는 아래 표의 마크다운 파일이다.

## 대상 스킬

| 로컬 경로                   | 업스트림 템플릿                                  |
| --------------------------- | ------------------------------------------------ |
| `coding-standards/SKILL.md` | `bootstrap-coding-standards/coding-standards.md` |
| `docs-standards/SKILL.md`   | `bootstrap-docs-standards/docs-standards.md`     |
| `work-cycle/SKILL.md`       | `bootstrap-work-cycle/work-cycle.md`             |
| `scratch-dir/SKILL.md`      | `bootstrap-scratch-dir/scratch-dir.md`           |

## 업데이트 절차

- `project-skills-bootstrap:bootstrap-<스킬명>` 스킬을 실행하면, 이미 설치된 스킬은 업스트림 템플릿과 비교하는 업데이트 흐름으로 진행된다.
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
