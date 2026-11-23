1. `agy --version`으로 Antigravity CLI가 설치되어 있는지 확인한다.
   - 설치되어 있지 않다면 사용자에게 설치와 로그인을 요청한다.

2. `~/.gemini/antigravity-cli/settings.json`의 `toolPermission`이 `always-proceed`인지 확인한다.
   - `toolPermission`이 `always-proceed`가 아니면 사용자에게 설정을 요청하고, 사용자가 설정하기 전에는 작업을 시작하지 않는다.
   - 이 파일은 사용자의 전역 설정이므로 오케스트레이터가 직접 고치지 않는다.

3. `~/.gemini/GEMINI.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - Antigravity CLI는 이 파일을 모든 세션에 사용자 전역 규칙으로 불러온다.
   - 파일이 없거나 지침이 없으면 `scratch-dir`에 따라 리포지토리 안의 스크래치 디렉터리에 초안을 작성하고, 사용자에게 이 초안으로 파일을 만들라고 요청한다.
   - 초안은 `~/.claude/CLAUDE.md`를 Antigravity CLI에 맞게 바꾼 내용과 `fluent-korean` 원문으로 구성한다.
     - `fluent-korean` 원문은 `~/.claude/plugins/marketplaces/fluent-korean/plugins/fluent-korean/output-styles/fluent-korean.md`에 있다.
     - Claude Code의 도구 이름은 Antigravity CLI의 도구 이름으로 바꾸고, Antigravity CLI에 없는 기능을 다루는 조항은 뺀다.
   - 사용자가 파일을 만들기 전에는 작업을 시작하지 않는다.

작업을 시작할 때 사용자에게 peer 세션의 기록 위치를 간단히 알린다.

- peer 세션의 기록은 다음 경로에 남는다.
  - `~/.gemini/antigravity-cli/conversations/<세션 ID>.db`
  - `~/.gemini/antigravity-cli/brain/<세션 ID>/`
- 세션 ID는 `agy-peers.md`에 기록해 두었으므로 정리할 때 참고하라고 안내한다.
