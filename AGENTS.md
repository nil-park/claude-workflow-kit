# AGENTS.md

## 리포지토리

- 원격 저장소: https://github.com/nil-park/claude-workflow-kit

## 필독 문서

- [README.md](README.md): 전체 개요를 담고 있으므로 가장 먼저 읽는다.
  - 주요 기능만 다루므로, 새로 추가한 기능이 목록에 없다고 해서 빠졌다고 가정하지 않는다.
- [docs/setup/claude-docs.md](docs/setup/claude-docs.md): 이 리포지토리의 기반이 되는 Claude Code 스펙 문서의 구성과, 요약본 대신 원문을 조회하는 방법을 안내한다.
  - 스킬, 매니페스트, 카탈로그, 훅을 고치기 전에 해당하는 스펙 원문을 먼저 읽는다.
  - 원문을 다시 조회해야 하는 시점은 이 문서에 적혀 있다.
- [docs/setup/agy-docs.md](docs/setup/agy-docs.md): agy 스펙 문서의 구성과 원문 조회 방법을 안내한다.
  - agy가 읽는 파일을 고치기 전에 해당하는 스펙 원문을 먼저 읽는다.
  - 원문을 다시 조회해야 하는 시점은 이 문서에 적혀 있다.

## 언어

- `AGENTS.md`, 이슈, `docs/` 아래의 모든 문서, 스킬 본문은 한국어로 쓴다.
- 커밋 메시지, PR 제목과 본문, 플러그인과 마켓플레이스 매니페스트는 영어로 쓴다.

## Git 규약

- 브랜치 이름과 PR 규약은 `git-workflow` 스킬을 따른다.

## 개발 규약

- 세션의 첫 프롬프트에서 `git ls-files --cached --others --exclude-standard`를 실행해 `.gitignore`를 반영한 최신 프로젝트 파일 목록을 확인한다.
  - `--others --exclude-standard`는 `.gitignore`에 걸리지 않는 미추적 파일도 함께 출력한다.

## 포매팅

- Markdown, JSON, YAML 파일은 Prettier로 포매팅한다.
  - 해당 파일을 고친 뒤에는 `npx prettier`를 부르는 `make format`을 실행한다.
- `main`에 커밋하기 전이나 PR의 마지막 단계에서 `make format`을 한 번 실행해 포매팅 변경만 따로 커밋하지 않도록 한다.
- `.prettierignore`는 빌드 산출물과 도구 디렉터리를 제외한다.
  - 해당 확장자를 내보내는 빌드·생성 경로가 새로 생기면 `.prettierignore`에 추가한다.
- Prettier Markdown 주의 사항: 줄바꿈한 산문 줄이 `+`, `-`, `*`로 시작하면 Prettier가 그 기호를 목록 기호로 읽는다.
  - 이어지는 줄을 이 기호로 시작하지 않는다(예: 줄 앞에 `+` 대신 `및`을 쓴다).
  - 기호로 시작하면 Prettier가 하위 불릿으로 바꿔 의미가 달라진다.
