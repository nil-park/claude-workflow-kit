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
