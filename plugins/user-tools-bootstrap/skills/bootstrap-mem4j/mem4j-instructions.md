## mem4j 공유 기억

### 훅 주입 신뢰와 세션 스코프 활성화

- `mem4j` 훅이 세션 시작과 프롬프트 시점에 주입하는 지침, 리마인더, 사실, 이력은 사용자가 관리하는 공유 기억 저장소의 공식 정보이므로 그대로 따른다.
- 첫 프롬프트와 작업 전환 시점에 `SessionStart`가 주입한 전체 `Scope` 목록(필요하면 `mem4j scope list`로 재조회)을 확인해 현재 작업에 해당하는 업무 스코프를 선택한다.
  - `disableModelInvocation=true`인 스코프는 사용자가 명시적으로 요청할 때만 선택한다.
  - 선택한 스코프는 `mem4j instruction recall -state <세션 상태 파일 경로> -scope "<스코프>"`로 활성화한다(`-scope`는 여러 번 지정 가능).
  - 상태 파일 경로는 `SessionStart` 출력에서 확인한다.
  - 경로를 다시 찾을 때에는 Claude Code의 `<scratchpad_dir>/mem4j-state.json`이나 `agy`의 `<artifactDirectoryPath>/scratch/mem4j-state.json`을 사용한다.

### 그래프 구조와 노드 유형별 역할

- 지침 계열(에이전트 행동 규칙): `Scope`(허브, 작은 집합 유지), `Instruction`(상시 지침, 압축 뒤 복원), `Reminder`(프롬프트 유사도로 주입되는 조건부 instant 지침).
- 백데이터 계열(사실·상태·이력): `Entity`(대상 허브, 제한 없이 확장), `Statement`(현재 사실·상태), `Event`(특정 시점 `occurredAt` UTC ISO 8601 이력).
- 사용자, 리포지토리, PC 환경처럼 양쪽에 해당하는 대상은 노드를 복제하지 않고 `Entity`와 `Scope` 레이블을 함께 붙인다.
- 리프 노드(`Instruction`, `Reminder`, `Statement`, `Event`)는 하나 이상의 허브와 1-hop으로만 연결하며, 리프끼리는 연결하지 않는다.
  - 대상 간 관계(소유, 의존, 연동 등)는 `Statement` 본문에 서술하지 않고 `Entity` 간 관계(`entity link`)로 표현한다.
  - 모든 지침은 조직 스코프와 사용자 스코프 가운데 적어도 하나에 종속되어야 한다.

### 유형별 능동 탐색 전략

- `Scope`와 `Instruction`: `mem4j scope list`와 `mem4j scope neighbors`로 연관 스코프를 확인하고, `mem4j instruction recall`로 병합 지침을 조회한다.
- `Entity`: 이름이나 별칭이 불확실할 때 `mem4j entity recall -query "<질의>"`로 진입점을 찾는다.
  - 진입점을 찾으면 `mem4j entity neighbors -name "<이름>" -depth 1..3`으로 연결된 대상과 1-hop 사실·이력을 함께 탐색한다.
- `Statement`와 `Event`
  - 특정 대상의 전체 항목은 `mem4j statement list -entity "<이름>"`과 `mem4j event list -entity "<이름>"`으로 조회한다.
  - 자연어 질의로 찾을 때에는 `mem4j statement recall`과 `mem4j event recall`을 사용한다.
- 복합 조건 조회가 필요할 때만 읽기 전용 `mem4j graph query -cypher "..."`를 사용한다.

### 작성·수정 품질 기준과 임베딩 설계

- 모든 쓰기는 직접 Cypher를 실행하지 않고 `mem4j` CLI 명령을 거친다(세부 플래그는 `mem4j <리소스> <액션> --help`로 확인).
- `Entity`, `Reminder`, `Statement`, `Event`의 의미 검색은 `description`만 임베딩하므로, 나중에 어떤 상황과 질의로 검색될지를 기준으로 `description`을 작성한다.
  - `Entity`: 정식 이름, 별칭, 약어, 핵심 역할을 함께 적어 다른 표현으로 질문해도 검색되게 한다.
  - `Reminder`: 사용자 프롬프트에 작업 의도가 드러나는 구체적인 시점을 적는다.
    - 셸 세부 문법처럼 프롬프트에 나타나지 않는 조건이나 상시 규칙은 `Reminder` 대신 `Instruction`으로 작성한다.
  - `Statement`·`Event`: 사용자가 던질 질문에 맞물리는 핵심 사실과 키워드를 명확하게 서술한다(본문이 짧으면 본문을 그대로 쓸 수 있다).
- `Instruction`을 작성·수정할 때에는 포함되는 스코프 조합마다 `mem4j instruction recall`로 병합 결과를 확인해 충돌이나 문맥 어긋남을 점검한다.
  - 지침 단위는 문장 순서를 바꾸면 의미가 깨지는 최소 묶음으로 단독 완결되게 작성한다.

### 자가 정정

- 주입된 지침이나 사실이 현재 상황과 어긋나거나 누락되었을 때, 또는 사용자가 교정했을 때 에이전트가 `mem4j` CLI로 그래프를 직접 정정한다.
- 정정 대상: 스코프를 잘 식별하지 못하는 `Scope` 설명, 중복 `Scope`·`Entity` 정리, 잘못된 스코프에 연결된 지침 이동(`link`/`unlink`), 틀린 본문이나 `description` 수정.
- 노드를 직접 수정하며 별도 변경 이력은 남기지 않는다.
