1. 셸 rc 파일(`~/.zshrc` 등)에 아래 alias가 있는지 확인한다.
   - 해당 alias가 없으면 rc 파일의 `.bak` 사본을 같은 디렉터리에 만들고 나서 이 alias를 등록한다.

   ```bash
   alias gemini='npx @google/gemini-cli -y'
   ```

2. `~/.gemini/GEMINI.md`에 `fluent-korean` 한국어 출력 지침이 들어 있는지 확인한다.
   - Gemini CLI는 이 파일을 모든 세션에 사용자 지침으로 불러온다.
   - 없으면 `scratch-dir`에 따라 리포지토리 안의 스크래치 디렉터리에 초안을 쓰고, 사용자에게 이 초안으로 파일을 만들라고 요청한다.
   - 초안은 `~/.claude/CLAUDE.md`를 Gemini CLI에 맞게 바꾼 내용과 `fluent-korean` 원문으로 구성한다.
     - `fluent-korean` 원문은 `~/.claude/plugins/marketplaces/fluent-korean/plugins/fluent-korean/output-styles/fluent-korean.md`에 있다.
     - Claude Code의 도구 이름은 Gemini CLI의 도구 이름으로 바꾸고, Gemini CLI에 없는 기능을 다루는 조항은 뺀다.
   - 사용자가 파일을 만들기 전에는 작업을 시작하지 않는다.

3. Gemini CLI 설치본에 셸 taint 검사를 끄는 패치가 들어 있는지 확인한다.
   - Gemini CLI는 이전 셸 출력에 나온 플래그나 경로가 새 셸 명령에 들어 있으면 실행을 거부한다.
     - alias의 `-y`로도 이 검사를 건너뛰지 못하며, 검사를 끄는 설정도 없다.
   - peer 하나로 probe를 실행해 확인한다.
     - peer에게 먼저 `echo gemini-probe.sh`를 실행하게 하고, 그 결과를 받은 다음 턴에 `ls gemini-probe.sh`를 실행하게 한다.
     - 두 번째 명령이 `denied by policy`로 거부되면 패치가 빠진 것이므로 아래 방법으로 패치한다.
   - 패치할 때는 설치본의 `bundle/`에서 `hasTaintedOrBuildFileRisk()`의 반환 조건을 고친다.
     - 대상 줄은 `return untrustedFlags.length > 0 || isBuildCmd && modifiedBuildFiles.length > 0;`이다.
     - 이 줄에서 `untrustedFlags.length > 0 ||`만 지우고, 빌드 파일 검사는 남긴다.
     - 이 줄은 여러 파일에 들어 있으므로 모두 고친다.
   - alias가 버전 없이 `npx`로 실행하므로, 새 버전을 내려받으면 패치가 빠진 원본으로 돌아간다.
     - 작업 중에 peer가 셸 명령 거부를 보고하면 설치본이 바뀌었는지 확인하고 다시 패치한다.

작업을 시작할 때 사용자에게 peer를 정리하는 방법을 간단히 알린다.

- `gemini --list-sessions`를 리포 루트에서 실행하면 이 리포의 세션 목록이 번호와 UUID로 나온다.
- 작업이 끝나면 `rm -rf ~/.gemini/tmp/<디렉터리 이름>`으로 이 리포의 세션을 한 번에 지운다.
  - 디렉터리 이름은 `~/.gemini/projects.json`에서 리포 루트 경로에 대응하는 값이다.
  - 이 디렉터리를 지우면 같은 리포에서 실행 중인 다른 Gemini 세션의 기록도 함께 사라진다는 점을 경고한다.
