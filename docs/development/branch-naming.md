# 브랜치 네이밍 규칙

## 기본 형식

```
{type}/{issue-number}-{brief-description}
```

## 타입

| 타입        | 설명                  | 예시                                      |
| ----------- | --------------------- | ----------------------------------------- |
| `feature/`  | 새로운 기능 추가      | `feature/42-add-user-authentication`      |
| `fix/`      | 버그 수정             | `fix/43-fix-memory-leak`                  |
| `hotfix/`   | 긴급 수정 (프로덕션)  | `hotfix/44-patch-critical-security-issue` |
| `refactor/` | 코드 리팩토링         | `refactor/45-simplify-database-layer`     |
| `docs/`     | 문서 작업             | `docs/46-update-api-documentation`        |
| `chore/`    | 의존성, 설정, 빌드 등 | `chore/47-update-dependencies`            |
| `test/`     | 테스트 관련 작업      | `test/48-add-unit-tests`                  |
| `perf/`     | 성능 개선             | `perf/49-optimize-database-queries`       |

## 규칙

- 영어 사용
- 간단한 유지보수성 수정 또는 hotfix는 예외적으로 이슈 번호 생략 가능
