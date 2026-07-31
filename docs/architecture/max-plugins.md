# `-max` 플러그인 설계

토큰 여유는 많고 작업 속도가 아쉬운 환경(회사 Max 20 플랜)을 위한 변종. 리뷰를 여러
전문 서브 에이전트에 병렬로 흩뿌려 놓치는 것을 줄이고, 그 대가로 토큰을 아끼지 않는다.
기존 `workflow-core`와 **독립**이며, 둘을 같이 켜지 않는다. 스킬과 에이전트 본문은
한국어로 쓴다 — [0001-max-plugins-in-korean.md](../decisions/0001-max-plugins-in-korean.md).

`github-workflow`·`gitlab-workflow`도 함께 끈다. 둘 다 `workflow-core`에 의존해서, 켜는
순간 `workflow-core`가 따라 켜지고 `review-criteria`가 둘이 된다. 그래서 1단계 동안은
브랜치·PR·MR 워크플로 스킬 없이 리뷰만 쓴다.

```mermaid
flowchart TB
  subgraph core["workflow-core-max"]
    M["메인 에이전트"]
    RC["review-criteria<br/>(팬아웃 오케스트레이션)"]
    A1["항상 도는 리뷰어 7"]
    A2["조건부 리뷰어 1"]
  end

  U([사용자]) <--> M
  M --> RC
  RC --> A1 & A2
  A1 & A2 -.보고.-> RC
  RC --> L[(".refs/review/…<br/>라운드 대장")]
```

## 단계

**1단계는 기존 `workflow-core:review-criteria`와 동등한 역할을 해내는 것까지다.**
판정 기준을 늘리지 않고, 같은 기준을 병렬로 더 촘촘히 적용하는 데만 집중한다. 실제로
써보고 잘 돌아가는 것을 확인한 뒤에 [확장 후보](#확장-후보)로 넘어간다.

산출물은 `workflow-core-max` 플러그인 하나다.

## 플러그인 경계

```
plugins/workflow-core-max/
  .claude-plugin/plugin.json
  skills/
    review-criteria/    팬아웃 오케스트레이션
    refs-scratch/       .refs/ 규약
  agents/               리뷰어 8종
```

`workflow-core`의 `docs-standards`는 스킬로 남지 않고 `docs-intent-auditor`·
`structure-auditor`로 분해돼 들어갔고, `pre-push`는 싣지 않는다. 그래서 문서를 **쓸 때**
쓰는 기능은 이 계열에 없다 — 판정 기준은 에이전트 안에만 있다.

스킬 사이 참조는 [skill-references.md](skill-references.md), 이렇게 고른 근거는
[0003](../decisions/0003-max-plugin-composition.md).

## 리뷰 팬아웃

### 대상

기존 `review-criteria`와 같다. 호스팅된 PR/MR, 브랜치, 커밋 전 워킹 디렉터리의 변경분
모두가 대상이고, 마지막 것이 자기리뷰다. 호스팅된 것은 `gh`/`glab`로 diff·설명·논의를
먼저 끌어온 뒤 코드는 로컬 체크아웃에서 읽는다.

수집은 메인이 **한 번만** 한다. 에이전트마다 다시 끌어오면 같은 것을 여러 번 가져오는
데다, 사이에 원격이 바뀌면 축마다 다른 변경을 보게 된다.

### 조율

에이전트끼리는 대화하지 않는다. 메인이 축마다 하나씩 띄우고 혼자 취합한다
([0002](../decisions/0002-parallel-review-fanout.md)).

취합의 절반은 **검증**이다. 에이전트는 없는 문제를 지어내므로, 메인은 인용된 `file:line`을
직접 열어 확인하고 반증을 시도해 살아남은 것만 대장에 올린다.

### 기준의 위치

메인이 에이전트에 대해 아는 것은 **정의의 `description`과 자기가 넘긴 프롬프트뿐**이다.
정의 본문은 에이전트의 시스템 프롬프트로 들어갈 뿐 메인에게 보이지 않는다. 그래서 메인은
투입할 에이전트를 고를 때도, 무엇이 검사되지 않았는지 알 때도, 사용자에게 적용 기준을
설명할 때도 본문에 기댈 수 없다.

따라서 **"무엇을"은 스킬, "어떻게"는 에이전트**로 가른다.

| 위치                          | 담는 것                                                           |
| ----------------------------- | ----------------------------------------------------------------- |
| `review-criteria`의 팬아웃 표 | 각 에이전트의 책임 범위 — 무엇을 보고, 무엇을 **안 보는지**       |
| 에이전트 `description`        | 같은 범위를 한 문장으로. 메인에게 목록으로 주입되는 이중 안전장치 |
| 에이전트 본문                 | 그 범위 안에서 어떻게 깊게 파는지 — 기법, 체크리스트, 출력 포맷   |

여기에 런타임 안전장치를 하나 건다. **각 에이전트는 보고 말미에 자신이 적용한 기준과
범위 밖이라 보지 않은 것을 함께 반환한다.** 메인은 정의 파일을 읽지 않고도 실제 검사
범위를 확인할 수 있고, 정의가 바뀌었는데 스킬 표가 따라오지 않은 드리프트도 여기서
드러난다.

### 리뷰어 (1단계)

기존 `review-criteria`의 판정 항목을 그대로 여덟으로 분해한 것이다. 새 기준을 더하지
않는다.

| 에이전트                    | 축                          |
| --------------------------- | --------------------------- |
| `correctness-hunter`        | 잠재 정확성 버그            |
| `resource-auditor`          | 누수와 낭비                 |
| `silent-failure-hunter`     | 문제를 숨기는 억제          |
| `test-auditor`              | 지키지 못하는 테스트        |
| `hygiene-auditor`           | 작업 흔적과 주변 관례 이탈  |
| `docs-intent-auditor`       | 말과 코드의 어긋남          |
| `structure-auditor`         | 변경이 그은 참조 선         |
| `client-resilience-auditor` | 외부 호출의 복원력 (조건부) |

위는 축 이름뿐이고, 각 에이전트의 실제 책임 범위는 `review-criteria` 스킬의 팬아웃 표에
있다. 거기서 "안 본다"가 "본다"만큼의 자리를 차지한다 — 축이 겹치면 같은 지적이 여러 번
올라와 메인의 검증 비용만 배로 늘기 때문이다.

조건부인 `client-resilience-auditor`는 HTTP/RPC/외부 서비스 클라이언트가 변경될 때만
투입한다.

`docs-intent-auditor`는 채우기 쉬운 축이라 잡소리가 잘 나온다. 정의 본문에서 "읽기 정말
어렵거나 곧 낡을 것만 지적하라"를 못 박는다.

### 라운드 채널

`.refs/` 아래 파일로 한다.

```
.refs/review/<작업-슬러그>/
  r1/, r2/, …      라운드마다 하나씩
    <agent>.md     각 에이전트가 낸 원본 보고
    verdict.md     메인이 검증을 마치고 확정한 라운드 결과
  ledger.md        라운드를 넘어 지속되는 지적 대장 (open/fixed/rejected/deferred)
```

**파일은 메인이 쓴다.** 에이전트는 읽기 전용이라 소스를 건드릴 수 없다. `ledger.md`가 있어야
3라운드에서 기각한 지적을 5라운드가 다시 올릴 때 알아본다.

채널 형태를 파일로 정한 근거와 기각한 대안은
[0002](../decisions/0002-parallel-review-fanout.md)에 있다.

## 확장 후보

1단계를 실제로 써본 뒤에 손댈 목록은 이슈로 옮겼다.

- 리뷰어 확장 후보: [#3](https://github.com/nil-park/claude-workflow-kit/issues/3)
- 작업 사이클 일반화와 `gitlab-workflow-max`: [#4](https://github.com/nil-park/claude-workflow-kit/issues/4)
