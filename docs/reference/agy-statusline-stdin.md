# agy statusline stdin 필드

Antigravity(agy)가 `statusline` 명령의 stdin으로 넘기는 JSON의 필드 목록이다.

- 관측 조건: Antigravity 1.3.5, Google AI Pro(개인) 및 GCP 엔터프라이즈 요금제, Windows 및 Linux, 2026-10
- 로컬 경로와 식별자는 값을 가렸다.

## 관측된 필드

### 세션

| 필드                 | 타입   | 실측 예시        | 설명                |
| -------------------- | ------ | ---------------- | ------------------- |
| `session_id`         | 문자열 | (가림)           | 세션 식별자(UUID)   |
| `conversation_id`    | 문자열 | (가림)           | 대화 식별자(UUID)   |
| `conversation_title` | 문자열 | `"Network Test"` | 대화 제목           |
| `transcript_path`    | 문자열 | (가림)           | 대화 기록 파일 경로 |
| `cwd`                | 문자열 | (가림)           | 현재 작업 디렉터리  |
| `version`            | 문자열 | `"1.3.5"`        | agy CLI 버전        |
| `product`            | 문자열 | `"antigravity"`  | 제품 식별자         |
| `email`              | 문자열 | (가림)           | 사용자 계정 이메일  |
| `terminal_width`     | 숫자   | `80`             | 터미널 가로 폭      |

### 모델과 실행 상태

| 필드                 | 타입   | 실측 예시                     | 설명                                     |
| -------------------- | ------ | ----------------------------- | ---------------------------------------- |
| `model.id`           | 문자열 | `"Gemini 3.8 Flash (Medium)"` | 모델 식별자                              |
| `model.display_name` | 문자열 | `"Gemini 3.8 Flash (Medium)"` | 모델 표시 이름                           |
| `model.effort`       | 문자열 | `"medium"`                    | 추론 강도(`low`, `medium`, `high` 등)    |
| `agent_state`        | 문자열 | `"working"`                   | 에이전트 실행 상태(`working`, `idle` 등) |
| `cycle_mode`         | 문자열 | `"accept-edits"`              | 사이클 모드                              |
| `sandbox.enabled`    | 불리언 | `false`                       | 샌드박스 활성화 여부                     |
| `vcs.type`           | 문자열 | `"git"`                       | 버전 관리 시스템 종류                    |
| `plan_tier`          | 문자열 | `"Google AI Pro"`             | 요금제 식별자 또는 이름                  |

### 작업 공간

| 필드                    | 타입   | 실측 예시 | 설명                                |
| ----------------------- | ------ | --------- | ----------------------------------- |
| `workspace.current_dir` | 문자열 | (가림)    | 현재 작업 디렉터리(`cwd`와 같은 값) |
| `workspace.project_dir` | 문자열 | (가림)    | 프로젝트 루트 디렉터리              |

### 컨텍스트

| 필드                                                       | 타입   | 실측 예시 | 설명                                              |
| ---------------------------------------------------------- | ------ | --------- | ------------------------------------------------- |
| `context_window.total_input_tokens`                        | 숫자   | `19473`   | 컨텍스트의 전체 입력 토큰 수                      |
| `context_window.total_output_tokens`                       | 숫자   | `327`     | 컨텍스트의 전체 출력 토큰 수                      |
| `context_window.context_window_size`                       | 숫자   | `1048576` | 컨텍스트 창의 최대 토큰 수                        |
| `context_window.used_percentage`                           | 숫자   | `1.85709` | 컨텍스트 사용률(0~100)                            |
| `context_window.remaining_percentage`                      | 숫자   | `98.1429` | 컨텍스트 잔여율(0~100)                            |
| `context_window.current_usage.input_tokens`                | 숫자   | `13595`   | 마지막 API 호출의 캐시되지 않은 입력 토큰 수      |
| `context_window.current_usage.output_tokens`               | 숫자   | `327`     | 마지막 API 호출의 출력 토큰 수                    |
| `context_window.current_usage.cache_creation_input_tokens` | 숫자   | `0`       | 마지막 API 호출에서 캐시에 쓴 토큰 수             |
| `context_window.current_usage.cache_read_input_tokens`     | 숫자   | `0`       | 마지막 API 호출에서 캐시에서 읽은 토큰 수         |
| `exceeds_200k_tokens`                                      | 불리언 | `false`   | 마지막 응답의 전체 토큰 수가 200k를 넘었는지 여부 |

### 쿼터

| 필드                             | 타입   | 실측 예시                | 설명                      |
| -------------------------------- | ------ | ------------------------ | ------------------------- |
| `quota.<key>.remaining_fraction` | 숫자   | `0.7978984`              | 잔여 쿼터 비율(0~1)       |
| `quota.<key>.reset_time`         | 문자열 | `"2026-10-14T17:41:24Z"` | 초기화 시각(ISO 8601 UTC) |
| `quota.<key>.reset_in_seconds`   | 숫자   | `16390`                  | 초기화까지 남은 초        |

#### 요금제별 관측 키

- Google AI Pro(개인): `gemini-5h`, `gemini-weekly`, `3p-5h`, `3p-weekly`
- GCP 엔터프라이즈: 모델별 14개 키(`gemini-3.1-pro-high`, `gemini-4-argon-medium` 등)
