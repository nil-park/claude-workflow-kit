---
name: bootstrap-mem4j
description: >-
  환경에 설치된 CLI(Claude Code, Antigravity)를 감지하고 mem4j Go 바이너리를 빌드한 뒤,
  자격 증명 확인·로컬 스코프 설정(~/.config/mem4j/config.json)·훅 등록·CLAUDE.md/GEMINI.md 관리용 블록 배치를
  수행한다. 바이너리가 이미 있으면 main의 최신 커밋으로 새로 빌드한다.
disable-model-invocation: true
---

## 1. CLI 환경 및 기존 설치 확인

어느 환경에서 실행하든 `claude`와 `agy` CLI의 설치 여부와 기존 `mem4j` 구성을 확인한다.

- `claude`: `which claude`(Windows에서는 `where.exe claude`)가 성공하거나 `~/.claude/` 디렉터리가 있는지 확인한다.
- `agy`: `which agy`(Windows에서는 `where.exe agy`)가 성공하거나 `~/.gemini/antigravity-cli/` 디렉터리가 있는지 확인한다.
- 기존 `mem4j` 바이너리 위치(`which mem4j` 또는 `where.exe mem4j`), `~/.config/mem4j/config.json`, `~/.config/mem4j/.secrets/` 파일 존재 여부를 확인한다.
  - 자격 증명 파일의 내용은 읽거나 출력하지 않고 존재 여부만 확인한다.

## 2. 바이너리 빌드 디렉터리 합의 및 빌드

바이너리를 설치할 디렉터리를 사용자와 합의한다.

- 기존 `mem4j` 바이너리가 있으면 그 디렉터리를 우선 제안한다.
- 기존 설치가 없으면 `~/.local/bin`을 제안한다.
- 선택한 디렉터리(`<bin-dir>`)가 없으면 생성한 뒤 `mem4j`를 빌드한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/mem4j@latest
```

- 새로 빌드된 바이너리의 절대 경로(`<bin-path>`)에서 경로 구분자는 항상 슬래시(`/`)로 표기한다.
  - Windows 사용자 경로 등에 공백이 포함되어 있으면 훅 `command` 문자열에서 `<bin-path>`를 큰따옴표(`"..."`)로 감싼다.

## 3. 자격 증명 파일 확인

`~/.config/mem4j/.secrets/` 아래에 아래 두 파일이 있는지 확인한다.

- `~/.config/mem4j/.secrets/neo4j.txt` (Neo4j Aura 접속 정보)
- `~/.config/mem4j/.secrets/gcp.json` (Vertex AI 임베딩용 GCP 서비스 계정 JSON)

누락된 파일이 있으면 `docs/setup/neo4j-aura.md` 절차를 안내하고, 사용자가 파일을 배치한 뒤에 다음 단계로 넘어간다.

## 4. 기본 스코프 및 로컬 설정 작성

사용자와 상의해 이 PC에서 기본으로 활성화할 스코프 목록을 정한다.

- 사용자 스코프 이름(예: 사용자 실명 또는 별칭)을 확인한다.
- 현재 OS 사용자 계정과 호스트 이름으로 `계정@호스트` 형식의 PC 환경 스코프를 제안한다.
- 작업 디렉터리별로 추가할 리포지토리 스코프 매핑이 있으면 함께 확인한다.
- `mem4j scope list`로 기존 스코프를 확인하고, 합의한 사용자·PC 환경 스코프가 없으면 `mem4j scope create`로 생성한다.
- 기존 `~/.config/mem4j/config.json`이 있으면 `~/.config/mem4j/config.json.bak`으로 백업한 뒤 아래 형식으로 저장한다.

```json
{
  "scopes": ["<사용자 스코프>", "<계정@호스트>"],
  "workspaces": {
    "<작업 디렉터리 절대 경로>": ["<리포지토리 스코프>"]
  }
}
```

## 5. CLI 훅 등록

1단계에서 감지된 CLI 설정 파일에 `mem4j hook` 명령을 등록한다. 다른 훅 설정은 모두 보존한다.

### Claude Code (`claude` 확인 시)

`~/.claude/settings.json`을 `~/.claude/settings.json.bak`으로 백업한 뒤 `hooks` 항목에 아래 두 이벤트를 추가한다.

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "<bin-path> hook claude-session-start"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "<bin-path> hook claude-user-prompt"
          }
        ]
      }
    ]
  }
}
```

### Antigravity (`agy` 확인 시)

`~/.gemini/config/hooks.json`을 `~/.gemini/config/hooks.json.bak`으로 백업한 뒤 `mem4j` 항목을 등록한다.

```json
{
  "mem4j": {
    "SessionStart": [
      {
        "type": "command",
        "command": "<bin-path> hook agy-session-start"
      }
    ],
    "PreInvocation": [
      {
        "type": "command",
        "command": "<bin-path> hook agy-pre-invocation"
      }
    ]
  }
}
```

## 6. CLAUDE.md 및 GEMINI.md 관리용 블록 배치

감지된 CLI의 전역 지침 파일(`~/.claude/CLAUDE.md`, `~/.gemini/GEMINI.md`)을 각각 `.bak`으로 백업한다.
이 스킬의 베이스 디렉터리(`Base directory for this skill`) 아래에 있는 `mem4j-instructions.md`를 읽는다.
읽은 내용을 각 파일의 `<!-- mem4j:start -->`와 `<!-- mem4j:end -->` 사이에 삽입하거나 기존 마커 블록을 교체한다.

- 마커 밖의 사용자 전역 지침은 수정하지 않는다.
- 마커 블록 형식은 아래와 같다.

```markdown
<!-- mem4j:start -->

(<base-dir>/mem4j-instructions.md 내용)

<!-- mem4j:end -->
```

## 7. 스모크 테스트 및 결과 보고

- `mem4j scope list`로 AuraDB 연결이 정상인지 확인한다.
- 빌드된 바이너리 경로, 등록한 스코프 목록, 갱신한 설정 파일 목록을 보고한다.
- 훅은 등록 이후 새로 시작한 세션부터 적용된다고 안내한다.
- 이 스킬을 다시 실행하면 `main`의 최신 커밋으로 바이너리를 새로 빌드하고 관리용 블록을 갱신한다고 안내한다.
