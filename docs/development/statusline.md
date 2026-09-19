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

- CLIProxyAPI 7.3.7을 거쳐 `gpt-*` 모델을 쓰는 세션을 대상으로 한다.
  - 필드와 동작은 v7.3.7 소스와 테스트에서 확인했고, 실제 서버의 응답은 아직 관측하지 않았다.
- 이 세션의 stdin에는 `rate_limits`가 없다고 가정한다.
  - CLIProxyAPI는 클라이언트로 보내는 응답에 `anthropic-ratelimit-unified-*` 헤더를 쓰지 않는다.
- CLIProxyAPI는 Codex 업스트림 응답을 받을 때마다 쿼터 헤더를 스냅샷으로 저장한다.
  - 자격 증명별 스냅샷과 요청한 모델의 스냅샷을 함께 교체한다.
  - websocket 경로의 `codex.rate_limits` 이벤트도 같은 `X-Codex-*` 이름으로 바꿔 저장한다.
  - 관리 API는 스냅샷을 `quota`와 `model_quotas.<모델 이름>`으로 반환하고, `observed_at`은 RFC3339 문자열이다.
- 아래 두 가지는 소스에서 확인하지 못해 가정으로 둔다.
  - `model_quotas`의 키는 stdin의 `model.id`와 같은 문자열이다.
  - ChatGPT 구독의 5h, 7d 창은 `Window-Minutes`가 각각 `300`, `10080`으로 들어온다.
- 추가 한도(예: `GPT-5.3-Codex-Spark`)는 `X-Codex-<짧은 이름>-*`나 `X-Codex-Additional-<한도 이름>-*`로 저장된다.
- 관리 API는 관리 키가 있어야 열린다.
  - 관리 키는 CLIProxyAPI 설정의 `remote-management.secret-key`나 환경 변수 `MANAGEMENT_PASSWORD`로 설정한다.
  - 관리 키가 설정되지 않은 서버는 404를 반환한다.
  - 키를 넘기지 않거나 틀린 키를 넘기면 401을 반환한다.
  - 기본 설정(`allow-remote: false`)에서는 localhost 호출만 받는다.
- Claude Code는 statusline 명령에 자기 환경 변수를 물려준다고 가정한다.

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
