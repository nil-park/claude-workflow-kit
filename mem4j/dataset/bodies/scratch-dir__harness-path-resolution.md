1. 하네스가 경로를 지정하면 그 경로를 사용한다.
2. 하네스가 경로를 지정하지 않으면 다음 경로에 디렉터리를 만든다.
   - `<OS-temp>/agent/<project-slug>/<timestamp>/`
   - `project-slug`: 리포지토리 루트의 절대 경로에서 경로 구분자와 `:`를 `-`로 바꾼 값(예: `C:\Git\my-repo` → `C--Git-my-repo`)
   - `timestamp`: 세션을 시작한 시각을 `YYYYMMDD-HHMMSS` 형식으로 적은 값
