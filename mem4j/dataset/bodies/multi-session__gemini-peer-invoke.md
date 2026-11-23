- peer는 서브 에이전트가 아니라 UUID로 구분하는 Gemini CLI 세션이다.
  - 모델은 `gemini-3.8-flash`를 쓴다.
  - 동시에 띄우는 peer는 8개로 한다.
- peer마다 UUID를 만들고, UUID와 맡긴 파일을 스크래치패드의 `gemini-peers.md`에 기록한다.
  - 컨텍스트가 압축되어도 이 파일에서 peer를 다시 찾는다.
- 지시는 스크래치패드에 파일로 쓰고, Bash의 `run_in_background`로 아래 명령을 실행한다.
  - 응답은 백그라운드 작업의 완료 알림으로 돌아온다.
  - 여러 peer에게 지시를 보낼 때는 각 명령을 한 번에 병렬로 실행한다.

```bash
# 첫 지시: 세션을 만든다
gemini -m gemini-3.8-flash --session-id <UUID> -p "$(cat <지시 파일의 절대 경로>)"

# 이후 지시: 같은 세션의 맥락을 이어 간다
gemini -m gemini-3.8-flash --resume <UUID> -p "$(cat <지시 파일의 절대 경로>)"
```

- 명령은 리포 루트에서 실행한다.
  - 세션은 프로젝트 디렉터리별로 저장되므로, 다른 디렉터리에서는 `--resume`이 세션을 찾지 못한다.
- 한 peer의 호출이 끝나기 전에는 같은 UUID로 다시 호출하지 않는다.
- alias에 `-y`가 들어 있으므로 `--approval-mode`는 넘기지 않는다.
  - 두 옵션을 함께 지정하면 Gemini CLI가 오류를 내고 종료한다.
