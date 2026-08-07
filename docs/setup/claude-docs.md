# Claude Code 문서

이 리포는 Claude Code의 스킬·플러그인·마켓플레이스·훅 스펙 위에서 동작한다. 그 스펙의 SoT는
아래 공식 문서다.

## 원문을 받아서 읽는다

WebFetch는 원문을 주지 않는다. URL을 가져온 뒤 작은 모델이 질문에 답하고, 돌아오는 것은 그
답변뿐이라 문서에 있는 필드가 없다고 나올 수 있다. 원문을 보려면 `curl`로 받는다.

```bash
curl -sL https://code.claude.com/docs/en/hooks.md -o .refs/hooks-original-$(date +%y%m%d).md
```

받은 파일은 `.refs/` 아래에 둔다. 파일명 끝의 `YYMMDD`가 확인 시점이다.

## 문서 목록

| 주제         | URL                                                    | 볼 때                                       |
| ------------ | ------------------------------------------------------ | ------------------------------------------- |
| 스킬         | https://code.claude.com/docs/en/skills.md              | `plugins/*/skills/` 아래를 고칠 때          |
| 마켓플레이스 | https://code.claude.com/docs/en/plugin-marketplaces.md | 카탈로그에 손대거나 배포 방식을 바꿀 때     |
| 플러그인     | https://code.claude.com/docs/en/plugins-reference.md   | 매니페스트 필드나 컴포넌트 경로가 헷갈릴 때 |
| 훅           | https://code.claude.com/docs/en/hooks.md               | 훅을 만들거나 고칠 때                       |
