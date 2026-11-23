- 커밋 없이 브랜치만 push해도 MR을 열 수 있다.
  - draft MR은 작업 단위 확정 단계에서 브랜치를 push한 뒤 연다.

```
glab mr create --target-branch "<origin 기본 브랜치>" --draft -a @me \
  -t "제목" -d "$(cat <본문 파일>)"
```

- 본문은 `glab mr edit`이 아니라 `glab mr update`로 고치고, Ready로 전환할 때에는 `glab mr update --ready`를 실행한다.
- 본문은 파일로 쓴다.
  - 여러 줄 텍스트를 명령줄에 직접 붙여 넣으면 이스케이프 실수가 생기기 쉽다.
  - `glab`은 `--body-file`을 지원하지 않으므로, 파일 내용을 `-d "$(cat <본문 파일>)"` 형태로 넘긴다.
