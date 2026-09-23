# statusline의 전제와 빌드

이 문서는 [설계 문서](../architecture/statusline.md)가 전제로 삼는 사실과, 빌드와 테스트 방법을 다룬다.

## 전제

### 계정

- claude.ai Pro 또는 Max 계정에서 사용한다고 가정한다.
- Claude Code는 Pro와 Max 요금제의 stdin에 `rate_limits`를 넣는다(Claude Code 2.1.277 문서 기준).
  - 다른 요금제에서는 쿼터 세그먼트가 표시되지 않는다.

### `rate_limits`

- 필드의 형태와 실측값은 [statusline stdin 필드](../reference/statusline-stdin.md)에 정리했다.
- `used_percentage`를 파싱할 때는 정수와 소수를 모두 받아들인다.
  - 실측값은 정수였지만 공식 문서의 예시에는 `23.5` 같은 소수가 있다.
- 아래 세 가지는 실측하지 않고 공식 문서에서 확인했다.
  - 세션에서 첫 API 응답을 받기 전에는 `rate_limits`가 없다.
  - 두 쿼터 창(`five_hour`, `seven_day`) 중 한쪽만 들어올 수 있다.
  - Claude Code는 `resets_at`이 지난 쿼터 창을 stdin에서 제외한다.

### CLIProxyAPI

- 설치 형태
  - CLIProxyAPI `v7.3.*`이 `127.0.0.1:8317`에서 실행 중이고, `--codex-login`으로 ChatGPT 계정에 로그인되어 있다.
  - 셸 함수가 `ANTHROPIC_BASE_URL`을 이 주소로 지정하고 `gpt-*` 모델로 Claude Code를 실행한다.
- 추가 설정
  - CLIProxyAPI 설정의 `remote-management.secret-key`에 관리 키를 넣는다.
    - 이 값이 비어 있으면 관리 API가 꺼진다.
  - 같은 키를 환경 변수 `CLIPROXY_MANAGEMENT_KEY`로 export해 Claude Code에 넘긴다.
    - statusline은 Claude Code에게서 이 환경 변수를 물려받는다.
  - CLIProxyAPI의 주소가 다르면 `CLIPROXY_URL`로 지정한다.

### 실행 시점

- Claude Code는 세션을 시작할 때, 응답이 도착할 때, `/compact`가 끝날 때 등 이벤트가 생길 때마다 명령을 실행한다.
  - 이벤트가 연달아 생기면, 마지막 이벤트 뒤 300ms가 지난 다음 한 번만 실행한다.
- 세션이 유휴 상태이면 이벤트가 생기지 않는다.
  - idle 세그먼트를 초 단위로 갱신하려면 `statusLine.refreshInterval`을 `1`로 설정해야 한다.

### 출력

- Claude Code는 stdout의 ANSI 색상 코드를 해석해서 표시한다.

## 빌드와 테스트

| 명령              | Go 코드에 대해 하는 일                                      |
| ----------------- | ----------------------------------------------------------- |
| `make statusline` | `build/statusline`(Windows에서는 `.exe`)을 빌드한다         |
| `make format`     | `gofmt -w`로 포맷을 고치고 `go vet`, `go test`를 실행한다   |
| `make test`       | `gofmt -l`로 포맷을 검사하고 `go vet`, `go test`를 실행한다 |

- `make format`과 `make test`는 Go 외의 기존 검사도 그대로 실행한다.
