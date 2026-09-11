# anti-claudeism Hook의 전제와 검증

이 문서는 [설계 문서](../architecture/anti-claudeism.md)가 전제로 삼는 사실과, 각 사실을 검증하는 방법을 다룬다.

## Stop Hook의 입력

Stop Hook이 표준 입력으로 받는 JSON에는 아래 필드만 있다.

| 필드                     | 내용                                                                 |
| ------------------------ | -------------------------------------------------------------------- |
| `session_id`             | 세션 식별자                                                          |
| `prompt_id`              | 처리 중인 사용자 프롬프트의 UUID(v2.1.196 이상, 첫 입력 전에는 없음) |
| `transcript_path`        | 대화 기록(transcript) JSONL의 경로                                   |
| `cwd`                    | 세션을 시작한 디렉터리                                               |
| `permission_mode`        | 권한 모드                                                            |
| `effort`                 | 이번 턴의 effort 레벨                                                |
| `hook_event_name`        | `"Stop"`                                                             |
| `stop_hook_active`       | Hook 때문에 이어지는 턴이면 `true`                                   |
| `last_assistant_message` | 마지막 응답 텍스트                                                   |
| `background_tasks`       | 진행 중인 백그라운드 작업                                            |
| `session_crons`          | 세션에 등록된 예약 실행                                              |

- **입력 JSON에는 이번 턴에 고친 파일의 목록이 없다.**
- Claude Code는 transcript JSONL을 비동기로 기록한다.
  - Hook이 읽는 시점에 이번 턴의 마지막 메시지가 아직 기록되지 않았을 수 있다.
- `cwd`가 프로젝트 루트라는 보장은 없다.
- Hook은 환경변수 `CLAUDE_PROJECT_DIR`로 프로젝트 루트를 받는다.
  - `CLAUDE_PROJECT_DIR`과 `cwd`는 경로 구분자가 다를 수 있다.
    - Windows에서 `CLAUDE_PROJECT_DIR`은 `C:/Git/...`, `cwd`는 `C:\Git\...` 형식이었다.
  - 두 경로를 비교하거나 상대경로를 만들려면 먼저 정규화해야 한다.
- Hook은 Skill과 함께 설치된 사전(설치본 사전)을 Hook 스크립트와 같은 디렉터리에서 찾는다.

## 실행 환경

- Hook은 `python3`라는 이름으로 실행된다.
  - `settings.json`의 Hook 항목에는 명령을 한 줄만 적을 수 있어서, OS마다 실행 파일 이름을 다르게 지정할 수 없다.
  - python.org 인스톨러로 설치한 Windows에는 `python3.exe`가 없다.
  - pyenv-win과 Microsoft Store 배포판에는 `python3.exe`가 있다.
- `python3`는 3.11 이상이어야 한다.
- Hook은 uv나 가상환경을 거치지 않고 시스템의 `python3`로 직접 실행된다.
  - Hook은 이 리포지토리의 개발 도구와 의존성을 쓸 수 없다.
  - Hook 스크립트는 Python 표준 라이브러리만 사용한다.
- 프로젝트의 `.python-version`은 pyenv와 uv가 사용할 Python 버전을 정한다.
  - `.python-version`에 적힌 버전이 설치되어 있지 않으면, pyenv의 shim이 처리하는 `python3`로는 Hook이 실행되지 않는다.
    - 이 실패는 따로 표시되지 않는다.
- Windows에서 만든 한국어 문서는 CP949로 저장되어 있을 수 있다.
- Windows에서 Python이 텍스트 stdout에 쓰는 기본 인코딩은 로케일 코드페이지다.
  - 한국어 Windows에서는 CP949다.

## Claude Code가 정한 동작

- API 오류로 끝난 턴에서는 Stop 대신 `StopFailure` 이벤트가 발생한다.
  - `StopFailure` Hook의 출력과 종료 코드는 무시된다.
- Claude Code는 Hook이 stdout으로 내보낸 바이트를 UTF-8로 해석한다.
  - Claude Code의 Hook 문서에는 인코딩 규정이 없다.
  - 이 동작은 실측으로 확인했다.
  - UTF-8로 해석하지 못한 바이트는 U+FFFD로 대체된다.
    - 이때 JSON 파싱은 실패하지 않는다.
    - 따라서 Hook이 CP949로 출력하면, 한글이 모두 U+FFFD로 표시된 탐지 결과가 오류 없이 Claude에 전달된다.
- Hook 출력은 `additionalContext`를 포함해 10,000자가 상한이다.
  - 10,000자를 넘는 출력은 파일로 저장되고, 미리보기와 파일 경로로 대체된다.
- Claude Code는 Hook이 8회 연속으로 턴 종료를 차단하면, Hook을 무시하고 턴을 끝낸다.
  - anti-claudeism Hook은 두 번째 Stop을 통과시키므로 이 상한에 이르지 않는다.
- Hook은 탐지 결과가 Claude에 한 번 전달되는 데까지만 보장한다.
  - Claude는 지적을 읽고도 아무것도 고치지 않은 채 턴을 끝낼 수 있다.
- Hook은 Bash로 고친 파일을 검사하지 않는다.
  - `sed -i`로 고친 파일, `cat > f <<EOF`로 쓴 파일, 스크립트가 내부에서 쓰는 파일이 모두 여기에 해당한다.
  - Bash의 `tool_use` 입력에는 파일 경로가 없고 명령 문자열만 있다.
  - 이런 파일도 나중에 `Write`나 `Edit`으로 고치면 그 턴에 검사된다.
  - Hook이 검사를 건너뛰었다는 신호는 없다.
    - 오류도 경고도 나지 않으므로, 탐지 결과가 없는 턴과 구별되지 않는다.
- auto mode에서는 시스템 프롬프트에 아래 지시가 들어간다.
  - "make file changes with sed, heredocs, or short scripts, rather than using the dedicated Read, Edit, or Write tools"
  - 이 모드에서는 Claude가 파일을 Bash로 고치므로, Hook이 검사할 파일이 사실상 없다.
  - `cat`, `grep`, `find`로 파일을 읽거나 검색하는 것은 Hook의 검사 대상과 관련이 없다.

## 탐지의 한계

- 사람이 읽는 문서는 1MB를 넘을 만큼 커지지 않는다고 전제한다.
- `anti-claudeism` Skill 본문에는 틀린 표현이 예시로 그대로 적혀 있다.
  - Hook이 이 예시를 탐지하지 않도록, 예시는 `>`로 시작하는 인용 줄로 적는다.
  - 판정 근거와 주의 사항에서 사전에 등재된 낱말을 가리킬 때는 `⁌⁍`로 감싼다.
- Hook이 인용 줄을 판정하는 규칙은 CommonMark의 인용 문법과 다르다.
  - CommonMark에서는 `>` 없이 이어지는 줄도 인용에 속하지만, Hook은 이 줄을 검사한다.
  - Hook은 코드 블록 안에서 `>`로 시작하는 줄을 검사하지 않는다.
- 사전 매칭은 앞뒤 문맥을 고려하지 않는다.
  - 예를 들어 ⁌소비자⁍가 consumer의 직역으로 등재되어 있으면, 미국 경제를 다루는 문서에 쓴 ⁌소비자⁍도 탐지된다.
- 사전 항목이 많을수록 오탐도 많아진다.
  - Claude는 오탐을 포함한 모든 탐지 결과를 확인해, 직역으로 쓰였는지 판단해야 한다.
  - 따라서 사전의 크기는 매 턴의 검토 비용으로 이어진다.
- Hook은 사전에 없는 표현을 탐지하지 못한다.
- 설치본 사전은 갱신할 때 전체가 교체된다.
  - 설치본 사전에 직접 추가한 항목은 다음 갱신에서 사라진다.
- 설치된 `anti-claudeism` Skill 본문은 갱신할 때 교체되지 않는다.
  - `bootstrap-anti-claudeism`으로 갱신하면, Claude가 템플릿과 설치본의 차이를 사용자에게 보고하고 반영할 변경을 함께 정한다.
- 모든 프로젝트에 적용할 사전 항목과 Skill 본문의 예시는 이 리포지토리에서 추가한다.
- `term`이 짧을수록 다른 낱말의 일부와 일치하기 쉽다.
  - 경계 조건을 지정하지 않으면 `축`은 `축소`·`건축`·`압축`에 모두 일치한다.
- 용언은 어간과 어미가 한 음절로 합쳐지기도 하므로, `term`의 끝에 경계 조건을 지정할 수 없다.
  - ⁌되어지다⁍를 활용하면 ⁌되어진다⁍, ⁌되어졌다⁍처럼 어간의 마지막 음절이 바뀐다.
  - 여러 활용형에 공통으로 들어 있는 문자열은 ⁌되어⁍뿐이다.
    - ⁌되어⁍를 등재하면 ⁌되어서⁍와 ⁌되어야⁍까지 탐지된다.
- 이중 피동이 모두 틀린 표현은 아니다.
  - `알려지다`, `밝혀지다`, `이루어지다`는 표준어다.
  - `-어지다`로 끝나는 말을 한 항목으로 등재하면 이 표준어까지 탐지되므로, 틀린 표현만 하나씩 사전에 넣는다.

## 전제 검증 방법

- Stop Hook의 입력 필드 표는 아래 방법으로 얻는다.
  - stdin을 그대로 파일에 쓰는 Hook을 임시로 등록하고, 한 턴을 실행한다.
  - 프로젝트의 `.claude/settings.local.json`에 Hook을 등록하면 해당 리포지토리에서만 실행된다.
- Hook을 등록한 뒤에는 `/hooks`로 등록 상태를 먼저 확인한다.
  - `0 hooks configured`가 나오면 현재 세션은 Hook을 인식하지 못한 상태다.
    - 이때는 `claude --resume`으로 세션을 다시 연다.
    - 세션 도중에 `.claude/settings.local.json`을 새로 만들면 이 상태가 된다.
  - Claude Code의 Hook 문서에 따르면, 설정 파일의 Hook을 고치면 파일 감시자가 변경을 감지한다.
    - 실측에서는 없던 설정 파일을 새로 만들었을 때 파일 감시자가 변경을 감지하지 못했다.
  - Hook이 실행되지 않았는지는 겉으로 알 수 없다.
    - 따라서 등록 상태를 확인하기 전에 얻은 측정 결과는 신뢰하지 않는다.
- transcript 기록 지연의 측정 결과는 [0001](../decisions/anti-claudeism/0001-transcript-lag.md)에 있다.
  - 다시 측정할 때는 기록해 둔 `transcript_path`의 파일을 1초 간격으로 두 번 읽고, 두 시점의 편집 목록과 파일 크기를 비교한다.
- stdout 인코딩은 아래 방법으로 검증한다.
  - 먼저 ASCII만 담은 대조군을 보내, Hook이 실제로 실행되는지 확인한다.
  - 같은 한글을 UTF-8 바이트와 CP949 바이트로 인코딩해 한 출력에 나란히 넣어 보낸다.
  - Claude Code가 받은 문자열의 코드포인트를 각 인코딩으로 디코딩한 결과와 비교한다.
- Hook 호출 비용은 인터프리터 시작 비용과 분리해 측정한다.
  - Hook을 하위 프로세스로 여러 번 실행하고 실제 경과 시간을 기록한다.
  - 같은 횟수만큼 인터프리터만 실행해, 시작에 걸린 시간을 따로 구한다.
  - Hook은 파일 전체에 사전의 모든 정규식을 적용하므로, 검사 시간은 파일 크기와 사전 항목 수의 영향을 받는다.
  - 탐지 건수가 많으면 탐지 결과마다 줄 번호를 계산하고 결과를 정렬하는 비용도 늘어난다.
    - 줄마다 사전의 표현이 들어 있는 파일을 만들어, 이 비용도 함께 측정한다.
