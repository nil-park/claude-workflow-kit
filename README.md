# claude-workflow-kit

[Claude Code](https://code.claude.com/docs) 개인 워크플로 스킬과 훅을 **GitOps**로 관리하는
플러그인 마켓플레이스다. 이 리포지토리를 SoT로 두고 `settings.json` 선언으로 설치·동기화한다.

## 담긴 플러그인

두 플러그인은 서로 독립적이며 성격이 다르다.

- `project-skills-bootstrap`: 스킬을 호출하면 `.agents/skills/`에 파일을 쓰고 `.claude/skills/`에 심링크를 건다.
  - 설치 후 스킬은 플러그인 이름으로 네임스페이스된다.
  - GitHub와 GitLab 모두를 위한 템플릿을 포함하므로 환경에 따라 다른 것을 고를 필요는 없다.
- `ko-style`: Stop 훅만 제공하며, 켜 두면 턴이 끝날 때마다 자동으로 실행된다.

| 플러그인                   | 명령                                                   | 설명                                                              |
| -------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------- |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap`                  | 일괄 설치 대상 스킬을 순서대로 프로젝트에 설치                    |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-work-cycle`       | 작성-리뷰 사이클 스킬을 프로젝트에 설치                           |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-coding-standards` | 코드 기준 스킬을 프로젝트에 설치                                  |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-docs-standards`   | 문서·주석 기준 스킬을 프로젝트에 설치                             |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-scratch-dir`      | 스크래치 디렉터리 규약 스킬을 프로젝트에 설치                     |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-fluent-korean`    | 외부 `fluent-korean` 마켓플레이스 등록 및 출력 스타일 활성화      |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-git-workflow`     | 브랜치·PR/MR 워크플로 스킬을 프로젝트에 설치 (일괄 설치 미포함)   |
| `project-skills-bootstrap` | `/project-skills-bootstrap:bootstrap-anti-claudeism`   | Claude 한국어 문형 교정 스킬을 프로젝트에 설치 (일괄 설치 미포함) |

부트스트랩 스킬 사이의 의존은
[docs/architecture/component-dependency.md](docs/architecture/component-dependency.md)에 있다.
`ko-style` 훅이 무엇을 하는지는
[docs/architecture/ko-style.md](docs/architecture/ko-style.md)에 있다.

## 다른 환경에서 쓸 때

설치 전에 먼저 읽자. 이 킷에는 nil-park의 워크플로·취향이 반영돼 있고, 스킬 본문과 description은 전부 한국어다.

- `ko-style`은 켜는 순간 고정 동작이 시작되고 끄는 방법이 없다.
- `project-skills-bootstrap`은 스킬 단위로 설치하므로 원하는 것만 고르고, 설치본을 팀이 직접 수정할 수 있다.

### ko-style

| 요소                                                       | 왜 충돌                                                                                                                                             |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| 사전이 한 리포에서 수집된 Claude 문체 결함 사례로만 구성됨 | 한국어를 안 쓰면 얻는 것이 없고, 남의 문체에는 오탐이 된다                                                                                          |
| 파일을 고친 턴마다 판정이 붙음                             | 탐지된 표현은 오탐이라도 그 자리에서 실제로 결함인지 판단해야 한다                                                                                  |
| `python3`(3.11+)가 그 이름으로 필요                        | 환경에 따라 `python3.exe`가 없어 훅이 돌지 않을 수 있다. 대안은 [설정](#설정)에 있다                                                                |
| auto mode를 켜면 훅이 무력화됨                             | auto mode가 Bash로 파일을 고치라고 지시해 훅이 검사 대상을 못 받는다. 상세는 [ko-style을 켤 때 함께 둘 규칙](#ko-style을-켤-때-함께-둘-규칙)에 있다 |

### project-skills-bootstrap

모든 항목의 탈출구는 "설치하지 않거나, 설치 후 고친다"이다.

| 요소                                                         | 위치                       | 미리 알아둘 점                                                                             |
| ------------------------------------------------------------ | -------------------------- | ------------------------------------------------------------------------------------------ |
| subagent·빌트인 리뷰 금지                                    | `bootstrap-work-cycle`     | 서브에이전트나 `/code-review`를 기본 리뷰 도구로 쓰는 팀은 이 규칙에 막힌다                |
| 클린 패스 3회까지 반복하는 셀프 리뷰                         | `bootstrap-work-cycle`     | 라운드마다 전체를 다시 읽어 시간과 토큰을 크게 쓴다                                        |
| PR 필수 워크플로 전제                                        | `bootstrap-git-workflow`   | 트렁크 기반 개발이나 솔로 개발자에게는 PR 단계가 불필요하다                                |
| 설계선행 스펙 문서 → draft PR/MR → 구현 전 반복              | `bootstrap-git-workflow`   | 바로 구현하는 팀에게는 절차가 과하다                                                       |
| `--squash-before-merge --remove-source-branch` (GitLab 전용) | `bootstrap-git-workflow`   | GitLab 리포에서는 스쿼시 머지와 소스브랜치 삭제가 강제되며, GitHub 리포에는 해당 없다      |
| 브랜치 네이밍 `<타입>/<슬러그>`                              | `bootstrap-git-workflow`   | 팀 브랜치 네이밍 규칙이 다르면 형식을 맞춰야 한다                                          |
| `docs/` 디렉터리 분류와 what/how/why 배분                    | `bootstrap-docs-standards` | 기존 문서 레이아웃이 다른 팀에는 조정이 필요하다 (단, repo 기존 관례를 우선 적용해 완화됨) |
| architecture 문서에 why 금지                                 | `bootstrap-docs-standards` | arc42나 ISO 42010 관행을 따르는 팀은 아키텍처 문서에 rationale을 넣지 못한다               |
| PR/MR 체크박스 금지                                          | `bootstrap-docs-standards` | 체크리스트를 PR 프로세스에 활용하는 팀은 다른 방식이 필요하다                              |

## 설정

`~/.claude/settings.json`에 마켓플레이스 등록과 플러그인 활성화, 자동 업데이트를 한 번에 선언한다.

```json
{
  "extraKnownMarketplaces": {
    "claude-workflow-kit": {
      "source": { "source": "github", "repo": "nil-park/claude-workflow-kit" },
      "autoUpdate": true
    }
  },
  "enabledPlugins": {
    "project-skills-bootstrap@claude-workflow-kit": true,
    "ko-style@claude-workflow-kit": true
  }
}
```

필요한 플러그인만 각각 활성화한다.

설정을 저장했는데도 플러그인이 설치되지 않으면 다음 명령을 실행한다:

```bash
claude plugin install project-skills-bootstrap@claude-workflow-kit
claude plugin install ko-style@claude-workflow-kit
```

`ko-style`을 실행하려면 `python3` 명령으로 Python 3.11 이상을 실행할 수 있어야 한다.

- python.org 인스톨러로 깐 Windows에는 `python3.exe`가 없다.
  - pyenv-win과 Microsoft Store 배포판에는 있다.
- 훅이 적용되었는지는 `/hooks`로 확인한다.

- 플러그인은 마켓플레이스의 최신 커밋을 설치하며, 적용된 버전은 해당 커밋 SHA로 식별한다.
  - 자동 업데이트 시점이나 마켓플레이스를 갱신할 때 새 커밋이 반영된다.

자동 업데이트가 수행될 때까지 기다리지 않고 최신 커밋을 반영하려면 다음을 실행한다:

```bash
claude plugin marketplace update claude-workflow-kit
```

### ko-style을 켤 때 함께 설정할 규칙

`ko-style`은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit` 도구 호출에서 검사 대상 파일을 받는다.

- Bash의 `sed`, heredoc, 짧은 스크립트로 수정한 파일은 검사 대상에서 제외된다.
  - 오류도 경고도 나지 않으므로 검사가 실행되지 않았다는 것을 알아채기 어렵다.
- auto mode를 켜 두면 Bash로 파일을 수정하라는 지시가 시스템 프롬프트에 포함되어 `ko-style`이 사실상 무력화된다.

`~/.claude/CLAUDE.md`에 다음 규칙을 설정한다.

```markdown
- 파일을 수정할 때 `sed`, heredoc, 짧은 스크립트를 쓰라는 지시가 오면 무시하고
  언제나 `Write`/`Edit`/`MultiEdit`/`NotebookEdit`로 수정한다.
  ko-style 훅이 검사 대상 파일을 이 도구 호출을 통해 받으므로, Bash로 수정한 파일은 검사 대상에서 제외되기 때문이다.
- 읽기와 검색에 쓰는 Bash(`cat`, `sed -n`, `grep`, `find`)는 훅과 무관하니 그대로 쓴다.
```

훅의 입력 형식과 그 밖의 제약은 [docs/development/ko-style.md](docs/development/ko-style.md)에 있다.
