1. 리포지토리 루트의 활성 peer 수량을 확인한다.
   - 수량이 충분하지 않다면 아래 커맨드를 실행해 peer 세션을 만들어 달라고 사용자에게 요청한다.

   ```bash
   for i in $(seq 1 8); do
     claude --bg --model 'claude-opus-4-6[1m]' --name "$(basename "$PWD")-peer-$i"
   done
   ```

작업을 시작할 때 사용자에게 peer를 정리하는 방법을 간단히 알린다.

- `claude agents`를 실행하면 TUI(Agent View라고 부른다)에서 각 세션의 상태를 확인할 수 있다.
  - 화살표 키로 에이전트를 선택해 엔터를 입력하면 세션 대화창으로 진입한다. 대화창에서 왼쪽 화살표를 누르거나 `/exit`를 입력하면 에이전트 뷰로 돌아온다.
  - 에이전트 뷰 화면에서 에이전트를 선택해 `Ctrl+X`를 두 번 연속 누르면 그 세션이 종료된다.
  - 상세 사용법은 [Agent View](https://code.claude.com/docs/en/agent-view)를 확인하라고 가이드한다.
