- **인코딩 미지정**: 인코딩 없는 파일 입출력, charset 없는 바이트↔문자열 변환과 Content-Type,
  한글을 `\uXXXX`로 escape하는 JSON 직렬화
  - 한글은 NFC로 정규화한다.
    - macOS는 NFD를 만든다.
