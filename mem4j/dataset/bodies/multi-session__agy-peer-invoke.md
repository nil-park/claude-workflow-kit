- peer는 서브 에이전트가 아니라 세션 ID로 구분하는 Antigravity CLI 세션이다.
  - 모델은 `gemini-3.8-flash-medium`을 쓴다.
  - 동시에 띄우는 peer는 8개로 한다.
- peer마다 식별용 이름과 맡긴 파일, 발급받은 세션 ID를 스크래치패드의 `agy-peers.md`에 기록한다.
  - 컨텍스트가 압축되어도 이 파일에서 peer를 다시 찾는다.
- 지시는 스크래치패드에 파일로 쓰고, Bash의 `run_in_background`로 아래 명령을 실행한다.
  - 응답은 백그라운드 작업의 완료 알림으로 돌아온다.
  - 여러 peer에게 지시를 보낼 때는 각 명령을 한 번에 병렬로 실행한다.

```bash
# 첫 지시: 세션을 만든다
agy -p "$(cat <지시 파일의 절대 경로>)" --model gemini-3.8-flash-medium --sandbox --output-format json

# 이후 지시: 같은 세션의 맥락을 이어 간다
agy -p "$(cat <지시 파일의 절대 경로>)" --conversation <세션 ID> --model gemini-3.8-flash-medium --sandbox --output-format json
```

- 셸 명령을 샌드박스 안에서 실행하도록 `--sandbox`를 넘긴다.
- 명령은 리포 루트에서 실행한다.
  - Antigravity CLI는 대화 기록을 작업 디렉터리별로 관리한다.
- Claude Code는 Bash 허용 규칙을 명령의 앞부분과 대조하므로, 호출 명령은 `agy -p`로 시작한다.
- Claude Code 권한 검사가 호출을 거부하면 사용자에게 허용 규칙 추가를 요청한다.
  - 추가할 규칙은 `Bash(agy -p:*)`이고, 파일은 리포의 `.claude/settings.local.json`이다.
  - Claude Code 설정 파일은 오케스트레이터가 직접 고치지 않는다.
- 첫 호출의 출력 JSON에서 `conversation_id` 값을 읽어 `agy-peers.md`에 세션 ID로 기록한다.
- 응답 본문은 출력 JSON의 `response` 값이다.
- 완료 알림에 나온 경로를 사용해 `Read` 도구로 peer 결과 파일을 읽고 JSON을 파싱해 응답을 추출한다.
- 권한은 사용자의 Antigravity 설정을 따르므로, 호출 명령에 `--dangerously-skip-permissions`를 넣지 않는다.
- 추가 디렉터리를 지정하면 작업 디렉터리가 바뀌므로 `--add-dir`는 넘기지 않는다.
- 한 peer의 호출이 끝나기 전에는 같은 세션 ID로 다시 호출하지 않는다.
