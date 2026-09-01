# 구성요소 의존 관계

```mermaid
flowchart LR
  subgraph gw["git-workflow"]
    gwf["git-workflow"]
  end
  subgraph wc["workflow-core"]
    cyc["work-cycle"]
    cs["coding-standards"]
    ds["docs-standards"]
    rs["refs-scratch"]
    pp["pre-push"]
  end
  subgraph ks["ko-style"]
    hook{{"ko_style"}}
    aci["anti-claude-ism"]
  end

  gwf --> cs
  gwf --> ds
  gwf --> rs
  gwf -.-> cyc
  cyc --> cs
  cyc --> ds
  cyc --> rs
```

화살표 `A → B`는 A가 B 없이 동작하지 못한다는 뜻이다. 화살표가 없으면 어느
쪽으로도 의존이 없다. 다이어그램에서 hook은 육각형으로, skill은 사각형으로 표현했다.

- 실선은 강결합이다. A 본문이 B의 skill 이름을 직접 참조한다.
- 점선은 약결합이다. A 본문이 B를 참조하지 않는다. B는 자기 트리거 조건이 충족될 때 활성화된다.
  - `git-workflow`의 구현 단계는 파일 수정이 포함되므로 `work-cycle`의 발동 조건에 해당한다.
  - B의 이름을 변경해도 B에 대한 A의 참조는 유효하다.
