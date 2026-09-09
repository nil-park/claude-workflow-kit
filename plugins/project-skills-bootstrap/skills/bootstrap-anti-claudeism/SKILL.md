---
name: bootstrap-anti-claudeism
description: >-
  anti-claudeism 스킬과 그 Stop 훅을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
파일들을 읽어 프로젝트에 설치하거나 업데이트한다. 설치되는 것은 문장 단위의 결함을 다루는
스킬과, 낱말 단위를 사전으로 탐지하는 Stop 훅 둘이다.

| 템플릿                      | 설치 위치                                                 |
| --------------------------- | --------------------------------------------------------- |
| `anti-claudeism.md`         | `.agents/skills/anti-claudeism/SKILL.md`                  |
| `anti_claudeism.py`         | `.agents/skills/anti-claudeism/anti_claudeism.py`         |
| `claudeism-dictionary.json` | `.agents/skills/anti-claudeism/claudeism-dictionary.json` |

## 설치

`.agents/skills/anti-claudeism/SKILL.md`가 없을 때 실행한다.

1. 위 표의 템플릿을 순서대로 대응하는 설치 위치에 복사한다.
2. `.claude/skills/anti-claudeism`에 `.agents/skills/anti-claudeism`을 가리키는
   심볼릭 링크를 만든다.
3. `.claude/settings.json`을 읽는다. 파일이 없으면 빈 객체 `{}`로 시작한다.
4. 아래 항목을 추가해 파일을 쓴다. `hooks.Stop`이 이미 있으면 배열에 덧붙이고, 같은 명령이
   등록되어 있으면 그대로 둔다.

   ```json
   {
     "hooks": {
       "Stop": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "python3 \"${CLAUDE_PROJECT_DIR}/.agents/skills/anti-claudeism/anti_claudeism.py\"",
               "timeout": 10,
               "statusMessage": "한국어 문체 검수 중"
             }
           ]
         }
       ]
     }
   }
   ```

5. 설치 결과를 보고하면서 아래 네 가지를 함께 안내한다.
   - 훅을 실행하려면 `python3`라는 이름으로 Python 3.11 이상을 실행할 수 있어야 한다.
     python.org 인스톨러로 설치한 Windows에는 `python3.exe`가 없다.
   - 훅 등록은 새 세션부터 잡히기도 하므로 `/hooks`로 등록 여부를 확인한다.
     `0 hooks configured`로 나오면 `claude --resume`으로 세션을 다시 켠다.
   - 훅은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit`으로 고친 파일만 검사한다. auto mode는
     Bash로 파일을 고치라고 지시하므로 그 모드에서는 훅이 검사 대상을 받지 못한다.
   - 프로젝트에서만 쓸 사전 항목은 `.claude/claudeism-dictionary.json`에 적는다. 설치본
     사전은 다음 업데이트에 통째로 교체된다.

## 업데이트

`.agents/skills/anti-claudeism/SKILL.md`가 이미 있을 때 실행한다.

1. `anti_claudeism.py`와 `claudeism-dictionary.json`은 템플릿으로 덮어쓴다. 이 둘은 설치본을
   고쳐 쓰는 파일이 아니다.
2. 마크다운 템플릿 `anti-claudeism.md`와 설치된 대응 파일을 읽어 차이를 사용자에게 보고한다.
3. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.
4. `.claude/settings.json`에 위 Stop 훅 명령이 등록되어 있는지 확인하고, 없으면 설치 절차의
   3번과 4번을 실행한다.

## UPDATE.md 기록

- `.agents/skills/UPDATE.md`가 없으면 아래 내용으로 만든다.
- 파일은 있으나 목록에 `anti-claudeism` 항목이 없으면 추가한다.

```markdown
# 스킬 업데이트

## 업스트림

- 리포지토리: https://github.com/nil-park/claude-workflow-kit
- 경로: `plugins/project-skills-bootstrap/skills/`
- 업스트림의 `bootstrap-<스킬명>/SKILL.md`는 설치·업데이트 로더이고, 실제 템플릿은 같은 디렉터리에 있는 아래 표의 파일이다.

## 대상 스킬

| 로컬 경로                                  | 업스트림 템플릿                                    |
| ------------------------------------------ | -------------------------------------------------- |
| `anti-claudeism/SKILL.md`                  | `bootstrap-anti-claudeism/anti-claudeism.md`       |
| `anti-claudeism/anti_claudeism.py`         | `bootstrap-anti-claudeism/anti_claudeism.py`       |
| `anti-claudeism/claudeism-dictionary.json` | `bootstrap-anti-claudeism/claudeism-dictionary.json` |

- 파이썬 스크립트와 사전 파일은 업데이트할 때 템플릿으로 덮어쓴다. 프로젝트에서만 쓸 사전 항목은 `.claude/claudeism-dictionary.json`에 적는다.
- `anti-claudeism`은 `.claude/settings.json`의 Stop 훅도 함께 설치하므로, 업데이트할 때 등록 여부를 확인한다.

## 업데이트 절차

- `project-skills-bootstrap:bootstrap-<스킬명>` 스킬을 실행하면, 이미 설치된 스킬은 업스트림 템플릿과 비교하는 업데이트 흐름으로 진행된다.
```
