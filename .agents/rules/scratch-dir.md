---
trigger: always_on
description: 하네스 스크래치패드와 `.refs` 디렉터리의 용도 구분
---

이 파일은 `scratch-dir` 스킬의 원본 템플릿을 상시 규칙으로 불러온다.
내용을 고칠 때에는 원본 템플릿만 수정한다.

원본 템플릿에 남아 있는 `{{scratch_dir}}`는 이 리포지토리에서 모두 `.refs`로 읽는다.
프로젝트 로컬 스크래치 디렉터리는 `<repo-root>/.refs/`이며, 이 경로는 이미 `.gitignore`에 등록되어 있다.

@[scratch-dir](../../plugins/project-skills-bootstrap/skills/bootstrap-scratch-dir/scratch-dir.md)
