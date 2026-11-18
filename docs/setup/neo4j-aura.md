# Neo4j Aura 설정

AuraDB Free 인스턴스를 만들고 Claude Code에 Aura MCP 서버를 연결한다.

- 기준 문서: https://neo4j.com/docs/mcp/current/mcp-for-aura/

## 설정 절차

1. https://console.neo4j.io/ 에서 계정을 만든다.
2. AuraDB Free 인스턴스를 생성한다.
   - 생성할 때 내려받는 자격증명 파일에 비밀번호가 들어 있으므로 저장소 밖에 보관한다.
3. 상단 바에서 조직(organization)으로 들어간 뒤 왼쪽 내비게이션의 Security에서 "Tool authentication with Aura user"를 켠다.
4. 인스턴스의 MCP 서버 URL을 확인한다.
   - 형식은 `https://<INSTANCE_ID>.mcp-instances.neo4j.io`다.
   - 콘솔에서는 Instances → 인스턴스 더보기 메뉴 → Inspect에서 확인할 수 있다.
   - 접속 URI(`neo4j+s://...`)나 Query API URL과는 다르다.
5. Claude Code에 MCP 서버를 등록한다.

   ```bash
   claude mcp add --transport http neo4j-mcp https://<INSTANCE_ID>.mcp-instances.neo4j.io
   ```

   - 기본 스코프는 `local`이며, 현재 저장소에만 적용되고 `~/.claude.json`에 저장된다.

6. Claude Code에서 `/mcp`를 실행하고 Aura 콘솔 계정으로 로그인한다.

## 제공되는 도구

연결을 마치면 아래 세 도구를 사용할 수 있다.

- `get-schema`: 레이블, 관계 타입, 프로퍼티 키로 구성된 그래프 스키마를 조회한다.
- `read-cypher`: 읽기 전용 Cypher 쿼리를 실행한다.
  - 쓰기 쿼리와 `dbms.components()` 같은 관리자 프로시저는 거부한다.
- `write-cypher`: 데이터를 변경하는 쓰기 Cypher 쿼리를 실행한다.
