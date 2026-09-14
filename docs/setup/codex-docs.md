# Codex 문서

Codex의 스킬·플러그인·마켓플레이스·훅·지침 스펙은 아래 공식 문서가 SoT다. 이 리포지토리의 플러그인을
Codex에서 쓰거나 Codex 호환성을 확인할 때 이 문서들을 참조한다.

## 원문을 받아서 읽는다

원문은 [Claude Code 문서](claude-docs.md#원문을-받아서-읽는다)와 같은 방식으로 `curl`로 받는다. Codex
문서는 페이지 URL 끝에 `.md`를 붙인 주소에서 마크다운 원문을 제공한다.

```bash
curl -sL https://developers.openai.com/codex/hooks.md -o .refs/codex-hooks-original-$(date +%y%m%d).md
```

## 문서 목록

| 주제          | URL                                                     | 볼 때                                                |
| ------------- | ------------------------------------------------------- | ---------------------------------------------------- |
| 스킬          | https://developers.openai.com/codex/skills.md           | Codex가 스킬을 읽는 위치나 메타데이터를 확인할 때    |
| 마켓플레이스  | https://developers.openai.com/codex/plugins/build.md    | `marketplace.json`의 위치와 필드를 확인할 때         |
| 플러그인      | https://developers.openai.com/codex/plugins/build.md    | `plugin.json` 필드나 컴포넌트 경로가 헷갈릴 때       |
| 플러그인 사용 | https://developers.openai.com/codex/plugins.md          | 플러그인 설치·제거 절차와 권한 동작을 확인할 때      |
| 훅            | https://developers.openai.com/codex/hooks.md            | `hooks.json`의 위치, 이벤트, 입출력 필드를 확인할 때 |
| 지침          | https://developers.openai.com/codex/guides/agents-md.md | `AGENTS.md`를 읽는 위치와 순서를 확인할 때           |
