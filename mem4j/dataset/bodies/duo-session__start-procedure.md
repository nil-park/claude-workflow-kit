1. `claude-queue list`를 실행해 `CWD`가 리포지토리 루트인 세션을 찾아 `main` 자신의 이름을 확인한다.
   - 해당 세션이 없거나 둘 이상이면 사용자에게 묻는다.
2. 공유 디렉터리에 `instruction-1.md`를 쓴다.
3. `peer`가 실행 중이 아니면 다른 터미널에서 실행할 다음 명령을 사용자에게 건네고, 이미 실행 중이면 큰따옴표 안의 영문 한 줄을 `peer`에 붙여 넣도록 사용자에게 건넨다.
   `agy --agent duo-session-peer -i "Your shared directory is <shared directory>. Follow <shared directory>/instruction-1.md and reply to <main name>."`
4. `peer`의 첫 회신에는 핸드셰이크가 함께 온다.
   - 핸드셰이크에 담긴 세 가지 `ANTIGRAVITY_*` 값을 하네스 스크래치패드에 기록한다.
