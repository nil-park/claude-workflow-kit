---
name: bootstrap-git-workflow
description: >-
  git-workflow 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
플랫폼에 맞는 템플릿을 읽어 프로젝트에 설치하거나 업데이트한다.

## 플랫폼 판정

1. `git remote get-url origin`으로 origin URL을 확인한다.
2. URL에 `github.com`이 포함되면 GitHub, `gitlab`이 포함되면 GitLab으로 판정한다.
3. 판정된 플랫폼에 해당하는 템플릿 파일만 읽는다.
   - GitHub: 베이스 디렉터리 아래 `git-workflow-github.md`
   - GitLab: 베이스 디렉터리 아래 `git-workflow-gitlab.md`

## 설치 옵션

아래 옵션을 사용자에게 묻는다. 선택한 옵션은 설치본이 아니라 리포지토리의 `CLAUDE.md`나 `AGENTS.md`에 규칙으로 추가한다.

- PR/MR 제목 앞에 이슈 트래커 키(예: `GAI-123`)를 붙일지 묻는다.
  - 해당하는 키가 없는 변경은 Conventional Commits 형태로 제목을 쓴다는 규칙도 함께 추가한다.
  - GitLab이면 GitLab 이슈 링크를 MR 본문에 넣는다는 규칙도 추가한다.
- GitLab이면 `glab mr create`에 `--squash-before-merge --remove-source-branch`를 붙일지 묻는다.

## 설치

`.agents/skills/git-workflow/SKILL.md`가 없을 때 실행한다.

1. 플랫폼 판정을 실행해 해당 템플릿을 읽는다.
2. 설치 옵션을 확인한다.
3. `.agents/skills/git-workflow/SKILL.md`에 템플릿을 그대로 쓴다.
4. `.claude/skills/git-workflow`에 `.agents/skills/git-workflow`을 가리키는
   심볼릭 링크를 만든다.
5. 선택한 옵션을 `CLAUDE.md`나 `AGENTS.md`에 규칙으로 추가한다.
6. 설치 결과를 보고한다.

## 업데이트

`.agents/skills/git-workflow/SKILL.md`가 이미 있을 때 실행한다.

1. 플랫폼 판정을 실행해 해당 템플릿을 읽는다.
2. `.agents/skills/git-workflow/SKILL.md`(현재 파일)를 읽는다.
3. 두 파일을 비교해 차이를 사용자에게 보고한다.
4. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.

## UPDATE.md 기록

- 베이스 디렉터리를 기준으로 `../bootstrap/UPDATE.md`(템플릿)를 읽는다.
- `.agents/skills/UPDATE.md`가 없으면 템플릿의 내용으로 만든다.
- 파일이 있는데 내용이 템플릿과 다르면 템플릿에 맞춘다. 프로젝트가 덧붙인 내용은 유지한다.
