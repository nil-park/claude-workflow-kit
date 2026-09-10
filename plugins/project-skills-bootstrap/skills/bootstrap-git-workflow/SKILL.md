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

템플릿을 그대로 사용한다.

### GitLab

- **MR 제목 prefix 여부**: MR 제목 앞에 이슈 트래커 키(Jira 키 등)를 붙일지 사용자에게 묻는다.
  - 붙이기로 하면 어떤 형식인지도 확인한다(예: `GAI-123`).
  - 포함하기로 하면, 설치할 파일의 `# GitLab MR` 절에 다음 항목들을 추가한다.

    > - **`<키>` 형식의 이슈 트래커 키는 MR 제목 앞에만 넣는다.**
    >   - `glab mr create`의 `-t` 값도 `"<키> 제목"` 형태로 쓴다.
    > - 변경 사항에 해당하는 키가 없으면 키를 빼고 Conventional Commits 형태로 제목을 쓴다
    >   - 예시: `fix: 배포 스크립트 경로 보정`
    > - GitLab 이슈 링크는 MR 본문에 넣는다.

- 포함하지 않기로 하면 템플릿 그대로 사용한다.

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

## UPDATE.md 기록

- `.agents/skills/UPDATE.md`가 없으면 아래 내용으로 만든다.
- 파일은 있으나 목록에 `git-workflow` 항목이 없으면 추가한다.

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

| 설치된 스킬    | 설치 위치                      | 다시 부를 부트스트랩 스킬                         |
| -------------- | ------------------------------ | ------------------------------------------------- |
| `git-workflow` | `.agents/skills/git-workflow/` | `project-skills-bootstrap:bootstrap-git-workflow` |
```
