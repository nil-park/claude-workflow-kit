# 남은 작업

`refactor/rework-workflow-plugins`(PR #10)에서 이어서 할 것. 위에서부터 값이 큰 순서다.

## -max 걷어내기

- `plugins/workflow-core-max/`, `plugins/gitlab-workflow-max/` 삭제
- `.claude-plugin/marketplace.json`에서 두 항목 제거
- `README.md`의 「`-max` 계열」 절과 충돌 요소 표의 관련 행 정리
- `docs/architecture/max-plugins.md` 삭제, `skill-references.md`의 「`-max` 계열」 절 제거
- ADR은 append-only이므로 0002·0003을 고치지 말고 **폐기 결정을 새 ADR로 쓴다.** 실패 원인(무한 라운드, 무기억 재팬아웃, 오케스트레이터 컨텍스트 고갈, 축 세분화가 곧 노이즈 배수)을 근거로 남긴다

## `work-cycle` 다듬기

- **전달 스킬과의 상하 관계를 정한다.** 지금은 `pr-workflow`·`mr-workflow`가 `work-cycle`을 이름으로 부르지 않아 골격이 둘로 갈린다. 종착점도 다르다 — `work-cycle`은 보고서, `pr-workflow`는 `gh pr ready`
- description 9행의 "`mr-workflow` 같은 전달 스킬이 위임할 때도 쓴다"는 -max 시절 전제다. 위임하는 쪽을 만들거나, 이 문장을 뺀다
- **커밋 단위 라운드의 종료 조건을 따로 정한다.** 「게이트와 기록」이 종료 조건을 하나로 묶어 두어, 커밋을 잘게 쪼개라는 3단계와 곱해지면 커밋 10개짜리 PR이 30라운드가 된다
- **대화형 미세 편집의 처리를 정한다.** "한 줄 개정도 대상"이라 탈출구가 막혀 있는데, 실제 사용은 "이거 빼고", "83행만 고치자" 같은 짧은 왕복이 대부분이다
- **보고 후 추가 지시가 들어왔을 때**가 정의돼 있지 않다. 새 사이클인지 같은 사이클의 라운드 재개인지
- 4단계 문구 통일 — 본문은 "제시할 항목", 표는 "보고할 항목"

## 참조·규약

- `docs/architecture/skill-references.md`에 `work-cycle` 노드 추가. `review-criteria`·`docs-standards`·`refs-scratch` 셋 다 이름을 직접 부르는 실선이라, `refs-scratch`를 점선 전용으로 서술한 36–38행도 함께 손본다
- `refs-scratch`에 **리포지토리 밖 분기**를 넣는다. 지금은 `<repo-root>/.refs/`뿐이라, `work-cycle`이 리포지토리 밖 파일을 다룰 때 위임이 빈손으로 끝난다
- 이슈 [#9](https://github.com/nil-park/claude-workflow-kit/issues/9) 잔여 — `pr-workflow`와 `mr-workflow`의 `--body-file` 예시에 박힌 `.refs/` 경로. -max를 걷어내면 이 둘만 남는다
- `README.md`의 플러그인 표에 `work-cycle` 추가

## 한국어 전환

- 스킬·에이전트 본문을 한국어로 옮긴다. [ADR 0001](docs/decisions/0001-max-plugins-in-korean.md)이 "기존 계열은 영어로 둔다"로 정해 두었으므로, 범위를 넓히는 결정을 새 ADR로 쓴다
- `CLAUDE.md`의 언어 규약(커밋 메시지·PR 제목/본문 영어)을 유지할지 함께 정한다
