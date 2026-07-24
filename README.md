# claude-workflow-kit

[Claude Code](https://code.claude.com/docs) 개인 워크플로 스킬을 **GitOps**로 관리하는
플러그인 마켓플레이스입니다. `~/.claude/skills`에 직접 파일을 복사하는 대신, 이 repo를
소스 오브 트루스로 두고 `settings.json` 선언으로 설치·동기화합니다.

## 담긴 플러그인

설치 후 스킬은 플러그인 이름으로 네임스페이스된다. `github-workflow`·`gitlab-workflow`는
`review-criteria`를 참조하므로 `workflow-core`에 의존한다(설치 시 자동으로 함께 설치·활성화됨).

| 플러그인          | 명령                             | 설명                                             |
| ----------------- | -------------------------------- | ------------------------------------------------ |
| `workflow-core`   | `/workflow-core:docs-standards`  | 프로젝트 문서/주석 작성·배치·정리 기준           |
| `workflow-core`   | `/workflow-core:review-criteria` | PR·MR·브랜치·작업 diff에 표준 코드리뷰 기준 적용 |
| `workflow-core`   | `/workflow-core:pre-push`        | 보호 브랜치(main/master) 직접 push 차단 훅 설치  |
| `github-workflow` | `/github-workflow:pr-workflow`   | GitHub 브랜치 + PR 워크플로 (`gh`)               |
| `gitlab-workflow` | `/gitlab-workflow:mr-workflow`   | GitLab 브랜치 + MR 워크플로 (`glab`)             |

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

설정만으로 플러그인 설치가 안 됐을 경우 다음 명령을 실행한다(의존성 `workflow-core`는 자동 설치):

```bash
claude plugin install github-workflow@claude-workflow-kit
```

플러그인 `version`을 고정하지 않아 **커밋 SHA가 곧 버전**이라 새 커밋이 바로 새 버전으로 잡힌다.

타이밍을 기다리지 않고 **확정적으로** 반영하려면 다음을 쓴다:

```bash
claude plugin marketplace update claude-workflow-kit
```
