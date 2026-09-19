# statusline

statusline은 Claude Code의 `statusLine` 명령으로 실행되는 Go 바이너리 하나로 이루어진다.
바이너리는 stdin으로 세션 JSON을 받아 stdout에 한 줄을 출력하고 종료한다.

## 출력

```text
Opus 5 | 145.7k (15%) ctx | $5.18 session | 8% 5h → 15:20 | 7% 7d → 09-25 07:00 | idle 10s
```

- 세그먼트는 아래 표의 순서로 출력하고, 세그먼트 사이는 앞뒤에 공백이 한 칸씩 붙은 `|`로 구분한다.
- 입력에 값이 없는 세그먼트는 생략하고, 남은 세그먼트만 구분자로 연결한다.

| 세그먼트 | stdin 필드                                                                 | 예시                  | 값의 색              |
| -------- | -------------------------------------------------------------------------- | --------------------- | -------------------- |
| 모델     | `model.display_name`                                                       | `Opus 5`              | 초록                 |
| 컨텍스트 | `context_window.total_input_tokens`, `context_window.used_percentage`      | `145.7k (15%) ctx`    | 사용률 색(기본 자홍) |
| 비용     | `cost.total_cost_usd`                                                      | `$5.18 session`       | 기본 전경색          |
| 5h 쿼터  | `rate_limits.five_hour.used_percentage`, `rate_limits.five_hour.resets_at` | `8% 5h → 15:20`       | 사용률 색(기본 초록) |
| 7d 쿼터  | `rate_limits.seven_day.used_percentage`, `rate_limits.seven_day.resets_at` | `7% 7d → 09-25 07:00` | 사용률 색(기본 초록) |
| idle     | `transcript_path`가 가리키는 파일의 수정 시각                              | `idle 10s`            | 경과 시간 색         |

- 각 세그먼트는 사용하는 필드 중 하나라도 없으면 생략한다.
  - 예외로 쿼터 세그먼트는 `resets_at`이 없어도 표시하고, `→`와 시각만 생략한다.
- `gpt`로 시작하는 모델의 stdin에 `rate_limits`가 없으면 쿼터 세그먼트의 값을 [CLIProxyAPI 쿼터](#cliproxyapi-쿼터)에서 가져온다.

### 색

| 색          | ANSI 코드    | 쓰는 곳                                                   |
| ----------- | ------------ | --------------------------------------------------------- |
| 빨강        | `\033[0;31m` | 사용률 80 이상, idle 1시간 이상                           |
| 노랑        | `\033[0;33m` | 사용률 50 이상, idle 5분 이상 1시간 미만                  |
| 초록        | `\033[0;32m` | 모델, 쿼터 기본색, idle 5분 미만                          |
| 자홍        | `\033[0;35m` | 컨텍스트 기본색                                           |
| 기본 전경색 | `\033[0;39m` | 비용 값, 쿼터 리셋 시각                                   |
| 회색        | `\033[0;90m` | 구분자, 레이블(`ctx`, `session`, `5h`, `7d`, `idle`), `→` |

- 색을 입힌 구간은 끝마다 `\033[0m`으로 색을 초기화한다.

### 사용률

- 컨텍스트와 쿼터의 사용률은 짝수 쪽으로 반올림한 정수로 표시하고, 색도 그 정수로 정한다.
  - 80 이상이면 빨강, 50 이상이면 노랑이다.
  - 그 밖에는 세그먼트의 기본색을 쓴다.

### 컨텍스트

- 토큰 수는 크기에 따라 단위를 붙인다.
  - 1,000,000 이상: 소수 첫째 자리까지 `M` (예: `1.2M`)
  - 1,000 이상: 소수 첫째 자리까지 `k` (예: `145.7k`)
  - 그 밖: 정수

### 쿼터

- 리셋 시각은 `resets_at`(Unix epoch 초)을 로컬 타임존으로 바꿔 표시한다.
  - 5h: `HH:MM` (예: `03:30`)
  - 7d: `MM-DD HH:MM` (예: `09-25 17:00`)
  - 시는 24시간제로 표기하고, 모든 필드는 두 자리로 맞추어 빈자리를 0으로 채운다.

### idle

| 경과 시간           | 형식    | 색   |
| ------------------- | ------- | ---- |
| 1분 미만            | `10s`   | 초록 |
| 1분 이상 5분 미만   | `1m5s`  | 초록 |
| 5분 이상 1시간 미만 | `12m3s` | 노랑 |
| 1시간 이상          | `2h15m` | 빨강 |

- 경과 시간은 transcript 파일의 수정 시각부터 현재까지다.
- transcript 파일이 없으면 세그먼트를 생략한다.

## CLIProxyAPI 쿼터

- CLIProxyAPI를 거쳐 `gpt-*` 모델을 쓰는 세션에서도 ChatGPT 구독의 5h, 7d 쿼터를 Claude 세션과 같은 형식으로 표시한다.
- 이 세션의 쿼터는 stdin의 `rate_limits` 대신 CLIProxyAPI 관리 API에서 읽는다.
  - 설치 형태와 추가 설정은 [전제 문서](../development/statusline.md#cliproxyapi)에 있다.

## 입력 처리

- stdin 전체를 UTF-8 JSON으로 파싱한다.
- 문법이 틀린 JSON이나 빈 입력을 받으면 stdin을 쓰는 세그먼트를 모두 생략한다.
- 필드의 타입이 틀리면 그 필드만 없는 것으로 보고, 나머지 필드는 그대로 사용한다.
- 세그먼트 하나를 만들다 panic이 발생하면 그 세그먼트만 생략한다.
- 파일 시스템에서는 transcript 파일의 수정 시각을 읽고, CLIProxyAPI 쿼터 캐시를 읽고 쓴다.
- 네트워크에는 CLIProxyAPI 관리 API를 조회할 때만 접근한다.

## 소스

```text
statusline/
├── go.mod           # Go 1.27, 표준 라이브러리만 사용
├── main.go          # 입력 파싱, 세그먼트 조립
├── main_test.go
├── cliproxy.go      # CLIProxyAPI 관리 API 조회
├── cliproxy_test.go
├── quota_cache.go   # CLIProxyAPI 쿼터 캐시
├── quota_cache_test.go
├── codex_quota.go   # 스냅샷 선택과 5h, 7d 창 해석
├── codex_quota_test.go
├── segments.go      # 세그먼트별 렌더링
├── segments_test.go
└── testdata/        # 테스트 입력 JSON
```

- `make statusline`의 빌드 결과물은 `build/statusline`(Windows에서는 `build/statusline.exe`)이다.

## 설치

- `user-tools-bootstrap` 플러그인의 `bootstrap-statusline` 스킬은 아래 두 가지를 수행해 statusline을 설치한다.
  - `go install`로 main의 최신 커밋을 빌드해 `~/.claude/bin/statusline`(Windows에서는 `.exe`)을 만든다.
  - `~/.claude/settings.json`의 `statusLine.command`에 그 바이너리의 절대 경로를 적고, `refreshInterval`을 `1`로 설정한다.
- 이 스킬을 다시 부르면 바이너리를 main의 최신 커밋으로 새로 빌드한다.
