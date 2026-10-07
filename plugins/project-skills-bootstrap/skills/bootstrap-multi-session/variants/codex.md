# codex

## peer-cli

```text
Codex CLI
```

## peer-setup

````text
## 사전 점검

1. `codex --version`으로 Codex CLI가 설치되어 있는지 확인한다.
   - 없으면 사용자에게 설치와 로그인을 요청한다.

2. `~/.codex/config.toml`의 `model`과 `model_reasoning_effort` 값을 읽어 사용자에게 보여 주고 확인을 요청한다.
   - 사용자가 두 값을 확인하기 전에는 작업을 시작하지 않는다.
   - 이 파일은 사용자의 전역 설정이므로 직접 고치지 않는다.

3. `~/.codex/config.toml`에 `approvals_reviewer = "auto_review"`가 설정되어 있는지 확인한다.
   - 이 설정이 없으면 peer가 네트워크를 쓰는 셸 명령을 실행하지 못한다.
   - 설정되어 있지 않으면 사용자에게 설정을 요청하고, 사용자가 설정하기 전에는 작업을 시작하지 않는다.

4. `~/.codex/AGENTS.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - Codex CLI는 이 파일을 모든 세션에 사용자 지침으로 불러온다.
   - 없으면 사용자에게 알리고, 사용자가 추가하기 전에는 작업을 시작하지 않는다.

작업을 시작할 때 사용자에게 peer 세션이 어디에 남는지 간단히 알린다.

- peer 세션의 기록은 `~/.codex/sessions/<YYYY>/<MM>/<DD>/rollout-*.jsonl`에 남는다.
- 세션 ID는 `codex-peers.md`에 기록해 두었으므로 정리할 때 참고하라고 안내한다.

## peer 호출

- peer는 서브 에이전트가 아니라 세션 ID로 구분하는 Codex CLI 세션이다.
  - 동시에 띄우는 peer는 8개로 한다.
- peer마다 지시를 스크래치패드에 파일로 쓰고, Bash의 `run_in_background`로 명령을 실행한다.
  - 응답은 `-o <결과 파일>`로 지정한 파일에 최종 응답만 따로 받는다.
  - 응답 완료 여부는 백그라운드 작업의 완료 알림으로 확인한다.
  - 여러 peer에게 지시를 보낼 때는 각 명령을 한 번에 병렬로 실행한다.

```bash
# 첫 지시: 세션을 만든다
codex exec --json -s workspace-write -o <결과 파일의 절대 경로> "$(cat <지시 파일의 절대 경로>)" < /dev/null

# 이후 지시: 같은 세션의 맥락을 이어 간다
codex exec -s workspace-write resume <세션 ID> -o <결과 파일의 절대 경로> "$(cat <지시 파일의 절대 경로>)" < /dev/null
```

- Claude Code는 Bash 허용 규칙을 명령의 앞부분과 대조하므로, 호출 명령은 `codex exec`로 시작한다.
- Claude Code 권한 검사가 호출을 거부하면 사용자에게 허용 규칙 추가를 요청한다.
  - 추가할 규칙은 `Bash(codex exec:*)`이고, 파일은 리포의 `.claude/settings.local.json`이다.
  - Claude Code 설정 파일은 오케스트레이터가 직접 고치지 않는다.
- 명령은 리포 루트에서 실행한다.
  - Codex는 세션을 재개할 때 명령을 실행한 디렉터리를 작업 디렉터리로 사용한다.
- 첫 호출의 `--json` 출력에서 `thread.started` 이벤트의 `thread_id` 값을 찾아 스크래치패드의 `codex-peers.md`에 기록한다.
  - 컨텍스트가 압축되어도 이 파일에서 peer를 다시 찾는다.
- 이후 호출은 `codex exec -s workspace-write resume <세션 ID>`로 같은 세션을 이어 간다.
  - `-s workspace-write`는 `resume` 앞에 매번 넘긴다.
- 호출 명령 끝에는 반드시 `< /dev/null`을 붙여 표준 입력을 닫는다.
- 완료 알림을 받으면 `-o`로 지정한 결과 파일을 `Read` 도구로 읽는다.
- 한 peer의 호출이 끝나기 전에는 같은 세션 ID로 다시 호출하지 않는다.
  - 같은 세션에 호출이 겹치면 Codex가 요청을 거부한다.
````

## peer-ping

```text
  - 실행 중인 호출은 새 메시지를 받지 못하므로, 그 peer의 백그라운드 작업 상태를 확인한다.
  - 작업이 끝났으면 그 작업의 결과 파일을 `Read`로 읽어 응답으로 쓴다.
  - 아직 실행 중이면 같은 세션 ID로 다시 호출하지 않고 기다린다.
```

## compared-model

```text
GPT Luna
```

## korean-writers

```text
GPT Luna는
```

## peer-model

```text
Codex CLI의 GPT
```

## assumption-doc-editing

```text
    - 대신 peer에게 도메인 가정 문서의 문체 교정 및 검수를 맡겨서 개정한다.
```

## peer-fluent-korean

```text
한국어 문장은 `~/.codex/AGENTS.md`의 `fluent-korean` 지침을 따른다. 이 지침은 한국어 표현 검수에도 적용한다.
```

## peer-report

```text
- **보고는 최종 응답 본문에 모두 적어라.**
  - 보고를 파일에만 남기지 마라.
  - 보고에 필요한 명령은 끝날 때까지 기다려 출력을 확인한 뒤 회신해라.
```

## peer-read-note

```text

```

## edit-tools

```text
`apply_patch` 도구
```

## web-tools

```text
`web_search`
```
