# Codex 지원의 전제

스펙 원문은 [Codex 문서](../setup/codex-docs.md)에서 찾는다.

## 전제

- 한 프로젝트에서 Claude Code와 Codex를 함께 쓴다.
- Codex에서는 fluent-korean 출력 스타일을 전역 지침(`~/.codex/AGENTS.md`)으로 대신한다.
- Codex에서는 모델이 스킬을 스스로 불러도 된다.

## 확인한 사실

- Codex CLI는 이 리포지토리의 `.claude-plugin/` 카탈로그와 매니페스트로 플러그인을 설치하고 스킬을 로드한다.
  - 스킬 이름은 Claude Code와 같은 `project-skills-bootstrap:<스킬>` 형식이다.
- anti-claudeism 훅은 Codex에서 동작하지 않는다.
  - Codex는 `.claude/settings.json`이 아니라 `.codex/hooks.json`에서 훅을 읽는다.
  - Codex는 파일을 `apply_patch`로 고치므로, 훅이 찾는 `Write`/`Edit` 호출이 없다.
  - Codex의 `Stop` 훅은 `additionalContext`를 지원하지 않는다.

## 검증 방법

```bash
codex plugin marketplace add nil-park/claude-workflow-kit
codex plugin add project-skills-bootstrap@claude-workflow-kit
codex plugin list   # 상태가 installed, enabled인지 확인한다
```
