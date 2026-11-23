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
