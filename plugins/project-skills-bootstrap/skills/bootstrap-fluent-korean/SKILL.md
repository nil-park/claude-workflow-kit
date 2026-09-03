---
name: bootstrap-fluent-korean
description: >-
  fluent-korean 출력 스타일을 현재 프로젝트에 설치하고 싶을 때 부른다.
---

`.claude/settings.json`에 `enabledPlugins["fluent-korean@fluent-korean"]`가 없을 때만 실행한다.

1. `.claude/settings.json`을 읽는다. 파일이 없으면 빈 객체 `{}`로 시작한다.
2. 아래 항목들을 추가해 파일을 쓴다.

   ```json
   {
     "extraKnownMarketplaces": {
       "fluent-korean": {
         "source": { "source": "github", "repo": "snflkd/fluent-korean" },
         "autoUpdate": true
       }
     },
     "enabledPlugins": {
       "fluent-korean@fluent-korean": true
     },
     "outputStyle": "fluent-korean:fluent-korean"
   }
   ```

3. 설치 결과를 보고하면서 아래 두 가지를 함께 안내한다.
   - `extraKnownMarketplaces`는 팀원이 폴더를 trust하기 전까지 적용되지 않는다.
   - `outputStyle` 변경은 `/clear` 실행 후 또는 새 세션을 시작해야 적용된다.
