`main`은 `시작` 1단계에 들어가기 전에 다음 항목을 확인한다.

- 1~4번과 6번을 통과하기 전에는 작업을 시작하지 않는다.
- 5번에서 상시 로드 규칙이 있으면 사용자가 진행하기로 결정했을 때만 시작한다.

1. `agy --version`으로 agy가 설치되어 있는지 확인한다.
   - 설치되어 있지 않으면 사용자에게 agy 설치와 로그인을 요청한다.
2. `claude-queue`가 `PATH`에 있는지 확인한다.
   - 없으면 사용자에게 `bootstrap-duo-session` 스킬 실행을 요청한다.
3. `.agents/agents/duo-session-peer/agent.md`가 있는지 확인한다.
   - 없으면 사용자에게 `bootstrap-duo-session` 스킬 실행을 요청한다.
4. `~/.gemini/antigravity-cli/settings.json`에서 `toolPermission`이 `always-proceed`이고 `allowNonWorkspaceAccess`가 `true`인지 확인한다.
   - 설정되어 있지 않으면 사용자에게 설정을 요청한다.
   - 이 파일은 사용자의 전역 설정이므로 `main`이 직접 고치지 않는다.
5. `.agents/rules/`에 상시 로드 규칙이 없는지 확인한다.
   - 규칙이 있으면 `peer`도 그 규칙을 함께 읽어 들인다고 사용자에게 알린다.
   - `work-cycle` 및 `git-workflow` 규칙이 `peer`의 역할과 겹친다는 점도 함께 알린다.
   - 계속 진행할지는 사용자가 결정한다.
6. `~/.gemini/GEMINI.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - agy는 이 파일을 모든 세션에 사용자 전역 규칙으로 불러온다.
   - 파일이 없거나 지침이 없으면 `scratch-dir`에 따라 리포지토리 안의 스크래치 디렉터리에 초안을 작성하고, 사용자에게 이 초안으로 파일을 만들라고 요청한다.
   - 초안은 `~/.claude/CLAUDE.md`를 agy에 맞게 바꾼 내용과 `fluent-korean` 원문으로 구성한다.
     - `fluent-korean` 원문은 `~/.claude/plugins/marketplaces/fluent-korean/plugins/fluent-korean/output-styles/fluent-korean.md`에 있다.
     - Claude Code의 도구 이름은 agy의 도구 이름으로 바꾸고, agy에 없는 기능을 다루는 조항은 뺀다.
     - `~/.claude/CLAUDE.md`에는 Claude Code의 `anti-claudeism` Stop 훅을 위해 `sed`나 heredoc 같은 셸 명령으로 파일을 고치지 못하게 하는 규칙이 있다.
       - `main`은 초안에서 이 규칙을 뺀다.
   - 사용자가 파일을 만들기 전에는 작업을 시작하지 않는다.
