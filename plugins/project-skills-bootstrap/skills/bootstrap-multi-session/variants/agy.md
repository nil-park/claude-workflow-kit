# agy

## peer-cli

```text
Antigravity CLI
```

## peer-setup

````text
## 사전 점검

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

## peer 호출

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
````

## peer-ping

```text
  - 실행 중인 호출은 새 메시지를 받지 못하므로, 그 peer의 백그라운드 작업 상태를 확인한다.
  - 작업이 끝났으면 그 작업의 출력 파일을 `Read`로 읽어 응답으로 쓴다.
  - 아직 실행 중이면 같은 세션 ID로 다시 호출하지 않고 기다린다.
```

## compared-model

```text
Gemini Flash
```

## korean-writers

```text
Gemini Flash는
```

## peer-model

```text
Gemini Flash
```

## assumption-doc-editing

```text
    - 대신 peer에게 도메인 가정 문서의 문체 교정 및 검수를 맡겨서 개정한다.
```

## peer-fluent-korean

```text
한국어 문장은 `~/.gemini/GEMINI.md`의 `fluent-korean` 지침을 따른다. 이 지침은 한국어 표현 검수에도 적용한다.
```

## peer-report

```text
- **보고는 최종 응답 본문에 모두 적어라.**
  - 보고를 파일에만 남기지 마라.
  - 보고에 필요한 명령은 끝날 때까지 기다려 출력을 확인한 뒤 회신해라.
```

## peer-read-note

```text

```

## edit-tools

```text
`write_to_file`, `replace_file_content`, `multi_replace_file_content` 도구
```

## web-tools

```text
`search_web`이나 `read_url_content`
```
