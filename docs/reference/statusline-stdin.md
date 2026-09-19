# statusline stdin 필드

Claude Code가 `statusLine` 명령의 stdin으로 넘기는 JSON의 필드 목록이다.

- 관측 조건: Claude Code 2.1.277, claude.ai Pro 계정, Windows, 2026-09-19
  - 세션에서 응답을 여러 번 받은 뒤에 캡처했다.
- 설명은 [공식 statusline 문서](https://code.claude.com/docs/en/statusline)를 따랐다.
- 로컬 경로와 식별자는 값을 가렸다.

## 관측된 필드

### 세션

| 필드                | 타입   | 실측 예시                       | 설명                                            |
| ------------------- | ------ | ------------------------------- | ----------------------------------------------- |
| `session_id`        | 문자열 | (가림)                          | 세션 식별자                                     |
| `session_name`      | 문자열 | (가림)                          | `--name`, `/rename`으로 지정한 이름이나 AI 제목 |
| `prompt_id`         | 문자열 | (가림)                          | 처리 중인 사용자 프롬프트의 UUID                |
| `transcript_path`   | 문자열 | (가림)                          | 대화 기록 파일 경로                             |
| `cwd`               | 문자열 | (가림)                          | 현재 작업 디렉터리                              |
| `scratchpad_dir`    | 문자열 | (가림)                          | 세션의 스크래치패드 디렉터리(공식 문서에 없음)  |
| `version`           | 문자열 | `"2.1.277"`                     | Claude Code 버전                                |
| `output_style.name` | 문자열 | `"fluent-korean:fluent-korean"` | 현재 출력 스타일 이름                           |

### 모델과 설정

| 필드                 | 타입   | 실측 예시         | 설명                                               |
| -------------------- | ------ | ----------------- | -------------------------------------------------- |
| `model.id`           | 문자열 | `"claude-opus-5"` | 모델 식별자                                        |
| `model.display_name` | 문자열 | `"Opus 5"`        | 모델 표시 이름                                     |
| `effort.level`       | 문자열 | `"medium"`        | `low`, `medium`, `high`, `xhigh`, `max` 중 현재 값 |
| `thinking.enabled`   | 불리언 | `true`            | extended thinking 사용 여부                        |
| `fast_mode`          | 불리언 | `false`           | fast mode 사용 여부                                |

### 작업 공간

| 필드                                     | 타입        | 실측 예시                                             | 설명                                  |
| ---------------------------------------- | ----------- | ----------------------------------------------------- | ------------------------------------- |
| `workspace.current_dir`                  | 문자열      | (가림)                                                | 현재 작업 디렉터리(`cwd`와 같은 값)   |
| `workspace.project_dir`                  | 문자열      | (가림)                                                | Claude Code를 시작한 디렉터리         |
| `workspace.added_dirs`                   | 문자열 배열 | (가림)                                                | `/add-dir`, `--add-dir`로 추가한 경로 |
| `workspace.repo.host`, `.owner`, `.name` | 문자열      | `"github.com"`, `"nil-park"`, `"claude-workflow-kit"` | `origin` 원격에서 읽은 리포 정보      |

### 비용

| 필드                         | 타입 | 실측 예시   | 설명                                |
| ---------------------------- | ---- | ----------- | ----------------------------------- |
| `cost.total_cost_usd`        | 숫자 | `11.942875` | 정가 기준으로 추정한 세션 비용(USD) |
| `cost.total_duration_ms`     | 숫자 | `2072015`   | 세션 시작 뒤 경과한 시간(ms)        |
| `cost.total_api_duration_ms` | 숫자 | `839493`    | API 응답을 기다린 시간의 합(ms)     |
| `cost.total_lines_added`     | 숫자 | `1227`      | 추가한 코드 줄 수                   |
| `cost.total_lines_removed`   | 숫자 | `281`       | 삭제한 코드 줄 수                   |

### 컨텍스트

| 필드                                                       | 타입   | 실측 예시 | 설명                                              |
| ---------------------------------------------------------- | ------ | --------- | ------------------------------------------------- |
| `context_window.total_input_tokens`                        | 숫자   | `216297`  | 컨텍스트의 입력 토큰 수(캐시 읽기와 쓰기 포함)    |
| `context_window.total_output_tokens`                       | 숫자   | `661`     | 컨텍스트의 출력 토큰 수                           |
| `context_window.context_window_size`                       | 숫자   | `1000000` | 컨텍스트 창의 최대 토큰 수                        |
| `context_window.used_percentage`                           | 숫자   | `22`      | 컨텍스트 사용률                                   |
| `context_window.remaining_percentage`                      | 숫자   | `78`      | 컨텍스트 잔여율                                   |
| `context_window.current_usage.input_tokens`                | 숫자   | `2`       | 마지막 API 호출의 캐시되지 않은 입력 토큰 수      |
| `context_window.current_usage.output_tokens`               | 숫자   | `661`     | 마지막 API 호출의 출력 토큰 수                    |
| `context_window.current_usage.cache_creation_input_tokens` | 숫자   | `871`     | 마지막 API 호출에서 캐시에 쓴 토큰 수             |
| `context_window.current_usage.cache_read_input_tokens`     | 숫자   | `215424`  | 마지막 API 호출에서 캐시에서 읽은 토큰 수         |
| `exceeds_200k_tokens`                                      | 불리언 | `true`    | 마지막 응답의 전체 토큰 수가 200k를 넘었는지 여부 |

### 쿼터

| 필드                                    | 타입 | 실측 예시    | 설명                           |
| --------------------------------------- | ---- | ------------ | ------------------------------ |
| `rate_limits.five_hour.used_percentage` | 숫자 | `19`         | 5시간 한도 사용률(0~100)       |
| `rate_limits.five_hour.resets_at`       | 숫자 | `1789798800` | 5시간 창의 리셋 시각(epoch 초) |
| `rate_limits.seven_day.used_percentage` | 숫자 | `7`          | 7일 한도 사용률(0~100)         |
| `rate_limits.seven_day.resets_at`       | 숫자 | `1790287200` | 7일 창의 리셋 시각(epoch 초)   |

### 프롬프트 캐시

| 필드                                  | 타입               | 실측 예시         | 설명                                                   |
| ------------------------------------- | ------------------ | ----------------- | ------------------------------------------------------ |
| `prompt_cache.warm`                   | 불리언             | `true`            | 캐시된 접두부가 TTL 안에 있는지 여부                   |
| `prompt_cache.caching_observed`       | 불리언             | `true`            | 이번 세션에서 캐시 토큰이 보고된 적이 있는지 여부      |
| `prompt_cache.ttl`                    | 문자열             | `"1h"`            | 캐시 수명(`"5m"` 또는 `"1h"`)                          |
| `prompt_cache.expires_at`             | 숫자               | `1789788735`      | 캐시가 만료되는 시각(epoch 초)                         |
| `prompt_cache.requests`               | 숫자               | `89`              | 메인 대화의 API 요청 수                                |
| `prompt_cache.misses`                 | 숫자               | `0`               | 캐시에 있던 내용을 다시 처리한 요청 수                 |
| `prompt_cache.expected_rebuilds`      | 숫자               | `0`               | 컨텍스트 압축이나 도구 결과 정리 뒤의 캐시 재구성 횟수 |
| `prompt_cache.hit_ratio`              | 숫자               | `0.9844677481654` | 전체 입력 토큰 중 캐시에서 읽은 비율(0~1)              |
| `prompt_cache.cache_write_tokens`     | 숫자               | `211050`          | 이번 세션에서 캐시에 쓴 토큰 수                        |
| `prompt_cache.miss_recache_tokens`    | 숫자               | `0`               | 미스로 집계된 요청이 캐시에 쓴 토큰 수                 |
| `prompt_cache.last_miss_at`           | 숫자 또는 `null`   | `null`            | 마지막 미스 시각(epoch 초)                             |
| `prompt_cache.last_miss_cause`        | 문자열 또는 `null` | `null`            | 마지막 미스의 추정 원인                                |
| `prompt_cache.recache_tokens_if_cold` | 숫자               | `216297`          | 캐시가 만료된 뒤 다음 요청이 다시 캐시할 토큰 수       |

## 공식 문서에는 있지만 관측되지 않은 필드

- 아래 필드는 조건이 맞을 때만 들어오므로 이번 캡처에는 없었다.

| 필드                                                                           | 들어오는 조건                            |
| ------------------------------------------------------------------------------ | ---------------------------------------- |
| `rate_limits.spend_limit.used_percentage`, `rate_limits.spend_limit.resets_at` | spend limit을 설정한 Claude apps gateway |
| `prompt_cache.miss_causes`                                                     | 진단된 캐시 미스가 있을 때               |
| `workspace.git_worktree`                                                       | 연결된 git worktree 안에서 실행할 때     |
| `vim.mode`                                                                     | vim 모드를 켰을 때                       |
| `agent.name`                                                                   | `--agent`나 agent 설정으로 실행할 때     |
| `pr.number`, `pr.url`, `pr.review_state`, `pr.kind`                            | 현재 브랜치에 열린 PR이나 MR이 있을 때   |
| `worktree.name`, `.path`, `.branch`, `.original_cwd`, `.original_branch`       | worktree 세션일 때                       |
