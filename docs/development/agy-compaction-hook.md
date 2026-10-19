# agy-compaction-hook 개발

이 문서는 agy-compaction-hook의 전제와 검증 방법을 정리한다. 아키텍처 구조는 [agy-compaction-hook 아키텍처 문서](../architecture/agy-compaction-hook.md)를 참고한다. 외부 규격 원본은 [agy 규격 문서](../setup/agy-docs.md)에서 확인한다.

## 전제

- 항상 최신 agy 버전을 사용한다고 전제하며, 아래 관찰 결과는 1.4.2(2026-10)에서 확인했다.
- agy 훅 규격에는 컨텍스트 압축 전용 이벤트가 없다.
- 컨텍스트 압축 시 세션 기록에 스텝 1개가 남는다.
  - `"type": "CHECKPOINT"` 값을 가지며 `content`는 `# Resuming from a compaction`으로 시작한다.
  - IDE 환경에서도 동일한 형태의 스텝이 남는다.
- `step_index`와 `initialNumSteps`는 압축 이후에도 계속 증가하므로 압축 발생을 나타내지 못한다.
- 컨텍스트 압축은 `PreInvocation` 실행 이후, 모델 호출 이전에 수행된다.
  - 압축 직후 첫 모델 호출은 재열람 메시지 없이 실행되며, 해당 호출의 `PostInvocation`이 체크포인트를 확인한다.
  - 메시지 없는 `/compact`는 모델 호출과 `PostInvocation` 없이 끝나므로, 다음 턴의 `PreInvocation`이 체크포인트를 확인한다.
- 턴의 마지막 응답 이후 `PostInvocation`이 주입한 `systemMessage`는 추가 모델 호출을 시작하지 않으며 모델은 다음 턴에서 이를 읽는다.
- 주입된 `systemMessage`는 `"source": "SYSTEM_SDK"`, `"type": "SYSTEM_MESSAGE"` 스텝으로 세션 기록에 저장된다.
  - 해당 스텝의 `content`는 agy가 붙이는 서두 문구 뒤에 주입된 본문을 담는다.
  - 다음 훅 실행 시점(같은 턴에서 도구 호출 뒤에 이어지는 `PreInvocation` 포함)에는 이미 세션 기록에 포함되어 있다.
- `PostInvocation`은 해당 모델 응답에 따른 도구 호출이 모두 끝난 뒤에 실행된다.
- 식별 표식은 주입된 스텝이 아닌 다른 스텝에도 나타날 수 있다.
  - 체크포인트 요약문이 이전 재열람 메시지를 인용할 수 있다.
  - 모델 응답이나 도구 실행 결과가 주입된 문구를 그대로 반복할 수 있다.
- `~/.gemini/antigravity-cli/settings.json` 내부의 훅 설정은 형식이 공개되지 않아 사용하지 않는다.
- `AGENTS.md`, `GEMINI.md`, 규칙, 에이전트 정의는 시스템 프롬프트에 포함되며 압축 이후에도 유지된다.
- 세션 기록 파일 크기는 최대 수 MB 수준이므로 매 호출마다 파일 전체를 읽는다.

## 검증 방법

- 단위 테스트: `cd agy-compaction-hook && go test ./...` 실행 (`make test`에도 포함)
- 임시 작업 공간에서의 스모크 테스트 (전역 `hooks.json`은 수정하지 않는다)
  1. 바이너리를 빌드하고 `<workspace>/.agents/hooks.json`의 `PreInvocation` 및 `PostInvocation`에 등록한다.
  2. 작업 공간에서 `agy -p "<prompt>" --output-format json`을 실행하고 `conversation_id`를 확인한다.
  3. `agy --conversation <id> -p "/compact <prompt>"` 명령을 실행한다.
     - Git Bash에서는 `MSYS_NO_PATHCONV=1`을 설정하여 `/compact`가 파일 경로로 변환되지 않도록 한다.
  4. 세션 기록에서 응답 뒤에 식별 표식을 포함한 `SYSTEM_MESSAGE` 스텝이 이어지는지 확인한다.
  5. 턴을 1회 더 진행한 뒤 훅이 두 번째 재열람 메시지를 추가하지 않는지 확인한다.
  6. 모델이 셸 도구를 호출하도록 유도하는 프롬프트로 `/compact <prompt>`를 실행하고 재열람 메시지가 1회만 나타나는지 확인한다.
- 설치 후 IDE 검증
  - 이미 체크포인트가 존재하는 IDE 대화를 재개하여 메시지를 전송한다.
  - 해당 대화의 세션 기록에 식별 표식을 포함한 `SYSTEM_MESSAGE` 스텝이 나타나는지 확인한다.

## 위험 요소

- agy의 세션 기록 형식이나 체크포인트 식별 문자열이 변경되면 오류 없이 감지가 중단된다.
- agy가 세션 기록에 주입된 스텝을 남기지 않도록 변경되면 호출마다 재열람 메시지가 반복 주입된다.
