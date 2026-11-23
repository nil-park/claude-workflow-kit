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
