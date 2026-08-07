# 필수 도구

이 저장소에서 작업할 때 필요한 CLI 도구 목록.

| 도구      | 용도                                                                                       | 버전 확인           |
| --------- | ------------------------------------------------------------------------------------------ | ------------------- |
| `make`    | Makefile 타깃(`make format`, `make test`) 실행                                             | `make --version`    |
| `npx`     | `make format`이 호출하는 Prettier 실행 (Node.js에 번들)                                    | `npx --version`     |
| `claude`  | Claude Code CLI — `make test`의 `claude plugin validate`, 플러그인 설치·업데이트           | `claude --version`  |
| `python3` | 훅 스크립트 실행. 3.11 이상. 플러그인을 설치해 쓰는 쪽에도 같은 이름으로 필요하다          | `python3 --version` |
| `uv`      | 훅 스크립트의 린트·타입 검사·테스트(`ruff`, `pyright`, `pytest`)를 돌린다                  | `uv --version`      |
| `eza`     | 세션 시작 시 프로젝트 구조 트리 확인 (`eza --tree --git-ignore -a --ignore-glob='.git' .`) | `eza --version`     |
