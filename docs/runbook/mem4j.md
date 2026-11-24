# mem4j 런북

`mem4j` CLI와 Neo4j Aura 콘솔에서 스코프, 지침, 리마인더, 백데이터를 조회하고 탐색하는 절차를 정리한다.

## CLI 기본 설정과 도움말

- Neo4j 기본 자격 증명 파일 경로는 `~/.config/mem4j/.secrets/neo4j.txt`다.
  - 다른 경로의 파일을 쓰려면 `-credentials <경로>`를 지정한다.
- 임베딩을 호출하는 명령의 GCP 기본 자격 증명 파일 경로는 `~/.config/mem4j/.secrets/gcp.json`이다.
  - 다른 경로의 파일을 쓰려면 `-gcp-credentials <경로>`를 지정한다.
- 명령 목록과 플래그는 `--help`나 `help` 서브커맨드로 확인한다.

```bash
mem4j --help
mem4j reminder recall --help
```

## CLI로 스코프와 지침 조회

### 전체 스코프 목록과 연관 스코프

```bash
mem4j scope list
mem4j scope neighbors -name "듀오 세션" -depth 1
```

### 스코프 조합에 적용되는 상시 지침

- 여러 스코프에 함께 연결된 `Instruction`을 조회하려면 해당 스코프를 모두 `-scope`로 지정한다.
- 세션 도중 활성 스코프를 누적하려면 `SessionStart`가 알려 준 세션 상태 파일 경로를 `-state`로 함께 넘긴다.

```bash
mem4j instruction recall -scope "<사용자>" -scope "듀오 세션"
mem4j instruction recall -state "<세션 상태 파일 경로>" -scope "코드 작성"
```

### 프롬프트 기반 리마인더 조회

- 활성 스코프 교집합 안에서 프롬프트와 코사인 유사도가 기본 임계값(`0.75`) 이상인 `Reminder`를 조회한다.

```bash
mem4j reminder recall -scope "<사용자>" -scope "GitHub 리포지토리" -query "변경 사항을 커밋하고 PR 올려줘"
```

### 특정 스코프에 연결된 모든 상시 지침

- 다른 스코프와의 결합 여부와 무관하게 특정 스코프에 연결된 `Instruction`을 모두 조회한다.

```bash
mem4j graph query -cypher "MATCH (i:Instruction)-[:APPLIES_IN]->(:Scope {name: '듀오 세션'}) RETURN i.key AS key, i.title AS title, [(i)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes, i.body AS body ORDER BY i.key"
```

## CLI로 백데이터 조회와 탐색

### 엔티티 목록, 의미 검색, 이웃 탐색

```bash
mem4j entity list
mem4j entity recall -query "루꼴라 앱"
mem4j entity neighbors -name "~/.config/mem4j/.secrets/gcp.json" -depth 2
```

### 사실(`Statement`)과 이력(`Event`) 조회

- 특정 `Entity`에 1-hop으로 연결된 항목을 바로 나열하거나, `-query`로 의미 검색을 수행한다.

```bash
mem4j statement list -entity "Recallar 안드로이드 앱"
mem4j statement recall -query "안드로이드 앱에서 토픽 편집할 수 있어?"
mem4j event list -entity "Recallar 안드로이드 앱"
mem4j event recall -query "앱 이름 변경"
```

## 훅 동작 확인과 로그 점검

- 훅은 `~/.config/mem4j/config.json`에서 기본 스코프와 워크스페이스 매핑을 읽는다.
- 샘플 JSON을 stdin으로 넘겨 각 훅의 출력을 직접 확인할 수 있다.

```bash
echo '{"cwd":"C:/Git/claude-workflow-kit","source":"startup"}' | mem4j hook claude-session-start
echo '{"cwd":"C:/Git/claude-workflow-kit","prompt":"응"}' | mem4j hook claude-user-prompt
echo '{"workspacePaths":["C:/Git/claude-workflow-kit"]}' | mem4j hook agy-session-start
echo '{"conversationId":"smoke-1","workspacePaths":["C:/Git/claude-workflow-kit"],"lastUserInput":"변경 사항 커밋하고 PR 올려줘"}' | mem4j hook agy-pre-invocation
```

- 훅 실행 이력과 필터링·유사도 점수는 `~/.config/mem4j/hook.log`에 JSONL 형식으로 누적된다.

```bash
tail -n 20 ~/.config/mem4j/hook.log
```

## Neo4j Aura 콘솔에서 조회

- Aura 콘솔 `Queries` 탭의 Saved Cypher 업로드 기능으로 [mem4j-aura-queries.csv](mem4j-aura-queries.csv)를 올려 사용한다.
- 업로드하면 `mem4j` 폴더 아래에 아래 여덟 쿼리가 등록된다.
  - `전체 스코프 노드 보기`: 등록된 모든 `Scope` 노드를 이름순으로 조회한다.
  - `듀오 세션 지침 표로 보기`: 특정 스코프에 연결된 `Instruction`을 표로 출력한다.
  - `듀오 세션 지침 그래프로 보기`: 특정 스코프에 직접 연결된 `Instruction`만 그래프로 표시한다.
  - `듀오 세션 지침과 전체 스코프 그래프로 보기`: 특정 스코프의 `Instruction`에 연결된 다른 `Scope`까지 함께 그린다.
  - `스코프 조합 적용 지침 보기`: `mem4j instruction recall`과 같은 교집합 조건으로 추린다.
  - `전체 리마인더 표로 보기`: 등록된 모든 `Reminder`와 연결 스코프를 표로 출력한다.
  - `전체 엔티티와 연결 그래프로 보기`: 모든 `Entity` 노드와 `Entity` 간 관계를 그래프로 표시한다.
  - `엔티티의 사실과 이력 표로 보기`: 특정 `Entity`에 1-hop으로 연결된 `Statement`와 `Event`를 표로 출력한다.
