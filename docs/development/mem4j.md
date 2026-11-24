# mem4j 개발

이 문서는 `mem4j`의 개발 전제와 플랫폼 제약을 정리한다. 아키텍처는 [mem4j 아키텍처 문서](../architecture/mem4j.md)에서 다룬다.

## 환경별 접근 방식

- CLI 환경에서는 각 CLI가 제공하는 훅을 사용한다.
- Claude 웹과 Gemini 웹 환경에서는 에이전트가 자율적으로 기억에 접근한다.

## CLI 훅 비교

아래 표는 Claude Code 2.1.328 훅 스펙 문서와 agy 훅 스펙 문서(2026-11-19 확인)에 근거한다.

| 항목                             | Claude Code                                                                                                 | agy                                                                                                                                                         |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SessionStart`에서 CLI 명령 실행 | 가능. `type: "command"`로 셸 명령을 실행하며, 출력은 첫 프롬프트 전에 컨텍스트로 들어간다.                  | 가능. 핸들러는 `type: "command"`만 지원하며, 출력의 `injectSteps`가 세션 시작 시 대화에 삽입된다.                                                           |
| 사용자 프롬프트 제출 시점 훅     | `UserPromptSubmit`: 프롬프트 제출 직후 Claude 처리 전에 실행된다. 컨텍스트 추가와 프롬프트 차단이 가능하다. | 전용 이벤트 없음. `PreInvocation`이 프롬프트 직후를 포함한 매 모델 호출 전에 실행되며, 입력에 `lastUserInput`이 포함된다. 스텝 주입과 호출 중단이 가능하다. |
| 컨텍스트 압축 뒤 `SessionStart`  | `source: "compact"`로 `SessionStart`가 다시 실행된다.                                                       | 다시 실행되지 않는다.                                                                                                                                       |

## 훅 전제와 플랫폼 제약

- Claude Code `additionalContext` 제약
  - 길이가 10,000자를 넘으면 Claude Code가 텍스트를 세션 디렉터리 파일로 저장하고 앞 2,000자 미리보기와 경로만 전달한다.
  - 시스템 명령어 형식의 문구는 프롬프트 인젝션 방어를 유발할 수 있으므로, 훅 출력 머리말은 출처와 사실을 밝히는 서술형으로 작성한다.
  - `~/.claude/CLAUDE.md`와 `~/.gemini/GEMINI.md`에 `<!-- mem4j:start -->`와 `<!-- mem4j:end -->` 관리용 마커 블록으로 `mem4j` 훅 신뢰 선언과 기본 탐색·정정 안내를 배치한다.
- `agy` `PreInvocation` 중복 호출 방지
  - 한 턴 안에서 도구를 여러 번 호출하면 `PreInvocation`도 매 호출마다 실행된다.
  - `agy-pre-invocation`은 세션 상태 파일(`mem4j-state.json`)에 마지막으로 처리한 사용자 입력 위치를 기록해, 같은 턴의 후속 모델 호출에서는 리마인더 조회를 건너뛴다.
  - `agy-pre-invocation`은 압축 뒤에 복원되지 않는 instant 지침 성격에 맞춰 `ephemeralMessage`로 주입한다.
- 세션 상태 파일 경로와 폴백
  - Claude Code는 훅 stdin의 `scratchpad_dir` 아래 `mem4j-state.json`을, `agy`는 `artifactDirectoryPath` 아래 `scratch/mem4j-state.json`을 사용한다.
  - 해당 디렉터리 필드가 비어 있는 환경에서는 `~/.config/mem4j/state/<세션ID>.json`으로 폴백한다.
  - Claude Code는 컴팩션 뒤 `SessionStart`가 다시 실행될 때 상태 파일의 기존 업무 스코프를 그대로 이어받아 상시 지침을 복원한다.
- `agy` 컴팩션 뒤 복원
  - `agy`는 컨텍스트 압축 뒤에 `SessionStart`를 다시 호출하지 않으므로, 압축 복원 시점의 재주입은 기존 `agy-compaction-hook`이나 후속 작업 범위로 분리한다.
- 스코프 선택 방식
  - 훅은 그래프의 `Scope` 링크를 일괄 자동 확장하지 않고, `config.json`과 세션 상태 파일에 명시된 스코프만 활성 스코프로 사용한다.
  - 에이전트는 첫 프롬프트와 작업 전환 시점에 `SessionStart`가 주입한 전체 `Scope` 목록이나 `mem4j scope list`를 확인해 필요한 업무 스코프를 선택한다.
  - 선택한 스코프는 `mem4j instruction recall -state <상태 파일> -scope <이름>`으로 활성화한다.
- 프롬프트 실행 로그
  - 프롬프트 길이 필터(기준은 [아키텍처 문서](../architecture/mem4j.md#훅-연동과-로컬-설정) 참고)와 `Reminder`·`Statement`·`Event` 조회 결과는 `~/.config/mem4j/hook.log`에 JSONL 형식으로 기록한다.
  - 이벤트, 필터 통과 여부, 활성 스코프, 매칭 키 목록, 최고 유사도 점수, 소요 시간, 오류를 남겨 `threshold` 조정 근거로 사용한다.
  - 로그 파일 쓰기가 실패해도 훅 동작은 중단하지 않는다.
- 타임아웃과 장애 허용
  - `mem4j hook`은 내부 실행 타임아웃을 10초로 설정해, 외부 API가 지연되더라도 세션 입력 대기가 길어지지 않게 제한한다.
  - 설정·자격 증명·네트워크 오류가 발생해도 종료 코드 0과 빈 응답을 반환한다.

## 훅 스모크 테스트

- `mem4j hook` 각 서브커맨드에 샘플 stdin JSON을 파이프로 넘겨 출력 JSON 형식을 확인한다.
  - 짧은 입력(`"응"`, `"ping"`)에서 임베딩 호출 없이 빈 응답이 나오는지 확인한다.
  - `agy-pre-invocation`에 같은 `conversationId`와 `lastUserInput`을 두 번 연속 넘겨 두 번째 호출이 건너뛰어지는지 확인한다.
- 실제 Claude Code와 `agy` 세션에서 `SessionStart` 주입 내용이 프롬프트 인젝션 경고 없이 정상 수용되는지 확인한다.
