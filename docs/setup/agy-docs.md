# agy 문서

이 리포지토리의 agy 지원은 agy의 규칙과 스킬 스펙에 의존한다. 규칙과 스킬 스펙의 SoT는 아래 공식 문서다.

## 원문을 받아서 읽는다

공식 문서 페이지는 HTML로 열린다. URL 끝에 `.md`를 붙이면 마크다운 원문을 받는다.
응답은 gzip으로 압축되어 온다. 그래서 `curl`에 `--compressed`를 붙인다.

```bash
curl -sL --compressed --create-dirs https://antigravity.google/docs/rules.md -o .spec-docs/agy-rules-original-$(date +%y%m%d).md
```

받은 파일은 `.spec-docs/` 아래에 저장한다. 파일명 끝의 `YYMMDD`가 확인 시점이다.

`.spec-docs/`에 받은 지 15일이 지나지 않은 원문이 있으면, 사용자가 다시 받으라고 지시하기 전에는 새로 받지 않고 그 파일을 읽는다.

## 문서 목록

| 주제         | URL                                              | 볼 때                                                |
| ------------ | ------------------------------------------------ | ---------------------------------------------------- |
| 규칙         | https://antigravity.google/docs/rules.md         | `AGENTS.md`를 고칠 때                                |
| 스킬         | https://antigravity.google/docs/skills.md        | 스킬 디렉터리 구조나 프런트매터를 확인할 때          |
| 훅           | https://antigravity.google/docs/hooks.md         | `hooks.json`의 위치, 이벤트, 입출력 필드를 확인할 때 |
| 플러그인     | https://antigravity.google/docs/plugins.md       | 플러그인 구조와 설치 절차를 확인할 때                |
| CLI 레퍼런스 | https://antigravity.google/docs/cli/reference.md | 플래그와 슬래시 명령을 확인할 때                     |
| 변경 기록    | `agy changelog`                                  | 설치한 버전에서 바뀐 동작을 확인할 때                |

- agy 내장 가이드가 규칙 문서로 안내하는 `rules-workflows`에는 `.md` 원문이 없다.
  - 규칙은 `rules.md`에서 읽는다.
- agy에 내장된 문서는 `~/.gemini/antigravity-cli/builtin/skills/agy-customizations/docs/`에 있다.
  - 내장 문서는 웹 문서보다 간략해서 규칙 프런트매터의 제약을 다루지 않는다.
