---
name: bootstrap-anti-claudeism
description: >-
  anti-claudeism 스킬과 그 Stop 훅을 현재 프로젝트에 설치하거나 업데이트하고 싶을 때 부른다.
---

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래
파일들을 읽어 프로젝트에 설치하거나 업데이트한다. 설치되는 것은 문장 단위를 판정하는 스킬과,
낱말 단위를 사전으로 탐지하는 Stop 훅 둘이다.

| 템플릿                         | 설치 위치                                                    |
| ------------------------------ | ------------------------------------------------------------ |
| `anti-claudeism.md`            | `.agents/skills/anti-claudeism/SKILL.md`                     |
| `references/word-level.md`     | `.agents/skills/anti-claudeism/references/word-level.md`     |
| `references/sentence-level.md` | `.agents/skills/anti-claudeism/references/sentence-level.md` |
| `anti_claudeism.py`            | `.agents/skills/anti-claudeism/anti_claudeism.py`            |
| `claudeism-dictionary.json`    | `.agents/skills/anti-claudeism/claudeism-dictionary.json`    |

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
2. 마크다운 템플릿 셋(`anti-claudeism.md`, `references/word-level.md`,
   `references/sentence-level.md`)과 설치된 대응 파일을 각각 읽어 차이를 사용자에게 보고한다.
3. 사용자와 상의해 반영할 변경과 유지할 내용을 정한 뒤 파일을 수정한다.
4. `.claude/settings.json`에 위 Stop 훅 명령이 등록되어 있는지 확인하고, 없으면 설치 절차의
   3번과 4번을 실행한다.
