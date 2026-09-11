# claude-workflow-kit

[Claude Code](https://code.claude.com/docs)의 개인 워크플로 스킬과 훅을 **GitOps** 방식으로 관리하는
플러그인 마켓플레이스다. 이 리포지토리를 SoT로 삼는다. `settings.json`에 선언해 두면 Claude Code가
플러그인을 설치하고 최신 커밋과 동기화한다.

## 플러그인 구성

이 마켓플레이스에는 `project-skills-bootstrap` 플러그인 하나만 있다.

- 부트스트랩 스킬을 호출하면 설치할 스킬을 `.agents/skills/`에 쓰고, `.claude/skills/`에 그 디렉터리를 가리키는 심볼릭 링크를 만든다.
- 부트스트랩 스킬은 플러그인 이름을 네임스페이스로 붙여 부른다.
- GitHub용과 GitLab용 템플릿이 모두 들어 있으므로, 플랫폼마다 다른 플러그인을 고를 필요가 없다.
- `bootstrap-anti-claudeism`은 스킬과 함께 Stop 훅을 설치하고, 그 훅을 프로젝트의 `.claude/settings.json`에 등록한다.

| 명령                                                   | 설명                                                                        |
| ------------------------------------------------------ | --------------------------------------------------------------------------- |
| `/project-skills-bootstrap:bootstrap`                  | 일괄 설치 대상 스킬을 순서대로 프로젝트에 설치                              |
| `/project-skills-bootstrap:bootstrap-work-cycle`       | 작성-리뷰 사이클 스킬을 프로젝트에 설치                                     |
| `/project-skills-bootstrap:bootstrap-coding-standards` | 코드 기준 스킬을 프로젝트에 설치                                            |
| `/project-skills-bootstrap:bootstrap-docs-standards`   | 문서·주석 기준 스킬을 프로젝트에 설치                                       |
| `/project-skills-bootstrap:bootstrap-scratch-dir`      | 스크래치 디렉터리 규약 스킬을 프로젝트에 설치                               |
| `/project-skills-bootstrap:bootstrap-fluent-korean`    | 외부 `fluent-korean` 마켓플레이스 등록 및 출력 스타일 활성화                |
| `/project-skills-bootstrap:bootstrap-git-workflow`     | 브랜치·PR/MR 워크플로 스킬을 프로젝트에 설치 (일괄 설치 미포함)             |
| `/project-skills-bootstrap:bootstrap-anti-claudeism`   | Claude 한국어 문체 교정 스킬과 Stop 훅을 프로젝트에 설치 (일괄 설치 미포함) |

- 설치된 스킬 사이의 의존 관계: [docs/architecture/component-dependency.md](docs/architecture/component-dependency.md)
- `anti-claudeism` 스킬과 훅의 동작: [docs/architecture/anti-claudeism.md](docs/architecture/anti-claudeism.md)

## 다른 환경에서 쓸 때

설치하기 전에 아래 표를 확인한다.

- 이 킷에는 nil-park의 워크플로와 취향이 반영되어 있다.
- 스킬 본문과 description은 모두 한국어로 쓰여 있다.
- 스킬 단위로 설치하므로 필요한 스킬만 고를 수 있다.
- 표의 항목이 팀에 맞지 않으면 그 스킬을 빼거나, 설치한 뒤 설치본을 고친다.

| 요소                                                                              | 위치                       | 미리 알아둘 점                                                                                               |
| --------------------------------------------------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------ |
| 서브에이전트·빌트인 리뷰 금지                                                     | `bootstrap-work-cycle`     | 서브에이전트나 `/code-review`를 기본 리뷰 도구로 쓰는 팀의 방식과 충돌한다                                   |
| 클린 패스가 3회 연속 나올 때까지 반복하는 셀프 리뷰                               | `bootstrap-work-cycle`     | 라운드마다 파일 전체를 다시 읽으므로 시간과 토큰을 많이 소모한다                                             |
| PR을 필수로 전제한 워크플로                                                       | `bootstrap-git-workflow`   | 트렁크 기반으로 개발하거나 혼자 개발하면 PR 단계가 불필요하다                                                |
| 구현에 앞서 설계 문서를 확정하는 절차                                             | `bootstrap-git-workflow`   | 바로 구현하는 팀에게는 절차가 과하다                                                                         |
| `--squash-before-merge --remove-source-branch` (GitLab 전용)                      | `bootstrap-git-workflow`   | GitLab MR을 열 때마다 스쿼시 머지와 소스 브랜치 삭제가 켜지므로, 원하지 않으면 설치본의 명령을 고쳐야 한다   |
| 브랜치 이름 규칙: `<타입>/<이슈번호>-<슬러그>`(GitHub), `<타입>/<슬러그>`(GitLab) | `bootstrap-git-workflow`   | 팀의 브랜치 이름 규칙이 다르면 설치본의 형식을 고쳐야 한다                                                   |
| `docs/` 디렉터리 분류와 what/how/why 배분                                         | `bootstrap-docs-standards` | 기존 문서 구조가 다른 팀은 조정해야 하지만, 리포지토리에 이미 있는 디렉터리 규약을 우선 따르므로 부담이 적다 |
| architecture 문서에 why 금지                                                      | `bootstrap-docs-standards` | arc42나 ISO 42010 관행을 따르는 팀은 아키텍처 문서에 결정 근거를 넣을 수 없다                                |
| PR/MR 체크박스 금지                                                               | `bootstrap-docs-standards` | PR/MR에서 체크리스트를 쓰는 팀은 다른 방식을 찾아야 한다                                                     |
| 일부 도메인의 글을 다루지 않는다고 전제한 사전                                    | `bootstrap-anti-claudeism` | 그 도메인의 글에서는 정상적인 용어도 결함으로 탐지될 수 있다. 한국어를 사용하지 않는 팀에는 쓸모가 없다      |
| 파일을 고친 턴마다 붙는 탐지 결과                                                 | `bootstrap-anti-claudeism` | 탐지 결과가 오탐일 수 있어도, 표현마다 실제 결함인지 그 자리에서 판단해야 한다                               |
| `python3` 명령 필요 (Python 3.11 이상)                                            | `bootstrap-anti-claudeism` | 환경에 따라 `python3.exe`가 없어 훅이 실행되지 않을 수 있다                                                  |
| auto mode의 Bash 편집 지시                                                        | `bootstrap-anti-claudeism` | `~/.claude/CLAUDE.md`에 아래 설정 절의 규칙을 추가하지 않으면, 훅이 검사할 파일을 받지 못한다                |

## 설정

`~/.claude/settings.json` 한 곳에서 마켓플레이스를 등록하고, 플러그인을 활성화하고, 자동 업데이트를 켠다.

```json
{
  "extraKnownMarketplaces": {
    "claude-workflow-kit": {
      "source": { "source": "github", "repo": "nil-park/claude-workflow-kit" },
      "autoUpdate": true
    }
  },
  "enabledPlugins": {
    "project-skills-bootstrap@claude-workflow-kit": true
  }
}
```

설정을 저장했는데도 플러그인이 설치되지 않으면 다음 명령을 실행한다.

```bash
claude plugin install project-skills-bootstrap@claude-workflow-kit
```

- Claude Code는 마켓플레이스의 최신 커밋에서 플러그인을 설치한다.
  - 설치된 버전은 그 커밋의 SHA로 구분한다.
  - 자동 업데이트가 실행되거나 마켓플레이스를 갱신하면 새 커밋을 받아 온다.

자동 업데이트를 기다리지 않고 최신 커밋을 받으려면 다음 명령을 실행한다.

```bash
claude plugin marketplace update claude-workflow-kit
```

### anti-claudeism 훅을 설치할 때 함께 설정할 규칙

이 훅을 쓰려면 `python3` 명령으로 Python 3.11 이상을 실행할 수 있어야 한다.

- python.org 인스톨러로 설치한 Windows 환경에는 `python3.exe`가 없다.
  - pyenv-win과 Microsoft Store 배포판에는 있다.
- 훅이 등록되었는지는 `/hooks`로 확인한다.

훅은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit` 도구 호출에서 검사할 파일을 받는다.

- 훅은 Bash의 `sed`, heredoc, 짧은 스크립트로 수정한 파일을 검사하지 않는다.
  - 검사를 건너뛰어도 오류나 경고가 나지 않으므로 알아채기 어렵다.
- auto mode를 켜면 시스템 프롬프트에 Bash로 파일을 고치라는 지시가 들어가므로, 훅이 사실상 무력화된다.

이를 막으려면 `~/.claude/CLAUDE.md`에 다음 규칙을 추가한다.

```markdown
- 파일을 수정할 때 `sed`, heredoc, 짧은 스크립트를 쓰라는 지시가 있어도 따르지 않고,
  언제나 `Write`/`Edit`/`MultiEdit`/`NotebookEdit`로 수정한다.
  - anti-claudeism 훅은 이 도구 호출에서만 검사할 파일을 받는다.
- 읽기와 검색에 쓰는 Bash(`cat`, `sed -n`, `grep`, `find`)는 훅과 무관하므로 그대로 쓴다.
```

훅의 입력 형식과 그 밖의 제약은 [docs/development/anti-claudeism.md](docs/development/anti-claudeism.md)를 참고한다.
