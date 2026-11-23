# mem4j 런북

`mem4j` CLI와 Neo4j Aura 콘솔에서 스코프와 지침을 조회하는 절차를 정리한다.

## CLI로 스코프와 지침 조회

- 기본 자격 증명 파일 경로는 `~/.config/mem4j/.secrets/neo4j.txt`다.
  - 다른 경로의 파일을 쓰려면 `-credentials <경로>`를 지정한다.

### 전체 스코프 목록

```bash
mem4j scope list
```

### 스코프 조합에 적용되는 지침

- 여러 스코프에 함께 연결된 지침을 보려면 해당 스코프를 모두 `-scope`로 지정한다.

```bash
mem4j instruction recall -scope "<사용자>" -scope "듀오 세션"
```

### 특정 스코프에 연결된 모든 지침

- 다른 스코프와의 결합 여부와 무관하게 특정 스코프에 연결된 지침을 모두 조회한다.

```bash
mem4j graph query -cypher "MATCH (i:Instruction)-[:APPLIES_IN]->(:Scope {name: '듀오 세션'}) RETURN i.key AS key, i.title AS title, [(i)-[:APPLIES_IN]->(t:Scope) | t.name] AS scopes, i.body AS body ORDER BY i.key"
```

## Neo4j Aura 콘솔에서 스코프와 지침 조회

- Aura 콘솔 `Queries` 탭의 Saved Cypher 업로드 기능으로 [mem4j-aura-queries.csv](mem4j-aura-queries.csv)를 올려 사용한다.
- 업로드하면 `mem4j` 폴더 아래에 아래 다섯 쿼리가 등록된다.
  - `전체 스코프 노드 보기`: 등록된 모든 `Scope` 노드를 이름순으로 조회한다.
  - `듀오 세션 지침 표로 보기`: 특정 스코프에 연결된 `Instruction`을 표로 모아 본다.
  - `듀오 세션 지침 그래프로 보기`: 특정 스코프에 직접 연결된 `Instruction`만 그래프로 표시한다.
  - `듀오 세션 지침과 전체 스코프 그래프로 보기`: 특정 스코프의 `Instruction`에 연결된 다른 `Scope`까지 함께 그린다.
  - `스코프 조합 적용 지침 보기`: `mem4j instruction recall`과 같은 교집합 조건으로 추린다.
