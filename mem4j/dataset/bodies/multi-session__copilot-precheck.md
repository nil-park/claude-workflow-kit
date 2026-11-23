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
