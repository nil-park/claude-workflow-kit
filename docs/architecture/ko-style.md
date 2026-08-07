# 번역투 탐지 훅

`ko-style`은 한국어 문서에 영문 직역체가 들어가면 Claude Code의 턴을 종료하지 못하게 하고,
탐지된 어휘를 알리는 Stop 훅이다.

어휘는 이 훅이 맡고, 문장이 성립하는지는 `workflow-core:docs-standards`의 리뷰 기준이 맡는다.

## 동작

```mermaid
flowchart TD
  stop["Stop 이벤트"] --> active{"stop_hook_active"}
  active -->|true| pass["턴을 끝낸다"]
  active -->|false| collect["이번 턴의 Write/Edit 경로 수집"]
  collect --> empty{"대상이 있나"}
  empty -->|없음| pass
  empty -->|있음| lines["한국어가 든 줄 추출"]
  lines --> match{"사전에 있나"}
  match -->|없음| pass
  match -->|있음| report["additionalContext로 알린다"]
  report --> again["Claude가 탐지 결과를 인지한다"]
```

- 판정은 사전의 키워드 매칭이다. 에이전트를 띄우지 않는다.
- 훅은 탐지된 어휘와 대체 표현을 알리고, 고칠지 말지는 Claude가 정한다.

## 판정 대상

- 이번 턴에 `Write`·`Edit`·`MultiEdit`·`NotebookEdit`으로 고친 파일 중 한국어가 든 줄이다.
- 사용자가 끊은 턴의 편집은 판정하지 않고 다음 턴으로 넘기지도 않는다.
- 경로는 `transcript_path`의 JSONL에서 마지막 사용자 입력 이후의 `tool_use` 블록으로 모은다.

## 사전

- 실제로 틀렸던 표현만 등재한다.
- 항목은 세 필드로 이루어지고, 스키마의 SoT는 사전 파일이다.

| 필드   | 내용                                    |
| ------ | --------------------------------------- |
| `term` | 매칭할 표현                             |
| `as`   | 무엇으로 판정했는지                     |
| `use`  | 대신 쓸 표현. `null`이면 그 항목을 끈다 |

- 훅이 내보내는 문구는 `term`이 `as`에 해당하면 `use`로 바꾸라는 형태다.
  - `as`에는 직역의 원어와 정서법의 이름이 함께 들어간다 — `consumer의 직역`, `이중 피동`.
- `term`은 정규식이다. 특수문자가 없으면 부분 문자열과 같게 동작한다.
  - `소비자`가 `소비자가`를, `되어지`가 `되어졌다`를 잡는다.
- 다른 단어에 파묻히는 표현은 경계를 건다.
  - `(?<![가-힣])축이`는 `축이라는`과 `축이었다`를 잡고 `건축이`와 `압축이`는 거른다.
- 컴파일되지 않는 `term`은 건너뛴다.

### 사전 파일

셋을 순서대로 읽어 합친다. 같은 `term`이 겹치면 뒤에 읽은 것이 이긴다.

| 순서 | 파일                                          | 범위                     |
| ---- | --------------------------------------------- | ------------------------ |
| 1    | `${CLAUDE_PLUGIN_ROOT}/hooks/dictionary.json` | 플러그인과 함께 배포된다 |
| 2    | `~/.claude/ko-style-dictionary.json`          | 이 사용자의 모든 작업    |
| 3    | `<cwd>/.claude/ko-style-dictionary.json`      | 이 프로젝트              |

- 1번 외에는 없어도 그대로 넘어간다.
- 프로젝트 경로는 Stop 입력의 `cwd`를 쓴다.
- 2번과 3번에 같은 `term`을 `use: null`로 두면 그 항목만 꺼진다.

## 신호

stdout에 `hookSpecificOutput.additionalContext`를 실어 알린다. 종료 코드는 탐지 여부와
무관하게 언제나 0이고, 2는 쓰지 않는다.

루프 보호는 입력의 `stop_hook_active`와 8회 연속 상한이다.

## 배포

`workflow-core`와 별개인 `ko-style` 플러그인으로 배포하고, `hooks/hooks.json`이 `Stop`에
스크립트를 건다. 스크립트는 Python으로 쓰고
`python3 "${CLAUDE_PLUGIN_ROOT}/hooks/ko-style.py"`로 부른다.
