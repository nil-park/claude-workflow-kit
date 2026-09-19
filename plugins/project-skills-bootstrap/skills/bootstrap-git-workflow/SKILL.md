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

아래 옵션을 사용자에게 확인한다.

- 설치본에는 옵션과 관계없이 템플릿을 그대로 쓴다.
  - 템플릿은 리포지토리의 `CLAUDE.md`나 `AGENTS.md`에 적힌 규칙을 우선하도록 쓰여 있다.
- 선택한 옵션은 두 파일 중 리포지토리가 에이전트 규칙을 적어 두는 쪽에 규칙으로 추가한다.
  - 둘 다 없거나 어느 쪽인지 판단할 수 없으면 사용자에게 묻는다.
  - 파일이 한국어가 아닌 언어로 쓰여 있으면 규칙을 그 언어로 옮겨 쓴다.

### PR/MR 제목 prefix 여부 (두 플랫폼 공통)

- PR/MR 제목 앞에 이슈 트래커 키(Jira 키 등)를 붙일지 사용자에게 묻는다.
- 붙이기로 하면 키의 형식을 확인한다(예: `GAI-123`).
- 그다음 아래 항목들을 설치본이 아니라 `CLAUDE.md`나 `AGENTS.md`에 추가한다.
  - `<키>`는 확인한 형식으로 바꾼다.
  - GitLab이면 `PR`을 `MR`로, `gh pr create`를 `glab mr create`로 바꾼다.

  > - **`<키>` 형식의 이슈 트래커 키는 PR 제목 앞에 붙인다.**
  >   - `gh pr create`의 `-t` 값도 `"<키> 제목"` 형태로 쓴다.
  > - 변경 사항에 해당하는 키가 없으면 키를 빼고 Conventional Commits 형태로 제목을 쓴다
  >   - 예시: `fix: 배포 스크립트 경로 보정`

- GitLab이면 다음 항목도 함께 추가한다.

  > - GitLab 이슈 링크는 MR 본문에 넣는다.

- 붙이지 않기로 하면 아무것도 추가하지 않는다.

### 스쿼시 머지와 소스 브랜치 삭제 여부 (GitLab 전용)

- MR을 열 때 `--squash-before-merge`와 `--remove-source-branch`를 붙일지 사용자에게 묻는다.
  - 두 플래그는 프로젝트의 머지 설정을 MR마다 덮어쓴다.
  - 붙이지 않으면 프로젝트의 머지 설정대로 머지된다.
- 붙이기로 하면 다음 항목을 설치본이 아니라 `CLAUDE.md`나 `AGENTS.md`에 추가한다.

  > - `glab mr create`에 `--squash-before-merge --remove-source-branch`를 붙인다.

- 붙이지 않기로 하면 아무것도 추가하지 않는다.

## 설치

`.agents/skills/git-workflow/SKILL.md`가 없을 때 실행한다.

1. 플랫폼 판정을 실행해 해당 템플릿을 읽는다.
2. 설치 옵션을 확인한다.
3. `.agents/skills/git-workflow/SKILL.md`에 템플릿을 그대로 쓴다.
4. `.claude/skills/git-workflow`에 `.agents/skills/git-workflow`을 가리키는
   심볼릭 링크를 만든다.
5. 선택한 옵션을 `CLAUDE.md`나 `AGENTS.md`에 규칙으로 추가한다.
6. 완료 후 설치 결과(플랫폼, 선택한 옵션, 규칙을 추가한 파일 포함)를 보고한다.

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
