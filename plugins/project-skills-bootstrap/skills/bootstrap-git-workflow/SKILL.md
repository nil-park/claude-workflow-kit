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

## 플랫폼별 설치 옵션

플랫폼에 따라 아래 옵션을 사용자에게 확인한다. 파일을 쓰기 전에 실행한다.

### GitHub

**PR 제목 이슈 번호 포함 여부**: PR 제목에 `[#47]` 형태로 이슈 번호를 포함할지 묻는다.
포함하기로 하면, 설치할 파일의 `# GitHub PR` 절에 다음 규칙을 추가한다.

> **이슈 번호는 PR 제목에 `[#47]` 형태로 넣는다.** 본문에도 `#47`을 한 번 적어
> 양방향으로 걸리게 한다. `gh pr create`의 `-t` 값도 `"[#47] 제목"` 형태로 쓴다.

포함하지 않기로 하면 템플릿 그대로 사용한다.

### GitLab

**MR 제목 prefix 여부**: MR 제목 앞에 이슈 트래커 키(Jira 키 등)를 붙일지 묻는다.
붙이기로 하면 어떤 형식인지도 확인한다(예: `GAI-123`).
포함하기로 하면, 설치할 파일의 `# GitLab MR` 절에 다음 규칙을 추가한다.

> **`<키>` 형식의 이슈 트래커 키는 MR 제목 앞에만 넣는다.** 이슈 링크는 MR 본문에 넣는다.
> `glab mr create`의 `-t` 값도 `"<키> 제목"` 형태로 쓴다.

포함하지 않기로 하면 템플릿 그대로 사용한다.

## 설치

`.agents/skills/git-workflow/SKILL.md`가 없을 때 실행한다.

1. 플랫폼 판정을 실행해 해당 템플릿을 읽는다.
2. 플랫폼별 설치 옵션을 확인한다.
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
