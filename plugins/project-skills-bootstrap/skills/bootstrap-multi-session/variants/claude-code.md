# claude-code

## peer-cli

```text
Claude
```

## peer-setup

````text
## 사전 점검

1. 리포지토리 루트의 활성 peer 수량을 확인한다.
   - 수량이 충분하지 않다면 아래 커맨드를 실행해 peer 세션을 만들어 달라고 사용자에게 요청한다.

   ```bash
   for i in $(seq 1 8); do
     claude --bg --model 'claude-opus-4-6[1m]' --name "$(basename "$PWD")-peer-$i"
   done
   ```

작업을 시작할 때 사용자에게 peer를 정리하는 방법을 간단히 알린다.

- `claude agents`를 실행하면 TUI(Agent View라고 부른다)에서 각 세션의 상태를 확인할 수 있다.
  - 화살표 키로 에이전트를 선택해 엔터를 입력하면 세션 대화창으로 진입한다. 대화창에서 왼쪽 화살표를 누르거나 `/exit`를 입력하면 에이전트 뷰로 돌아온다.
  - 에이전트 뷰 화면에서 에이전트를 선택해 `Ctrl+X`를 두 번 연속 누르면 그 세션이 종료된다.
  - 상세 사용법은 [Agent View](https://code.claude.com/docs/en/agent-view)를 확인하라고 가이드한다.

## peer 호출

- peer는 서브 에이전트가 아니라 `ListAgents`로 확인되는 터미널에 떠 있는 다른 세션이다.
- 너는 위 커맨드로 만든 세션(peer)에 사용자가 지시하는 내용의 조사와 리뷰, 문서 작성을 맡긴다.
- peer에게 일을 맡길 때마다 **보고는 사용자가 아니라 나에게 보내라**는 문구를 빠뜨리지 않는다.
````

## compared-model

```text
Gemini Flash
```

## peer-model-note

```text
- Sonnet 4.6과 Opus 4.6은 한국어 문장을 잘 쓰지만 맥락 파악과 코딩 능력은 Opus 5에 미치지 못한다.
```

## peer-model

```text
한국어 문장력이 좋은 모델
```

## assumption-doc-editing

```text
    - 대신 peer에게 도메인 가정 문서의 문체 교정 및 검수를 맡겨서 개정한다.
```

## peer-fluent-korean

```text
출력 스타일은 `fluent-korean`의 한국어 문장 출력 지침을 따른다. 출력 지침은 한국어 표현 검수에도 적용한다.
```

## peer-report

```text
- **보고는 사용자가 아니라 나에게 보내라.**
```

## peer-read-note

```text

```

## edit-tools

```text
`Write`와 `Edit`
```

## web-tools

```text
`Web Search`나 `Web Fetch`
```
