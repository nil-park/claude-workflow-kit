# 설치된 스킬의 출처와 업데이트 방법

`claude-workflow-kit` 마켓플레이스의 `project-skills-bootstrap` 플러그인은 스킬을
`.agents/skills/`에 복사해 설치한다. Claude Code 전용 스킬의 설치 위치는 `.claude/skills/`이다.
원본의 변경은 설치본에 자동으로 반영되지 않는다.

- 마켓플레이스: <https://github.com/nil-park/claude-workflow-kit>

변경 사항을 반영하려면 `/project-skills-bootstrap:bootstrap`을 부른다. 이 스킬은 설치된
스킬만 업데이트하고, 설치되지 않은 스킬은 안내만 한다. 에이전트가 템플릿과 설치본의 차이를
보고하면 반영할 변경 사항을 상의해서 정하며, 프로젝트별 변경 사항은 유지할 수 있다.
