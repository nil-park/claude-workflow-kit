# 설치된 스킬의 출처와 업데이트 방법

`claude-workflow-kit` 마켓플레이스의 `project-skills-bootstrap` 플러그인은 스킬을
`.agents/skills/`에 복사해 설치한다. Claude Code 전용 스킬의 설치 위치는 `.claude/skills/`이다.
원본의 변경은 설치본에 자동으로 반영되지 않는다.

- 마켓플레이스: <https://github.com/nil-park/claude-workflow-kit>

변경 사항을 반영하려면 다음 순서로 업데이트를 진행한다. 1단계와 2단계는 Claude Code에만 적용된다.

1. 원격 커밋과 캐시된 커밋을 비교한다.
   - 원격: `git ls-remote https://github.com/nil-park/claude-workflow-kit HEAD`
   - 캐시: `~/.claude/plugins/installed_plugins.json`
     - `project-skills-bootstrap@claude-workflow-kit`의 `gitCommitSha`
2. 두 커밋이 일치하지 않으면 다음 명령을 순서대로 실행한 뒤 Claude Code를 재시작한다.
   - `claude plugin marketplace update claude-workflow-kit`
   - `claude plugin update project-skills-bootstrap@claude-workflow-kit`
3. `/project-skills-bootstrap:bootstrap`을 호출한다.
   - 이 스킬은 설치된 스킬만 업데이트하고, 설치되지 않은 스킬은 안내만 한다.
   - 에이전트가 템플릿과 설치본의 차이를 보고하면 반영할 변경 사항을 상의해서 정하며, 프로젝트별 변경 사항은 유지할 수 있다.
