# gemini

## peer-cli

```text
Gemini CLI
```

## peer-setup

````text
## 사전 점검

1. 셸 rc 파일(`~/.zshrc` 등)에 아래 alias가 있는지 확인한다.
   - 해당 alias가 없으면 rc 파일의 `.bak` 사본을 같은 디렉터리에 만들고 나서 이 alias를 등록한다.

   ```bash
   alias gemini='npx @google/gemini-cli -y'
   ```

2. `~/.gemini/GEMINI.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - Gemini CLI는 이 파일을 모든 세션에 사용자 지침으로 불러온다.
   - 없으면 `scratch-dir`에 따라 리포지토리 안의 스크래치 디렉터리에 초안을 쓰고, 사용자에게 이 초안으로 파일을 만들라고 요청한다.
   - 초안은 `~/.claude/CLAUDE.md`를 Gemini CLI에 맞게 바꾼 내용과 `fluent-korean` 원문으로 구성한다.
     - `fluent-korean` 원문은 `~/.claude/plugins/marketplaces/fluent-korean/plugins/fluent-korean/output-styles/fluent-korean.md`에 있다.
     - Claude Code의 도구 이름은 Gemini CLI의 도구 이름으로 바꾸고, Gemini CLI에 없는 기능을 다루는 조항은 뺀다.
   - 사용자가 파일을 만들기 전에는 작업을 시작하지 않는다.

작업을 시작할 때 사용자에게 peer를 정리하는 방법을 간단히 알린다.

- `gemini --list-sessions`를 리포 루트에서 실행하면 이 리포의 세션 목록이 번호와 UUID로 나온다.
- 작업이 끝나면 `rm -rf ~/.gemini/tmp/<디렉터리 이름>`으로 이 리포의 세션을 한 번에 지운다.
  - 디렉터리 이름은 `~/.gemini/projects.json`에서 리포 루트 경로에 대응하는 값이다.
  - 이 디렉터리를 지우면 같은 리포에서 실행 중인 다른 Gemini 세션의 기록도 함께 사라진다는 점을 경고한다.

## peer 호출

- peer는 서브 에이전트가 아니라 UUID로 구분하는 Gemini CLI 세션이다.
  - 모델은 `gemini-3.8-flash`를 쓴다.
  - 동시에 띄우는 peer는 8개로 한다.
- peer마다 UUID를 만들고, UUID와 맡긴 파일을 스크래치패드의 `gemini-peers.md`에 기록한다.
  - 컨텍스트가 압축되어도 이 파일에서 peer를 다시 찾는다.
- 지시는 스크래치패드에 파일로 쓰고, Bash의 `run_in_background`로 아래 명령을 실행한다.
  - 응답은 백그라운드 작업의 완료 알림으로 돌아온다.
  - 여러 peer에게 지시를 보낼 때는 각 명령을 한 번에 병렬로 실행한다.

```bash
# 첫 지시: 세션을 만든다
gemini -m gemini-3.8-flash --session-id <UUID> -p "$(cat <지시 파일의 절대 경로>)"

# 이후 지시: 같은 세션의 맥락을 이어 간다
gemini -m gemini-3.8-flash --resume <UUID> -p "$(cat <지시 파일의 절대 경로>)"
```

- 명령은 리포 루트에서 실행한다.
  - 세션은 프로젝트 디렉터리별로 저장되므로, 다른 디렉터리에서는 `--resume`이 세션을 찾지 못한다.
- 한 peer의 호출이 끝나기 전에는 같은 UUID로 다시 호출하지 않는다.
- alias에 `-y`가 들어 있으므로 `--approval-mode`는 넘기지 않는다.
  - 두 옵션을 함께 지정하면 Gemini CLI가 오류를 내고 종료한다.
````

## compared-model

```text
Gemini Flash
```

## peer-model-note

```text

```

## peer-model

```text
Gemini 3.8 Flash
```

## assumption-doc-editing

```text
    - 도메인 가정 문서는 오케스트레이터가 고치고, peer에게 문체 검수를 맡긴다.
    - peer가 지적한 문장의 대안을 받아 오케스트레이터가 골라 반영한다.
```

## peer-fluent-korean

```text
한국어 문장은 `~/.gemini/GEMINI.md`의 `fluent-korean` 지침을 따른다. 이 지침은 한국어 표현 검수에도 적용한다.
```

## peer-report

```text
- **결과는 최종 응답에 담아라.** 최종 응답은 사용자가 아니라 내가 받는다.
```

## peer-read-note

```text
  - gitignore된 파일은 `read_file`로 읽지 못하므로 `cat`으로 읽어라.
```

## edit-tools

```text
`write_file`과 `replace` 도구
```

## web-tools

```text
`google_web_search`나 `web_fetch`
```
