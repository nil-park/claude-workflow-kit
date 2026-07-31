# 0003. `workflow-core-max`가 대체하는 범위와 남은 스킬의 처리

## 배경

켜고 끄는 단위가 플러그인이라, `workflow-core`를 끄면 그쪽 스킬이 전부 같이 꺼진다.
`-max` 계열([설계](../architecture/max-plugins.md))은 `workflow-core`를 대체하므로,
`review-criteria` 말고 남는 `docs-standards`·`pre-push`·`refs-scratch`를 어떻게 할지,
그리고 나머지 플러그인을 함께 켤 수 있는지 정해야 했다.

## 결정

- **`refs-scratch`는 스킬로 들고 간다.**
- **`docs-standards`는 분해해 에이전트에 나눈다.** 배치·결정 기록 규칙·산문·주석·참조
  방향은 `docs-intent-auditor`가, "문서 셋 이상이 링을 닫으면 안 된다"는
  `structure-auditor`가 가져간다.
- **`pre-push`는 싣지 않는다.**
- **`github-workflow`·`gitlab-workflow`도 함께 끈다.**

## 근거

`refs-scratch`가 스킬로 남는 이유는 **발동 주체가 메인 에이전트**이기 때문이다. 라운드
채널이 이 규약 위에 서 있는데 파일을 쓰는 것은 메인뿐이라, 읽기 전용인 에이전트 쪽으로
흩을 수가 없다.

`docs-standards`는 반대다. 그 내용을 실제로 쓰는 것은 문서와 주석을 판정하는 에이전트이고,
분해하면 `review-criteria` → `docs-standards` 강결합도 함께 사라진다.

`pre-push`는 리뷰와 무관하다. 리포지토리당 한 번 훅을 깔면 끝나는 설치용 스킬이라 없어서
아쉬울 순간이 거의 없고, 필요하면 그때 `workflow-core`를 잠깐 켜면 된다.

`github-workflow`·`gitlab-workflow`를 끄는 것은 선택이 아니라 의존성의 결과다. 둘 다
`workflow-core`에 의존해서, 켜는 순간 `workflow-core`가 따라 켜지고 `review-criteria`가
둘이 된다. 그래서 1단계 동안은 브랜치·PR·MR 워크플로 스킬 없이 리뷰만 쓴다.

## 버리는 것

`docs-standards`를 분해하면서 **문서를 쓸 때 쓰는 쪽 기능을 잃는다** — "이 문서 어디에
둬야 해?", "ADR 추가". 판정 기준은 에이전트 안에 남지만, 메인이 문서를 **쓰는 시점에는**
규약을 모르는 채로 쓰고 나중에 지적받는 구조가 된다.

1단계 목표가 `review-criteria`와의 동등이고 문서 작성은 애초에 `review-criteria`의 역할이
아니었으므로 감수한다. 되살릴지는
[#4](https://github.com/nil-park/claude-workflow-kit/issues/4)에서 판단한다.

두 계열 사이에 동기화 장치도 두지 않는다. 같은 규약이 양쪽에 각각 존재하게 되지만, 따로
진화할 수 있는 쪽을 택했다.

## 검토한 대안

| 대안                                   | 왜 안 골랐나                                                                                                              |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| 세 스킬을 그대로 복사해 싣는다         | `docs-standards`는 판정 기준이 에이전트에 있어야 축별로 나뉜다. 스킬로 두면 같은 기준이 두 곳에 생긴다                    |
| 심링크로 원본을 공유한다               | 마켓플레이스 내 심링크는 캐시로 복사될 때 역참조되지만, Windows에서 `core.symlinks=false`면 clone 시 텍스트 파일로 깨진다 |
| `workflow-core`에 의존성을 건다        | 의존하면 `workflow-core`가 함께 켜지고 `review-criteria`가 둘이 된다                                                      |
| `review-criteria`만 담은 얇은 플러그인 | 나머지 스킬을 쓰려면 `workflow-core`를 켜야 하고, 결국 위와 같은 중복이 생긴다                                            |
