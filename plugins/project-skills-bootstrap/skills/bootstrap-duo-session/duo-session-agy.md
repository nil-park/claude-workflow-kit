---
name: duo-session-agy
description: >-
  main(이 Claude Code 세션)이 지시하고 리뷰하며, 다른 터미널의 agy 세션(peer)이 리포지토리 파일을 모두 편집한다.
  두 세션은 duo-relay로 메시지를 주고받는다.
disable-model-invocation: true
---

# agy 듀오 세션 작업

- 사용자가 이 방식으로 지시한 작업에만 적용한다.
  - 지시를 받기 전에는 기다린다.
  - 질문을 지시로 받거나 의도를 추론해 작업을 시작하지 않는다.
- `main`은 이 Claude Code 세션이고, `peer`는 사용자가 `agy --agent duo-session-agy-peer`로 시작한 대화형 agy 세션이다.

## 역할

- `peer`는 한국어 문서와 코드를 가리지 않고 리포지토리의 모든 파일을 편집한다.
  - 파일을 편집한 뒤에는 포매터를 실행한다.
  - 커밋 메시지와 PR 제목·본문도 작성한다.
- `main`은 나머지 작업을 모두 맡는다.
  - 작업 범위·전제 문서와 설계 문서의 영문 초안 작성
  - 리뷰, diff 확인, 테스트
  - `git`과 `gh` 명령 실행(커밋, 브랜치, PR)
- `main`은 리포지토리 파일에 한국어를 직접 쓰지 않는다.
  - 한국어 문장이 어색하면 `main`은 지적 사항으로 적어 `peer`에 돌려보낸다.
- `main`은 `peer`에 보내는 모든 메시지를 영어로 쓴다.

## 공유 디렉터리

- 경로: `<OS-temp>/agent/<project-slug>/duo-session-agy/`
  - `project-slug`: 리포지토리 루트의 절대 경로에서 경로 구분자와 `:`를 `-`로 바꾼 값(예: `C:\Git\my-repo` → `C--Git-my-repo`)
- `main`은 듀오 세션을 시작할 때 공유 디렉터리를 비우며, 디렉터리가 없으면 새로 만든다.
- `main`은 공유 디렉터리의 `instruction-<N>.md`에 메시지 본문을 쓰고, `peer`는 `reply-<N>.md`에 회신을 쓴다.
- `peer`는 공유 디렉터리의 `commit-message.md`에 커밋 메시지를 쓰고, `pr-title.md`와 `pr-body.md`에 PR 제목과 본문을 쓴다.

## 사전 점검

`main`은 `시작` 1단계에 들어가기 전에 다음 항목을 확인한다.

- 1~5번과 7번을 통과하기 전에는 작업을 시작하지 않는다.
- 6번에서 상시 로드 규칙이 있으면 사용자가 진행하기로 결정했을 때만 시작한다.

1. `agy --version`으로 agy가 설치되어 있는지 확인한다.
   - 설치되어 있지 않으면 사용자에게 agy 설치와 로그인을 요청한다.
2. `duo-relay`가 `PATH`에 있는지 확인한다.
   - 없으면 사용자에게 `bootstrap-duo-session` 스킬 실행을 요청한다.
3. `.agents/agents/duo-session-agy-peer/agent.md`가 있는지 확인한다.
   - 없으면 사용자에게 `bootstrap-duo-session` 스킬 실행을 요청한다.
4. `.claude/settings.json`의 `permissions.allow`에 `Bash(duo-relay:*)`가 있는지 확인한다.
   - 없으면 사용자에게 추가를 요청한다.
5. `~/.gemini/antigravity-cli/settings.json`에서 `toolPermission`이 `always-proceed`이고 `allowNonWorkspaceAccess`가 `true`인지 확인한다.
   - 설정되어 있지 않으면 사용자에게 설정을 요청한다.
   - 이 파일은 사용자의 전역 설정이므로 `main`이 직접 고치지 않는다.
6. `.agents/rules/`에 상시 로드 규칙이 없는지 확인한다.
   - 규칙이 있으면 `peer`도 그 규칙을 함께 읽어 들인다고 사용자에게 알린다.
   - `work-cycle` 및 `git-workflow` 규칙이 `peer`의 역할과 겹친다는 점도 함께 알린다.
   - 계속 진행할지는 사용자가 결정한다.
7. `~/.gemini/GEMINI.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - agy는 이 파일을 모든 세션에 사용자 전역 규칙으로 불러온다.
   - 파일이 없거나 지침이 없으면 `scratch-dir`에 따라 리포지토리 안의 스크래치 디렉터리에 초안을 작성하고, 사용자에게 이 초안으로 파일을 만들라고 요청한다.
   - 초안은 `~/.claude/CLAUDE.md`를 agy에 맞게 바꾼 내용과 `fluent-korean` 원문으로 구성한다.
     - `fluent-korean` 원문은 `~/.claude/plugins/marketplaces/fluent-korean/plugins/fluent-korean/output-styles/fluent-korean.md`에 있다.
     - Claude Code의 도구 이름은 agy의 도구 이름으로 바꾸고, agy에 없는 기능을 다루는 조항은 뺀다.
     - `~/.claude/CLAUDE.md`에는 Claude Code의 `anti-claudeism` Stop 훅을 위해 `sed`나 heredoc 같은 셸 명령으로 파일을 고치지 못하게 하는 규칙이 있다.
       - `main`은 초안에서 이 규칙을 뺀다.
   - 사용자가 파일을 만들기 전에는 작업을 시작하지 않는다.

## 시작

1. `peer`가 실행 중이 아니면 다른 터미널에서 `agy --agent duo-session-agy-peer`로 `peer`를 시작해 달라고 사용자에게 요청한다.
2. 공유 디렉터리를 비운다.
3. 1라운드 지시문을 쓰고 `duo-relay push --dir <shared directory> --to peer --round 1 --file <shared directory>/instruction-1.md`를 실행한다.
   - `peer`가 아직 등록하지 않았으므로 이 명령은 파일만 쓴다.
4. `main`은 `peer`에 붙여 넣을 다음 한 줄을 사용자에게 건넨다.
   `Your shared directory is <shared directory>. Register, then follow round-001-to-peer.md in it.`
5. 백그라운드 명령으로 `duo-relay wait --dir <shared directory> --role main --round 1 --timeout 1800`을 실행한다.

## 라운드 진행

1. `instruction-<N>.md`를 쓰고 `duo-relay push ... --to peer --round <N> --file ...`를 실행한다.
2. 백그라운드 명령으로 `duo-relay wait ... --role main --round <N> --timeout 1800`을 실행한다.
3. `wait`가 끝나면 출력된 회신을 읽고 `peer`의 변경 사항을 리뷰한다.
4. 지적 사항을 다음 라운드로 보낸다.
   - 지적이 없는 라운드는 클린 패스로 센다.

- 종료 코드 2(타임아웃): 사용자에게 지연을 알리고 새 `wait`를 시작하지 않는다.
- 종료 코드 3(중복 실행): `main`의 `wait`가 이미 실행 중이므로 새 `wait`를 시작하지 않는다.
- 사용자가 `peer`를 다시 시작하면, 시작 4단계의 문구에서 파일 이름을 현재 라운드 파일로 바꿔 사용자에게 다시 건넨다.
  - `peer`는 다시 등록한다.

## 지시문 작성

- 편집할 모든 파일과 읽어야 할 모든 파일의 경로를 적는다.
- 지적 사항에 번호를 매기고, 항목마다 파일과 줄 번호를 적는다.
- 무엇이 잘못되었는지와 고친 결과가 만족해야 할 조건을 적는다.
  - 바꿀 한국어 문장을 직접 써 주지 않는다.
- `main`이 `peer`의 한국어 문장을 인용할 때는 `anti-claudeism` Stop 훅이 검사를 건너뛰도록 표기한다.
  - 인라인 인용은 `⁌`와 `⁍`로 감싼다.
  - 줄 단위 인용에는 `main`이 마크다운 인용 블록(`>`)을 쓴다.
- 같은 종류의 기준 위반이 다시 나오면, `main`은 `peer`에게 해당 기준 문서를 다시 읽으라고 지시문에 적는다.

## 리뷰

- 매번 리뷰하기 직전에 `docs-standards`, `coding-standards`, `anti-claudeism`의 원문을 다시 읽는다.
  - 이전에 읽은 기억에 의존하지 않는다.
- 변경된 파일을 모두 처음부터 끝까지 읽고 `git diff`를 확인한다.
- `.agents/skills/anti-claudeism/anti_claudeism.py -f`로 변경된 파일을 검사한다.
  - 공유 디렉터리에 `commit-message.md`, `pr-title.md`, `pr-body.md`가 작성되어 있으면 함께 검사한다.
  - 실제 결함은 `peer`에게 리뷰 지적 사항으로 보낸다.
- 작업 범위·전제 문서와 설계 문서는 `main`의 영문 초안에서 뜻이 달라진 곳이 있는지도 함께 확인한다.

## 작업 범위·전제 문서와 설계 문서

- `main`은 전제의 범위를 벗어나는 내용을 요구하지 않는다.
  - 지시문에서 방어 로직과 설명을 추가하도록 요청하지 않는다.
  - 리뷰에서 전제를 벗어난 지적 사항을 보내지 않는다.
- 작업 범위·전제 문서에는 다음 항목을 기록한다.
  - 현재 작업의 기준(코드, 설계 문서, 또는 외부 명세)
  - 규모, 실행 환경, 런타임 전제
  - 사용자의 판정 사항
  - `peer`의 반론에 따라 `main`이 철회한 지적 사항과 `peer`의 근거
    - `main`은 철회한 지적 사항을 다시 보내지 않는다.
- 작업 범위·전제 문서의 내용 가운데 추후 개발자에게 필요한 사항만 `docs-standards`에 따라 리포지토리 문서에 반영한다.

1. `main`이 하네스 스크래치패드에 영문 불릿으로 초안을 쓴다.
2. `peer`가 초안을 번역하지 않고 리포지토리의 한국어 문서로 다시 쓴다.
   - 작업 범위·전제 문서는 `git-workflow`가 사용자 확인용 문서를 두는 리포지토리의 임시 디렉터리에 쓴다.
   - 설계 문서는 리포지토리의 문서 디렉터리에 쓴다.
3. `main`이 리뷰하고 지적 사항을 보낸다.
   - 지적이 없는 라운드가 나올 때까지 2단계와 3단계를 반복한다.
4. `main`이 사용자에게 보고한다.
   - 작업 범위·전제 문서는 사용자의 확인으로 단계를 마친다.
   - 설계 문서는 사용자의 최종 검토가 끝나면 `main`이 커밋한다.

- `main`은 첫 지시문에서 `peer`에게 영문 초안의 경로를 전달하고, 작업 중에 초안이 바뀌면 이를 `peer`에게 알린다.
- 4단계 이후에는 `main`이 영문 초안만 수정한다.
  - 사용자에게 보고하기 직전에 `main`은 `peer`에게 영문 초안에 맞추어 한국어 문서를 최신화하도록 요청한다.
- 작업이 끝날 때 `main`은 사용자에게 전제 사항을 다시 확인받는다.

## 구현

- `main`이 지시하고, `peer`가 편집하고, `main`이 리뷰해 지적 사항을 보낸다.
- `main`은 테스트를 실행한다.
- 클린 패스가 3회 연속으로 나오면 구현 단계를 마치고, `main`이 사용자에게 보고한다.
- `main`의 최종 보고에는 다음 항목을 포함한다.
  - 작업 중에 넓어진 범위
  - 사용자 확인 없이 `main`이 직접 정한 전제
  - 작업 진행 중에 뒤집힌 결정 사항
  - `peer`의 반론에 따라 `main`이 철회한 지적 사항과 `peer`의 근거
  - 실행한 검증과 건너뛴 검증

## 커밋과 PR

- `main`은 커밋하기 직전에 `peer`에게 커밋 메시지 작성을 요청한다.
  - `peer`는 공유 디렉터리의 `commit-message.md`에 커밋 메시지를 쓴다.
- `main`은 PR을 열거나 수정할 때 `peer`에게 PR 제목과 본문 작성을 요청한다.
  - `peer`는 공유 디렉터리의 `pr-title.md`와 `pr-body.md`에 PR 제목과 본문을 쓴다.
- `main`은 무엇이 어떻게 바뀌었는지와 그 이유를 영어로 전달하고, `peer`는 리포지토리 규칙이 정한 언어로 본문을 작성한다.
- `main`은 이 파일들도 리포지토리 파일과 같은 방식으로 리뷰해 지적 사항을 돌려보낸다.
- 리뷰가 끝나면 `main`은 이 파일들을 넘겨 `git commit -F`, `gh pr create --body-file`, `gh pr edit --body-file` 명령을 실행한다.
