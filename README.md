# claude-workflow-kit

[Claude Code](https://code.claude.com/docs) 개인 워크플로 스킬을 **GitOps**로 관리하는
플러그인 마켓플레이스입니다. `~/.claude/skills`에 직접 파일을 복사하는 대신, 이 repo를
소스 오브 트루스로 두고 `settings.json` 선언으로 설치·동기화합니다.

## 담긴 플러그인

설치 후 스킬은 플러그인 이름으로 네임스페이스된다. `github-workflow`·`gitlab-workflow`는
`review-criteria`를 참조하므로 `workflow-core`에 의존한다(설치 시 자동으로 함께 설치·활성화됨).

| 플러그인          | 명령                             | 설명                                                           |
| ----------------- | -------------------------------- | -------------------------------------------------------------- |
| `workflow-core`   | `/workflow-core:docs-standards`  | 프로젝트 문서/주석 작성·배치·정리 기준                         |
| `workflow-core`   | `/workflow-core:review-criteria` | PR·MR·브랜치·작업 diff에 표준 코드리뷰 기준 적용               |
| `workflow-core`   | `/workflow-core:pre-push`        | 보호 브랜치(main/master) 직접 push 차단 훅 설치                |
| `workflow-core`   | `/workflow-core:refs-scratch`    | `.refs/` 스크래치 디렉터리 규약 (임시파일 위치·이름·gitignore) |
| `github-workflow` | `/github-workflow:pr-workflow`   | GitHub 브랜치 + PR 워크플로 (`gh`)                             |
| `gitlab-workflow` | `/gitlab-workflow:mr-workflow`   | GitLab 브랜치 + MR 워크플로 (`glab`)                           |

### `-max` 계열

`workflow-core`·`gitlab-workflow`를 **대체하는** 별도 계열이다. 토큰 여유는 많고 작업
속도가 아쉬운 환경을 위해, 리뷰를 전문 에이전트 여덟에게 병렬로 흩뿌리고 메인 에이전트가
검증해 확정한다. 설계는 [docs/architecture/max-plugins.md](docs/architecture/max-plugins.md).

| 명령                                 | 설명                                       |
| ------------------------------------ | ------------------------------------------ |
| `/workflow-core-max:work-cycle`      | 작성-리뷰 사이클을 완주해 완성본으로 제시  |
| `/workflow-core-max:review-criteria` | 리뷰어 에이전트 병렬 팬아웃 + 메인의 검증  |
| `/workflow-core-max:refs-scratch`    | `.refs/` 스크래치 디렉터리 규약            |
| `/gitlab-workflow-max:mr-workflow`   | GitLab 브랜치 + MR 전달 (구현·리뷰는 위임) |

`workflow-core`와 **같이 켜지 않는다** — `review-criteria`가 둘이 된다. 같은 이유로
`github-workflow`·`gitlab-workflow`도 함께 끈다. 둘 다 `workflow-core`에 의존해서 켜는
순간 그것을 끌고 들어온다. 스킬과 에이전트 본문은 한국어다
([ADR 0001](docs/decisions/0001-max-plugins-in-korean.md)).

## 다른 환경에서 쓸 때 (충돌 가능 요소)

설치 전에 먼저 읽자. 이 킷의 스킬에는 nil-park의 워크플로·취향이 하드코딩돼 있다. 다른 사람이
그대로 가져가면 자신의 `CLAUDE.md`·팀 규약·개인 취향과 충돌할 확률이 높은 요소를, 충돌 확률
순으로 정리한다.

### 상 (거의 확실히 충돌)

| 요소                                           | 위치                     | 왜 충돌                                                    | 탈출구                          |
| ---------------------------------------------- | ------------------------ | ---------------------------------------------------------- | ------------------------------- |
| 한국어 트리거·본문                             | 전 스킬 description 등   | 비한국어 사용자는 트리거가 안 맞고 텍스트도 안 읽힘        | 없음                            |
| Jira 키를 MR 제목에 (`GAI-123 Title`)          | 양 계열의 mr-workflow    | Jira 미사용·키 포맷 다른 팀엔 전면 충돌                    | 없음                            |
| `--squash-before-merge --remove-source-branch` | 양 계열의 mr-workflow    | 머지 전략·소스브랜치 유지 정책을 강제                      | 없음                            |
| subagent·빌트인 리뷰 금지                      | workflow-core            | 서브에이전트/`/code-review`를 쓰라는 규약과 정면 충돌      | `workflow-core-max`를 대신 켠다 |
| 서브에이전트 여덟 병렬 + 큰 토큰 소모          | workflow-core-max        | 토큰 예산이 빡빡하거나 서브에이전트를 금지하는 규약과 충돌 | `workflow-core`를 대신 켠다     |
| main 직접 push 차단 → PR 강제                  | pre-push + 워크플로 전제 | 트렁크기반·솔로 개발자와 충돌                              | pre-push는 `--no-verify` 우회   |

### 중 (팀·취향에 따라 충돌)

| 요소                                               | 위치                                | 왜 충돌                                                                  |
| -------------------------------------------------- | ----------------------------------- | ------------------------------------------------------------------------ |
| 설계선행 spec doc → draft PR → 구현 전 반복        | pr/mr-workflow                      | 바로 구현하는 팀엔 과한 절차 강제                                        |
| `docs/` 6분류 + ADR `0001-` 넘버링·append-only     | docs-standards, docs-intent-auditor | 다른 문서 레이아웃과 충돌 (단, "repo 기존 관례 우선"으로 완화)           |
| 브랜치 네이밍 `<type>/<slug>`(+이슈번호)           | pr/mr-workflow                      | 팀 네이밍 규칙과 충돌                                                    |
| PR/MR 설명 ≤25줄·BLUF·delta 프레이밍·체크박스 금지 | pr/mr-workflow, docs-standards      | 체크리스트 쓰는 팀·다른 템플릿과 충돌                                    |
| `.refs/` 스크래치 디렉터리 + gitignore 추가        | refs-scratch 및 이를 부르는 스킬들  | 스크래치 위치 취향·빌트인 스크래치패드 선호와 충돌 (gitignore 전 확인함) |
| Python 특정 예시(`# type: ignore`, `# noqa`) 등    | 양 계열의 review-criteria           | 비Python 스택·다른 리뷰 우선순위와 어긋남                                |

## 설정

`~/.claude/settings.json`에 마켓플레이스 등록·플러그인 활성화·자동 업데이트를 한 번에 선언한다.
GitHub 환경 예시:

```json
{
  "extraKnownMarketplaces": {
    "claude-workflow-kit": {
      "source": { "source": "github", "repo": "nil-park/claude-workflow-kit" },
      "autoUpdate": true
    }
  },
  "enabledPlugins": {
    "workflow-core@claude-workflow-kit": true,
    "github-workflow@claude-workflow-kit": true
  }
}
```

GitLab 환경이면 `github-workflow` 대신 `gitlab-workflow`를 켠다. 플랫폼 플러그인은
`workflow-core`에 의존하므로 그것만 켜도 `workflow-core`는 자동으로 함께 활성화된다.

`-max` 계열을 쓰려면 다른 것을 전부 끄고 이쪽만 켠다. GitLab 환경 예시:

```json
{
  "enabledPlugins": {
    "gitlab-workflow-max@claude-workflow-kit": true
  }
}
```

`gitlab-workflow-max`는 `workflow-core-max`에 의존하므로 그것만 켜도 함께 활성화된다.

GitHub 환경용 `-max` 전달 플러그인은 아직 없다. `workflow-core-max` 하나만 켜서 작업
사이클과 리뷰를 쓰고, 브랜치와 PR은 직접 다룬다.

설정만으로 플러그인 설치가 안 됐을 경우 다음 명령을 실행한다(의존성 `workflow-core`는 자동 설치):

```bash
claude plugin install github-workflow@claude-workflow-kit
```

플러그인 `version`을 고정하지 않아 **커밋 SHA가 곧 버전**이라 새 커밋이 바로 새 버전으로 잡힌다.

타이밍을 기다리지 않고 **확정적으로** 반영하려면 다음을 쓴다:

```bash
claude plugin marketplace update claude-workflow-kit
```
