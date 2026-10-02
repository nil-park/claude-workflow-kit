# copilot

## peer-cli

```text
GitHub Copilot CLI
```

## peer-setup

````text
## 사전 점검

1. `copilot --version`으로 GitHub Copilot CLI가 설치되어 있는지 확인한다.
   - 없으면 사용자에게 설치와 `copilot login`을 요청한다.

2. `~/.copilot/settings.json`의 `model`과 `effortLevel` 값을 읽어 사용자에게 보여 주고 확인을 요청한다.
   - 사용자가 두 값을 확인하기 전에는 작업을 시작하지 않는다.
   - 이 파일은 사용자의 전역 설정이므로 직접 고치지 않는다.

3. `~/.copilot/copilot-instructions.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - Copilot CLI는 이 파일을 모든 세션에 개인 지침으로 불러온다.
   - 없으면 사용자에게 알리고, 사용자가 추가하기 전에는 작업을 시작하지 않는다.

4. Claude Code 설정에 peer 실행과 결과 파일 읽기를 허용하는 규칙이 있는지 확인한다.
   - 아래 세 파일을 읽어 허용 규칙을 확인한다.
     - `<리포 루트>/.claude/settings.json`
     - `<리포 루트>/.claude/settings.local.json`
     - `~/.claude/settings.json`
   - 각 규칙은 위 세 파일 중 어느 한 곳에 있으면 된다.
     - `Bash(copilot -C <리포 루트>:*)`
     - `Read(//<Claude Code 임시 디렉터리>/**)`
   - 자리표시자에는 리포 루트와 임시 디렉터리의 실제 절대 경로를 쓴다.
   - 빠진 규칙이 있으면 사용자에게 리포의 `.claude/settings.local.json`에 추가해 달라고 요청한다.
   - 사용자가 규칙을 추가하기 전에는 작업을 시작하지 않는다.
   - Claude Code 설정 파일은 오케스트레이터가 직접 고치지 않는다.

작업을 시작할 때 사용자에게 peer 세션이 어디에 남는지 간단히 알린다.

- peer 세션의 기록은 `~/.copilot/session-state/<UUID>/`에 남는다.
- UUID는 `copilot-peers.md`에 기록해 두었으므로 정리할 때 참고하라고 안내한다.

## peer 호출

- peer는 서브 에이전트가 아니라 UUID로 구분하는 Copilot CLI 세션이다.
  - 동시에 띄우는 peer는 8개로 한다.
- peer마다 UUID를 만들고, UUID와 맡긴 파일을 스크래치패드의 `copilot-peers.md`에 기록한다.
  - 컨텍스트가 압축되어도 이 파일에서 peer를 다시 찾는다.
- 지시는 스크래치패드에 파일로 쓰고, Bash의 `run_in_background`로 아래 명령을 실행한다.
  - 응답은 백그라운드 작업의 완료 알림으로 돌아온다.
  - 여러 peer에게 지시를 보낼 때는 각 명령을 한 번에 병렬로 실행한다.

```bash
copilot -C <리포 루트> --allow-all-tools -s --session-id <UUID> -p "$(cat <지시 파일의 절대 경로>)"
```

- Copilot CLI는 `-p`로 실행할 때 `defaultMode`를 적용하지 않고, 프롬프트 처리를 마친 뒤 종료한다.
- 비대화형 실행에는 `defaultPermissionMode`가 적용되지 않으므로, 호출 명령에 `--allow-all-tools`를 지정한다.
- Claude Code는 Bash 허용 규칙을 명령의 앞부분과 대조하므로, 호출 명령은 `copilot -C`로 시작한다.
- 처음 호출할 때 `--session-id`에 UUID를 넘기면 Copilot CLI가 그 UUID로 세션을 만든다.
  - 이후 같은 UUID로 호출하면 Copilot CLI가 그 세션의 맥락을 이어 간다.
- `-s`를 넘겨 통계 없이 peer의 응답만 받는다.
- 완료 알림에 나온 경로를 사용해 `Read` 도구로 peer 결과 파일을 읽는다.
- 한 peer의 호출이 끝나기 전에는 같은 UUID로 다시 호출하지 않는다.
- `--fleet`을 넘기면 peer가 서브에이전트를 띄우므로 넘기지 않는다.
````

## peer-ping

```text
  - 실행 중인 호출은 새 메시지를 받지 못하므로, 그 peer의 백그라운드 작업 상태를 확인한다.
  - 작업이 끝났으면 그 작업의 출력 파일을 `Read`로 읽어 응답으로 쓴다.
  - 아직 실행 중이면 같은 UUID로 다시 호출하지 않고 기다린다.
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
Copilot CLI의 GPT
```

## assumption-doc-editing

```text
    - 도메인 가정 문서는 오케스트레이터가 고치고, peer에게 문체 검수를 맡긴다.
    - peer가 지적한 문장의 대안을 받아 오케스트레이터가 골라 반영한다.
```

## peer-fluent-korean

```text
한국어 문장은 `~/.copilot/copilot-instructions.md`의 `fluent-korean` 지침을 따른다. 이 지침은 한국어 표현 검수에도 적용한다.
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
`web_search`나 `web_fetch`
```
