---
name: duo-session
description: >-
  main(이 Claude Code 세션)이 지시하고 리뷰하며, 다른 터미널의 agy 세션(peer)이 리포지토리 파일을 모두 편집한다.
  두 세션은 claude-queue와 agy agentapi send-message로 메시지를 주고받는다.
disable-model-invocation: true
---

이 저장소는 `duo-session` 스킬의 배포 원본을 직접 보유하고 있다. 따라서 사본을 두지
않고 원본 템플릿을 그대로 참조한다. 내용을 고칠 때에는 아래 원본 파일만 수정한다.

원본 템플릿에 적힌 `.agents/skills/anti-claudeism/anti_claudeism.py` 경로는 이 저장소에서
`plugins/project-skills-bootstrap/skills/bootstrap-anti-claudeism/anti_claudeism.py`로
읽는다.

@plugins/project-skills-bootstrap/skills/bootstrap-duo-session/duo-session.md
