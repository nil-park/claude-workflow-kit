---
name: bootstrap
description: >-
  docs-standards, coding-standards, work-cycle, fluent-korean, scratch-dir 스킬을
  현재 프로젝트에 한 번에 설치하거나, 이 플러그인으로 설치한 스킬을 한 번에 업데이트하고
  싶을 때 부른다. anti-claudeism, git-workflow, multi-session은 새로 설치하지 않고,
  설치되어 있을 때 업데이트만 한다.
---

## 대상 스킬

| 스킬               | 설치 여부를 확인할 곳                                                     | 부트스트랩 스킬                                       | 신규 설치 |
| ------------------ | ------------------------------------------------------------------------- | ----------------------------------------------------- | --------- |
| `docs-standards`   | `.agents/skills/docs-standards/SKILL.md`                                  | `project-skills-bootstrap:bootstrap-docs-standards`   | 포함      |
| `coding-standards` | `.agents/skills/coding-standards/SKILL.md`                                | `project-skills-bootstrap:bootstrap-coding-standards` | 포함      |
| `work-cycle`       | `.agents/skills/work-cycle/SKILL.md`                                      | `project-skills-bootstrap:bootstrap-work-cycle`       | 포함      |
| `fluent-korean`    | `.claude/settings.json`의 `enabledPlugins["fluent-korean@fluent-korean"]` | `project-skills-bootstrap:bootstrap-fluent-korean`    | 포함      |
| `scratch-dir`      | `.agents/skills/scratch-dir/SKILL.md`                                     | `project-skills-bootstrap:bootstrap-scratch-dir`      | 포함      |
| `anti-claudeism`   | `.agents/skills/anti-claudeism/SKILL.md`                                  | `project-skills-bootstrap:bootstrap-anti-claudeism`   | 제외      |
| `git-workflow`     | `.agents/skills/git-workflow/SKILL.md`                                    | `project-skills-bootstrap:bootstrap-git-workflow`     | 제외      |
| `multi-session`    | `.claude/skills/multi-session/SKILL.md`                                   | `project-skills-bootstrap:bootstrap-multi-session`    | 제외      |

- 신규 설치에서 제외한 스킬은 사용자가 명시적으로 요청할 때에만 해당 부트스트랩 스킬을 불러 설치한다.

## 설치 여부 판정

가장 먼저 위 표의 여덟 스킬이 각각 설치되어 있는지 확인한다.

- 하나도 설치되어 있지 않으면 신규 설치 절차를 실행한다.
- 하나라도 설치되어 있으면 업데이트 절차를 실행한다.

## 신규 설치

1. 신규 설치에 포함한 다섯 스킬의 부트스트랩 스킬을 표의 순서대로 실행한다.
2. 결과를 보고할 때 설치되지 않은 스킬을 안내한다.

## 업데이트

1. 설치된 스킬의 부트스트랩 스킬만 표의 순서대로 불러 업데이트 절차를 실행한다.
   - `fluent-korean`은 마켓플레이스의 `autoUpdate`로 갱신되므로 부르지 않는다.
   - 설치되지 않은 스킬은 신규 설치 포함 여부와 상관없이 설치하지 않는다.
2. 결과를 보고할 때 설치되지 않은 스킬을 안내한다.

## 설치되지 않은 스킬 안내

- 설치되지 않은 스킬의 이름을 표의 순서대로 쉼표로 이어 아래 문구로 안내한다.
  - 한 스킬일 때: `참고: <스킬 이름> 스킬은 설치되어 있지 않습니다.`
  - 여러 스킬일 때: `참고: <스킬 이름 목록> <개수> 스킬은 설치되어 있지 않습니다.`
    - 개수는 "두, 세, 네, 다섯, 여섯, 일곱"으로 적는다.
    - 예: `참고: anti-claudeism, git-workflow 두 스킬은 설치되어 있지 않습니다.`
- 모든 스킬이 설치되어 있으면 안내하지 않는다.

## UPDATE.md 기록

신규 설치나 업데이트 절차를 마친 뒤에 `.agents/skills/UPDATE.md`를 확인한다.

- 각 부트스트랩 스킬이 이 파일을 만들거나 최신 내용으로 맞추므로, 대개 손댈 것이 없다.
- 위 표에서 설치된 것이 `fluent-korean`뿐이면 이 파일을 만들지 않는다.
- 베이스 디렉터리 아래 `UPDATE.md`(템플릿)를 읽는다.
- 파일이 없으면 템플릿의 내용으로 만든다.
- 파일이 있는데 내용이 템플릿과 다르면 템플릿에 맞춘다. 프로젝트가 덧붙인 내용은 유지한다.

## 사용자 가이드

- `docs-standards`: 리포지토리 안의 문서를 어떤 위치에 어떤 기준으로 작성할지 정하고, 주석·이슈·PR/MR 본문과 리뷰에도 적용한다.
- `coding-standards`: 코드 작성과 리뷰에서 따를 공통 기준을 정한다.
- `work-cycle`: 문서를 작성하거나 구현한 뒤 `docs-standards`와 `coding-standards`를 기준으로 리뷰 라운드를 세 번 연속 지적이 없을 때까지 반복한다.
- `fluent-korean`: 한국어 출력의 문장 구성과 표현 규칙을 정한다.
- `scratch-dir`: 커밋하지 않는 임시 작업 파일의 위치와 이름을 정한다.
- `anti-claudeism`: Opus와 Sonnet 5의 한국어 문체 결함 교정 기준을 정하고, 낱말 단위의 결함을 Stop 훅으로 탐지한다.
- `git-workflow`: 브랜치 생성과 GitHub 또는 GitLab의 PR/MR 작업 절차를 정한다.
- `multi-session`: 여러 Claude Code 세션에 설계, 구현, 한국어 검수를 나눠 맡기는 절차를 정한다.
