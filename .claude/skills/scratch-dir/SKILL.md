---
name: scratch-dir
description: >-
  커밋하지 않는 작업 파일을 어디에 둘지 결정할 때 부른다. 하네스 스크래치패드와
  `.refs` 디렉터리의 용도를 구분하고, 각 디렉터리에서 파일을 이름 짓는 규약을 다룬다.
  여러 줄 텍스트를 CLI 플래그나 stdin으로 넘겨야 할 때 부른다.
---

이 저장소는 `scratch-dir` 스킬의 배포 원본을 직접 보유하고 있다. 따라서 사본을 두지
않고 원본 템플릿을 그대로 참조한다. 내용을 고칠 때에는 아래 원본 파일만 수정한다.

원본 템플릿에 남아 있는 `{{scratch_dir}}` 자리는 이 저장소에서 모두 `.refs`로 읽는다.
즉 프로젝트 로컬 스크래치 디렉터리는 `<repo-root>/.refs/`이며, 이 경로는 이미
`.gitignore`에 등록되어 있다.

@plugins/project-skills-bootstrap/skills/bootstrap-scratch-dir/scratch-dir.md
