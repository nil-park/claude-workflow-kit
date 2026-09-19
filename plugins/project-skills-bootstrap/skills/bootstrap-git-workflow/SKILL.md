---
name: bootstrap-git-workflow
description: >-
  git-workflow 스킬을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
플랫폼에 맞는 템플릿을 읽어 프로젝트에 설치하거나 업데이트한다.

## 플랫폼 판정

설치와 업데이트 모두, 템플릿을 읽기 전에 이 단계를 실행한다.

1. `git remote get-url origin`으로 origin URL을 확인한다.
2. URL에 `github.com`이 포함되면 GitHub, `gitlab`이 포함되면 GitLab으로 판정한다.
3. 판정된 플랫폼에 해당하는 템플릿 파일만 읽는다.
   - GitHub: 베이스 디렉터리 아래 `git-workflow-github.md`
   - GitLab: 베이스 디렉터리 아래 `git-workflow-gitlab.md`

## 설치 옵션

플랫폼과 관계없이, 파일을 쓰기 전에 아래 옵션을 사용자에게 확인한다.

- **PR/MR 제목 prefix 여부**: PR/MR 제목 앞에 이슈 트래커 키(Jira 키 등)를 붙일지 사용자에게 묻는다.
- 붙이기로 하면 키의 형식을 확인한다(예: `GAI-123`).
- 그다음 설치할 파일의 `# GitHub PR` 또는 `# GitLab MR` 절에 다음 항목들을 추가한다.
  - `<키>`는 확인한 형식으로 바꾼다.
  - GitLab이면 `PR`을 `MR`로, `gh pr create`를 `glab mr create`로 바꾼다.

  > - **`<키>` 형식의 이슈 트래커 키는 PR 제목 앞에 붙인다.**
  >   - `gh pr create`의 `-t` 값도 `"<키> 제목"` 형태로 쓴다.
  > - 변경 사항에 해당하는 키가 없으면 키를 빼고 Conventional Commits 형태로 제목을 쓴다
  >   - 예시: `fix: 배포 스크립트 경로 보정`

- GitLab이면 다음 항목도 함께 추가한다.

  > - GitLab 이슈 링크는 MR 본문에 넣는다.

- 붙이지 않기로 하면 템플릿 그대로 사용한다.

## 설치

`.agents/skills/git-workflow/SKILL.md`가 없을 때 실행한다.

1. 플랫폼 판정을 실행해 해당 템플릿을 읽는다.
2. 설치 옵션을 확인한다.
3. `.agents/skills/git-workflow/SKILL.md`에 내용을 쓴다.
4. `.claude/skills/git-workflow`에 `.agents/skills/git-workflow`을 가리키는
   심볼릭 링크를 만든다.
5. 완료 후 설치 결과(플랫폼 및 선택한 옵션 포함)를 보고한다.

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
