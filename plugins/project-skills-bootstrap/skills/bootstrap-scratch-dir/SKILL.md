---
name: bootstrap-scratch-dir
description: >-
  scratch-dir 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
`scratch-dir.md`를 읽어 프로젝트에 설치하거나 업데이트한다.

## 스크래치 디렉터리 결정

설치와 업데이트 모두, 실제 파일을 쓰기 전에 이 단계를 먼저 실행한다.

1. `.gitignore`를 읽어 이미 gitignore된 디렉터리 목록을 파악한다.
   빌드 산출물(`dist/`, `build/` 등)과 의존성 디렉터리(`node_modules/` 등)는 제외한다.
2. 후보가 있으면 사용자에게 제시하고 그 중 하나를 선택하거나 새 이름을 입력하도록 한다.
   후보가 없으면 새 이름을 협의한다(추천: `.refs`).
3. 새 이름을 쓰기로 결정했다면 `.gitignore`에 등록한다.
4. 확정된 디렉터리 이름을 이후 단계에서 `{{scratch_dir}}` 치환에 사용한다.

## 설치

`.agents/skills/scratch-dir/SKILL.md`가 없을 때 실행한다.

1. 베이스 디렉터리 아래 `scratch-dir.md`를 읽고, `{{scratch_dir}}`를 확정된 디렉터리 이름으로 치환한다.
2. `.agents/skills/scratch-dir/SKILL.md`에 내용을 쓴다.
3. `.claude/skills/scratch-dir`에 `.agents/skills/scratch-dir`을 가리키는
   심볼릭 링크를 만든다.
4. 완료 후 설치 결과를 보고한다.

## 업데이트

`.agents/skills/scratch-dir/SKILL.md`가 이미 있을 때 실행한다.

1. 베이스 디렉터리 아래 `scratch-dir.md`(템플릿)를 읽고, `{{scratch_dir}}`를 확정된 디렉터리 이름으로 치환한다.
2. `.agents/skills/scratch-dir/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- `.agents/skills/UPDATE.md`가 없으면 아래 내용으로 만든다.
- 파일은 있으나 목록에 `scratch-dir` 항목이 없으면 추가한다.

```markdown
# 설치된 스킬의 출처와 업데이트 방법

이 디렉터리의 스킬은 아래 마켓플레이스에서 제공하는
`project-skills-bootstrap` 플러그인이 복사해 설치한 버전이다. 원본의 변경은 자동으로
반영되지 않는다. 변경 사항을 반영하려면 아래 절차를 실행한다.

- 마켓플레이스: `claude-workflow-kit` (<https://github.com/nil-park/claude-workflow-kit>)
- 플러그인: `project-skills-bootstrap`

원본의 변경 사항을 반영할 때에는 해당 스킬의 부트스트랩 스킬을 다시 부른다. 설치본이
있으면 에이전트는 업데이트 절차에 따라 템플릿과 설치본의 차이를 보고한다. 반영할 변경
사항은 사용자와 상의해서 정하며, 프로젝트별 변경 사항은 유지할 수 있다.

| 설치된 스킬   | 설치 위치                     | 다시 부를 부트스트랩 스킬                        |
| ------------- | ----------------------------- | ------------------------------------------------ |
| `scratch-dir` | `.agents/skills/scratch-dir/` | `project-skills-bootstrap:bootstrap-scratch-dir` |
```
