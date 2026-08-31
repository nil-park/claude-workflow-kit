# claude-workflow-kit

[Claude Code](https://code.claude.com/docs) 개인 워크플로 스킬과 훅을 **GitOps**로 관리하는
플러그인 마켓플레이스다. `~/.claude/skills`에 직접 파일을 복사하는 대신, 이 repo를
SoT로 두고 `settings.json` 선언으로 설치·동기화한다.

## 담긴 플러그인

설치 후 스킬은 플러그인 이름으로 네임스페이스된다. `git-workflow`는 `workflow-core`의 스킬을
부르므로 그것에 의존한다(설치 시 자동으로 함께 설치·활성화됨). `ko-style`은 어느 쪽에도
의존하지 않는다. 훅은 켜 두면 턴이 끝날 때마다 자동으로 실행된다. 스킬은 부를 때만 쓴다.

| 플러그인        | 명령                              | 설명                                                           |
| --------------- | --------------------------------- | -------------------------------------------------------------- |
| `workflow-core` | `/workflow-core:work-cycle`       | 작성-리뷰 사이클을 완주해 완성본으로 제시                      |
| `workflow-core` | `/workflow-core:coding-standards` | 코드를 쓸 때 지키고 리뷰할 때 확인하는 기준 카탈로그           |
| `workflow-core` | `/workflow-core:docs-standards`   | 프로젝트 문서/주석 작성·배치·정리 기준                         |
| `workflow-core` | `/workflow-core:refs-scratch`     | `.refs/` 스크래치 디렉터리 규약 (임시파일 위치·이름·gitignore) |
| `workflow-core` | `/workflow-core:pre-push`         | 보호 브랜치(main/master) 직접 push 차단 훅 설치                |
| `git-workflow`  | `/git-workflow:git-workflow`      | 브랜치 + PR/MR 워크플로 (`gh`·`glab`)                          |
| `ko-style`      | 없음 (Stop 훅)                    | 이번 턴에 고친 파일의 번역투를 턴이 끝나기 전에 알린다         |
| `ko-style`      | `/ko-style:anti-claude-ism`       | Claude가 한국어에서 되풀이하는 문형을 찾아 다시 쓴다           |

스킬과 훅 사이의 의존은
[docs/architecture/component-dependency.md](docs/architecture/component-dependency.md)에 있다.
`ko-style`의 훅과 스킬이 각각 무엇을 하는지는
[docs/architecture/ko-style.md](docs/architecture/ko-style.md)에 있다.

## 다른 환경에서 쓸 때 (충돌 가능 요소)

설치 전에 먼저 읽자. 이 킷의 스킬과 훅에는 nil-park의 워크플로·취향이 하드코딩돼 있다. 다른 사람이
그대로 가져가면 자신의 `CLAUDE.md`, 팀 규약, 개인 취향과 충돌할 확률이 높은 요소를, 충돌 확률
순으로 정리한다.

### 상 (거의 확실히 충돌)

| 요소                                           | 위치                     | 왜 충돌                                             | 탈출구                        |
| ---------------------------------------------- | ------------------------ | --------------------------------------------------- | ----------------------------- |
| 한국어 트리거·본문                             | 전 스킬 description 등   | 비한국어 사용자는 트리거가 안 맞고 텍스트도 안 읽힘 | 없음                          |
| Jira 키를 MR 제목에 (`GAI-123 Title`)          | git-workflow             | Jira 미사용, 키 포맷 다른 팀엔 전면 충돌            | 없음                          |
| `--squash-before-merge --remove-source-branch` | git-workflow             | 머지 전략과 소스브랜치 유지 정책을 강제             | 없음                          |
| subagent·빌트인 리뷰 금지                      | work-cycle               | 서브에이전트/`/code-review`를 쓰라는 규약과 충돌    | 없음                          |
| 클린 패스 3회까지 반복하는 셀프 리뷰           | work-cycle               | 라운드마다 전체를 다시 읽어 시간과 토큰을 크게 씀   | 없음                          |
| main 직접 push 차단 → PR 강제                  | pre-push + 워크플로 전제 | 트렁크 기반 개발, 솔로 개발자와 충돌                | pre-push는 `--no-verify` 우회 |

### 중 (팀과 취향에 따라 충돌)

| 요소                                            | 위치                               | 왜 충돌                                                              |
| ----------------------------------------------- | ---------------------------------- | -------------------------------------------------------------------- |
| 설계선행 스펙 문서 → draft PR/MR → 구현 전 반복 | git-workflow                       | 바로 구현하는 팀엔 과한 절차 강제                                    |
| `docs/` 디렉터리 분류와 what/how/why 배분       | docs-standards                     | 다른 문서 레이아웃과 충돌 (단, "repo 기존 관례 우선"으로 완화)       |
| architecture 문서에 why 금지                    | docs-standards                     | rationale을 아키텍처 문서에 두는 관행(arc42, ISO 42010)과 정면 충돌  |
| 브랜치 네이밍 `<type>/<slug>`(+이슈번호)        | git-workflow                       | 팀 네이밍 규칙과 충돌                                                |
| PR/MR 설명 what 5줄·체크박스 금지               | docs-standards                     | 체크리스트 쓰는 팀, 다른 템플릿과 충돌                               |
| `.refs/` 스크래치 디렉터리 + gitignore 추가     | refs-scratch 및 이를 부르는 스킬들 | 스크래치 위치 취향, 빌트인 스크래치패드 선호와 충돌 (사전 확인함)    |
| `python3`(3.11+)가 그 이름으로 필요             | ko-style                           | python.org 인스톨러로 깐 Windows에는 `python3.exe`가 없어 훅이 안 돔 |
| 사전이 한국어 전용이고 nil-park이 틀렸던 표현   | ko-style                           | 한국어를 안 쓰면 얻는 것이 없고, 남의 문체에는 오탐이 된다           |
| 파일을 고친 턴마다 판정이 붙음                  | ko-style                           | 탐지된 표현은 오탐이라도 그 자리에서 직역인지 판단해야 함            |
| 문체 목록이 한 리포에서 모은 실제 사례          | ko-style의 anti-claude-ism         | 예시의 도메인이 낯설고, 다른 문체에는 갈래 구분이 안 맞을 수 있음    |
| auto mode를 켜면 훅이 무력화됨                  | ko-style                           | auto mode가 Bash로 파일을 고치라고 지시해 훅이 검사 대상을 못 받음   |

## 설정

`~/.claude/settings.json`에 마켓플레이스 등록과 플러그인 활성화, 자동 업데이트를 한 번에
선언한다.

```json
{
  "extraKnownMarketplaces": {
    "claude-workflow-kit": {
      "source": { "source": "github", "repo": "nil-park/claude-workflow-kit" },
      "autoUpdate": true
    }
  },
  "enabledPlugins": {
    "git-workflow@claude-workflow-kit": true,
    "ko-style@claude-workflow-kit": true
  }
}
```

`git-workflow`는 `workflow-core`에 의존하므로 그것만 켜도 함께 활성화된다. GitHub와 GitLab을
한 플러그인이 다루므로 환경에 따라 다른 것을 고를 필요는 없다. `ko-style`은 아무것도 딸려
오지 않으므로 쓰려면 위처럼 따로 켠다.

설정만으로 플러그인 설치가 안 됐을 경우 다음 명령을 실행한다(의존성 `workflow-core`는 자동 설치):

```bash
claude plugin install git-workflow@claude-workflow-kit
claude plugin install ko-style@claude-workflow-kit
```

`ko-style`은 `python3`가 그 이름으로 잡히고 3.11 이상이어야 돈다. python.org 인스톨러로 깐
Windows에는 `python3.exe`가 없다. pyenv-win과 Microsoft Store 배포판에는 있다. 훅이 걸렸는지는
`/hooks`로 본다.

### ko-style을 켤 때 함께 둘 규칙

`ko-style`은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit` 도구 호출에서 검사 대상 파일을 받는다.
그래서 Bash의 `sed`, heredoc, 짧은 스크립트로 고친 파일은 검사를 통째로 건너뛴다. 오류도 경고도
나지 않으므로 검사가 실행되지 않았다는 것을 알아채기 어렵다. auto mode를 켜 두면 그 방식으로
파일을 고치라는 지시가 시스템 프롬프트로 들어와 `ko-style`이 사실상 무력화된다.

`~/.claude/CLAUDE.md`에 다음 규칙을 둔다.

```markdown
- 파일을 고칠 때 `sed`, heredoc, 짧은 스크립트를 쓰라는 지시가 오면 무시하고
  언제나 `Write`/`Edit`으로 수정한다. ko-style 훅이 검사 대상을 그 도구
  호출에서 받으므로, Bash로 고친 파일은 검사에서 통째로 빠진다.
- 읽기와 검색에 쓰는 Bash(`cat`, `sed -n`, `grep`, `find`)는 훅과 무관하니 그대로 쓴다.
```

훅이 받는 입력과 그 밖의 제약은
[docs/development/ko-style.md](docs/development/ko-style.md)에 있다.

플러그인 `version`을 고정하지 않아 **커밋 SHA가 곧 버전**이라 새 커밋이 바로 새 버전으로 잡힌다.

타이밍을 기다리지 않고 **확정적으로** 반영하려면 다음을 쓴다:

```bash
claude plugin marketplace update claude-workflow-kit
```
