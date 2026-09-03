import io
import json
import re
from pathlib import Path

import ko_style
import pytest


def write_transcript(tmp_path: Path, records: list[dict[str, object]]) -> Path:
    path = tmp_path / "transcript.jsonl"
    path.write_text("\n".join(json.dumps(record) for record in records) + "\n", encoding="utf-8")
    return path


def user_input(text: str = "고쳐줘") -> dict[str, object]:
    return {"type": "user", "message": {"role": "user", "content": text}}


def tool_use(name: str, args: dict[str, object]) -> dict[str, object]:
    block = {"type": "tool_use", "name": name, "input": args}
    return {"type": "assistant", "message": {"role": "assistant", "content": [block]}}


def edit(path: str) -> dict[str, object]:
    return tool_use("Edit", {"file_path": path, "old_string": "a", "new_string": "b"})


def write_dictionary(path: Path, items: list[dict[str, str]]) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(items, ensure_ascii=False), encoding="utf-8")
    return path


CONSUMER = {"term": "소비자", "as": "컴퓨터 용어에서 consumer의 직역", "use": "컨슈머"}


def test_edited_files_starts_over_at_each_user_input(tmp_path: Path) -> None:
    transcript = write_transcript(
        tmp_path,
        [user_input(), edit("/repo/before.md"), user_input(), edit("/repo/after.md")],
    )
    assert ko_style.edited_files(transcript) == [Path("/repo/after.md")]


def test_edited_files_keeps_going_past_injected_and_sidechain_records(tmp_path: Path) -> None:
    injected: dict[str, object] = {
        "type": "user",
        "isMeta": True,
        "message": {"content": [{"type": "text", "text": "스킬 본문"}]},
    }
    subagent_prompt: dict[str, object] = {"type": "user", "isSidechain": True, "message": {"content": "조사해줘"}}
    transcript = write_transcript(
        tmp_path,
        [user_input(), edit("/repo/a.md"), injected, subagent_prompt, edit("/repo/b.md")],
    )
    assert ko_style.edited_files(transcript) == [Path("/repo/a.md"), Path("/repo/b.md")]


def test_edited_files_drops_what_an_interrupted_turn_wrote(tmp_path: Path) -> None:
    interrupted: dict[str, object] = {
        "type": "user",
        "message": {"content": [{"type": "text", "text": "[Request interrupted by user]"}]},
    }
    transcript = write_transcript(tmp_path, [user_input(), edit("/repo/half-written.md"), interrupted])
    assert ko_style.edited_files(transcript) == []


def test_edited_files_takes_only_edit_tools_and_dedupes(tmp_path: Path) -> None:
    transcript = write_transcript(
        tmp_path,
        [
            user_input(),
            tool_use("Bash", {"command": "sed -i s/a/b/ /repo/shell.md"}),
            edit("/repo/a.md"),
            tool_use("NotebookEdit", {"notebook_path": "/repo/nb.ipynb", "new_source": "x"}),
            edit("/repo/a.md"),
        ],
    )
    assert ko_style.edited_files(transcript) == [Path("/repo/a.md"), Path("/repo/nb.ipynb")]


def test_edited_files_survives_a_broken_line(tmp_path: Path) -> None:
    transcript = tmp_path / "transcript.jsonl"
    transcript.write_text(
        json.dumps(user_input()) + "\n{ this is not json\n" + json.dumps(edit("/repo/a.md")) + "\n",
        encoding="utf-8",
    )
    assert ko_style.edited_files(transcript) == [Path("/repo/a.md")]


def test_edited_files_returns_empty_when_transcript_is_gone(tmp_path: Path) -> None:
    assert ko_style.edited_files(tmp_path / "nowhere.jsonl") == []


def test_later_dictionary_wins_the_same_term(tmp_path: Path) -> None:
    first = write_dictionary(tmp_path / "first.json", [CONSUMER])
    second = write_dictionary(tmp_path / "second.json", [{"term": "소비자", "as": "다시 정한 것", "use": ""}])
    entries, _ = ko_style.load_dictionary([first, second])
    assert [(entry.term, entry.judged_as, entry.use) for entry in entries] == [("소비자", "다시 정한 것", "")]


def test_missing_dictionary_is_skipped(tmp_path: Path) -> None:
    first = write_dictionary(tmp_path / "first.json", [CONSUMER])
    entries, _ = ko_style.load_dictionary([first, tmp_path / "nowhere.json"])
    assert [entry.term for entry in entries] == ["소비자"]


def test_dictionary_saved_as_cp949_is_read(tmp_path: Path) -> None:
    path = tmp_path / "cp949.json"
    path.write_bytes(json.dumps([CONSUMER], ensure_ascii=False).encode("cp949"))
    entries, _ = ko_style.load_dictionary([path])
    assert [entry.judged_as for entry in entries] == ["컴퓨터 용어에서 consumer의 직역"]


def test_broken_dictionary_is_skipped(tmp_path: Path) -> None:
    broken = tmp_path / "broken.json"
    broken.write_text("[{ not json", encoding="utf-8")
    good = write_dictionary(tmp_path / "d.json", [CONSUMER])
    entries, _ = ko_style.load_dictionary([broken, good])
    assert [entry.term for entry in entries] == ["소비자"]


def test_dictionary_that_is_not_an_array_is_skipped_with_a_warning(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    wrong = tmp_path / "wrong.json"
    wrong.write_text(json.dumps({"entries": [CONSUMER]}, ensure_ascii=False), encoding="utf-8")
    entries, _ = ko_style.load_dictionary([wrong])
    assert entries == []
    assert "배열이 아니다" in capsys.readouterr().err


def test_uncompilable_term_is_skipped(tmp_path: Path) -> None:
    path = write_dictionary(tmp_path / "d.json", [{"term": "(", "as": "깨진 정규식", "use": ""}, CONSUMER])
    entries, _ = ko_style.load_dictionary([path])
    assert [entry.term for entry in entries] == ["소비자"]


def test_ok_entry_is_kept_out_of_detection(tmp_path: Path) -> None:
    path = write_dictionary(tmp_path / "d.json", [CONSUMER, {"term": "소비", "as": "ok", "use": ""}])
    entries, ok = ko_style.load_dictionary([path])
    assert [entry.term for entry in entries] == ["소비자"]
    assert [pattern.pattern for pattern in ok] == ["소비"]


def test_ok_entry_filters_a_match_by_the_detected_string(tmp_path: Path) -> None:
    target = tmp_path / "doc.md"
    target.write_text("축이라는 기준\n건축이 무너진다\n", encoding="utf-8")
    path = write_dictionary(
        tmp_path / "d.json",
        [{"term": "(?<![가-힣])축이", "as": "axis의 직역", "use": ""}, {"term": "축", "as": "ok", "use": ""}],
    )
    entries, ok = ko_style.load_dictionary([path])
    assert ko_style.scan(target, entries, ok) == []
    assert [finding.line for finding in ko_style.scan(target, entries, [])] == [1]


def test_scan_reports_every_position_in_order(tmp_path: Path) -> None:
    target = tmp_path / "doc.md"
    target.write_text("첫 줄\n소비자가 온다\n\n소비자 큐\n", encoding="utf-8")
    entries, ok = ko_style.load_dictionary([write_dictionary(tmp_path / "d.json", [CONSUMER])])
    assert [(finding.line, finding.matched) for finding in ko_style.scan(target, entries, ok)] == [
        (2, "소비자"),
        (4, "소비자"),
    ]


def test_scan_reads_cp949(tmp_path: Path) -> None:
    target = tmp_path / "doc.md"
    target.write_bytes("소비자 큐\n".encode("cp949"))
    entries, ok = ko_style.load_dictionary([write_dictionary(tmp_path / "d.json", [CONSUMER])])
    assert [finding.matched for finding in ko_style.scan(target, entries, ok)] == ["소비자"]


def test_scan_skips_a_file_that_is_not_text(tmp_path: Path) -> None:
    target = tmp_path / "blob.bin"
    target.write_bytes(b"\xff\xfe\x00\x80\x81\x82")
    entries, ok = ko_style.load_dictionary([write_dictionary(tmp_path / "d.json", [CONSUMER])])
    assert ko_style.scan(target, entries, ok) == []


def test_scan_skips_a_file_over_the_size_cap(tmp_path: Path) -> None:
    target = tmp_path / "big.md"
    target.write_text("소비자\n" + "가" * ko_style.MAX_FILE_BYTES, encoding="utf-8")
    entries, ok = ko_style.load_dictionary([write_dictionary(tmp_path / "d.json", [CONSUMER])])
    assert ko_style.scan(target, entries, ok) == []


def test_scan_skips_a_file_that_is_gone(tmp_path: Path) -> None:
    entries, ok = ko_style.load_dictionary([write_dictionary(tmp_path / "d.json", [CONSUMER])])
    assert ko_style.scan(tmp_path / "nowhere.md", entries, ok) == []


ROOT = Path("/repo").resolve()
DOC = ROOT / "docs" / "queue.md"


def finding_of(term: str, judged_as: str, use: str, matched: str) -> ko_style.Finding:
    entry = ko_style.Entry(term, judged_as, use, re.compile(term))
    return ko_style.Finding(DOC, 12, matched, entry)


def test_describe_matches_the_designed_wording() -> None:
    finding = finding_of("소비자", "컴퓨터 용어에서 consumer의 직역", "컨슈머", "소비자")
    assert ko_style.describe(finding, ROOT) == (
        'docs/queue.md:12  "소비자"가 컴퓨터 용어에서 consumer의 직역으로 쓰였다면 "컨슈머"로 수정한다.'
    )


def test_describe_drops_the_replacement_when_use_is_empty() -> None:
    finding = finding_of("재수출", "re-export의 직역", "", "재수출")
    assert ko_style.describe(finding, ROOT) == 'docs/queue.md:12  "재수출"이 re-export의 직역으로 쓰였다면 수정한다.'


def test_describe_picks_the_josa_from_the_detected_string() -> None:
    finding = finding_of("되어지", "이중 피동", "", "되어졌")
    assert '"되어졌"이 이중 피동으로 쓰였다면 수정한다.' in ko_style.describe(finding, ROOT)


@pytest.mark.parametrize(
    ("judged_as", "expected"),
    [("파이프라인의 축", "축으로"), ("표준어", "표준어로"), ("직렬", "직렬로"), ("queue", "queue로")],
)
def test_describe_puts_euro_only_after_a_closed_syllable(judged_as: str, expected: str) -> None:
    finding = finding_of("소비자", judged_as, "", "소비자")
    assert f"{expected} 쓰였다면" in ko_style.describe(finding, ROOT)


def test_describe_picks_the_josa_after_the_replacement_too() -> None:
    finding = finding_of("소비자", "consumer의 직역", "출력", "소비자")
    assert ko_style.describe(finding, ROOT).endswith('쓰였다면 "출력"으로 수정한다.')


def test_describe_falls_back_to_the_absolute_path_outside_the_project() -> None:
    finding = finding_of("소비자", "consumer의 직역", "", "소비자")
    assert ko_style.describe(finding, Path("/elsewhere").resolve()).startswith(f"{DOC}:12")


def test_describe_falls_back_to_the_absolute_path_without_a_project_root() -> None:
    finding = finding_of("소비자", "consumer의 직역", "", "소비자")
    assert ko_style.describe(finding, None).startswith(f"{DOC}:12")


SHIPPED = Path(ko_style.__file__).with_name(ko_style.DICTIONARY_NAME)


def scan_text(tmp_path: Path, text: str) -> list[str]:
    target = tmp_path / "doc.md"
    target.write_text(text, encoding="utf-8")
    entries, ok = ko_style.load_dictionary([SHIPPED])
    return [finding.matched for finding in ko_style.scan(target, entries, ok)]


@pytest.mark.parametrize(
    ("text", "matched"),
    [
        ("소비자 큐를 만든다", "소비자"),
        ("배압이 걸린다", "배압"),
        ("모듈을 재수출한다", "재수출"),
        ("축이라는 기준", "축이"),
        ("Cloud Run 축은 잡 성공이 곧 롤아웃이다", "축은"),
        ("두 축 모두 헬스체크가 같다", "축"),
        ("그렇게 되어졌다", "되어졌"),
        ("그렇게 되어진다", "되어진"),
        ("되어지다", "되어지"),
        ("실패 경로를 덮는 테스트", "덮는"),
        ("경계 조건을 덮지 못한다", "덮지"),
        ("형제까지 덮도록 확장자로 건다", "덮도"),
        ("한 항목이 여럿을 덮은 채로 둔다", "덮은"),
        ("이 앱은 Cloud Run에 산다", "산다"),
        ("Cloud Run에 사는 앱", "사는"),
        ("이슈 링크는 MR 본문이 진다", "진다"),
        ("아무것도 끌고 오지 않는다", "끌고 오"),
        ("의존성을 끌고 온다", "끌고 온"),
        ("그 위에 세운 가정", "세운"),
        ("훅의 입출력 계약을 조사했다", "계약"),
        ("설계가 딛고 선 사실", "딛고"),
        ("둘을 가르는 것은 갱신 주체다", "가르는"),
        ("무엇을 볼지 먼저 가른다", "가른다"),
        ("스테이지가 늘면 테이블만 갈린다", "갈린다"),
        ("둘로 갈라 쓴다", "갈라"),
        ("그 제약이 우리를 묶는다", "묶는"),
        ("적재 측 비용은 낮다", "측"),
        ("성능 부분을 개선한다", "부분"),
        ("ingester에는 닿지 않는다", "닿지"),
        ("상한이 듣는지 재는 수단", "재는"),
        ("경로별로 무엇이 좁히나", "이 좁히"),
        ("아무것도 막지 않는 라벨", "막지"),
        ("push를 막아 PR을 거치게 한다", "막아"),
        ("쿠키는 host-only로 굽는다", "굽는"),
        ("검증에 구멍이 생긴다", "구멍"),
        ("폴백을 두면 방어선이 무너진다", "방어선"),
        ("새는 방향으로 조용히 틀린다", "새는"),
        ("게이트가 조용히 열린다", "게이트"),
        ("쿼리가 통째로 떨어져 나간다", "떨어져 나"),
        ("프루닝이 걸렸는지는 이 지표가 답한다", "답한"),
        ("값에 stage를 함께 싣는다", "싣는"),
        ("한 문자열에 나란히 실어 보낸다", "실어"),
        ("조회가 이 둘에 어떻게 기대는지", "기대는"),
        ("이 설정을 타는 스크립트", "타는"),
        ("설정을 타지 않고 직접 접속한다", "타지"),
        ("필요해지면 그때 얹으면 된다", "얹으"),
        ("이 MR이 닫는 간극은 다음과 같다", "닫는"),
    ],
)
def test_shipped_dictionary_catches_what_it_is_registered_for(tmp_path: Path, text: str, matched: str) -> None:
    assert scan_text(tmp_path, text) == [matched]


@pytest.mark.parametrize(
    "text",
    [
        "건축이 무너진다",
        "압축이 풀린다",
        "축소된 결과를 본다",
        "축적된 데이터가 많다",
        "그렇게 되어서 그렇다",
        "되어야 한다",
        "회사는 이렇게 한다",
        "조사는 끝났다",
        "응답이 느려진다",
        "테스트가 만들어진다",
        "눈이 뒤덮인 산",
        "기존 파일을 덮어쓰지 않는다",
        "병렬 세션이 덮어쓰므로 이름을 나눈다",
        "다음 주기의 검출이 이전 기록을 덮어쓴다",
        "이 테스트는 새 값이 옛 값을 덮어쓴 결과를 확인한다",
        "두 세션이 같은 파일을 덮어써서 한쪽 수정이 사라진다",
        "나중에 읽은 사전이 앞선 항목을 덮어쓸 수도 있다",
        "프로젝트 설정이 기본값을 덮어씀으로써 훅의 동작이 달라진다",
        "이 스크립트는 원본 사진에 라벨을 덮어씌운 이미지를 만든다",
        "데이터를 끌어온다",
        "아무것도 딸려 오지 않는다",
        "논거로 내세운 기준",
        "조건을 앞세운 질의",
        "첫발을 내딛고 나아간다",
        "이름이 헷갈린다",
        "무엇을 볼지 가르치는 문서",
        "짐을 묶어 둔다",
        "관측 결과를 본다",
        "서버 측면의 비용",
        "대부분은 그렇다",
        "일부분만 고친다",
        "이 함수는 두 집합이 부분집합 관계인지 판정한다",
        "정규식이 부분일치를 허용한다",
        '"부분"은 없어도 되는 군더더기다',
        "'부분'은 없어도 되는 군더더기다",
        "“부분”은 없어도 되는 군더더기다",
        "‘부분’은 없어도 되는 군더더기다",  # noqa: RUF001 — 곡선 작은따옴표가 제외되는지 보는 케이스라 그 문자여야 한다
        "사전은 `부분`을 군더더기로 등재했다",
        "두 면이 맞닿지 않는다",
        "존재는 확인했다",
        "현재는 그렇지 않다",
        "간격이 좁다",
        "폭이 좁아진다",
        "가로막는 벽이 없다",
        "앞을 가로막아 선다",
        "숨구멍이 트인다",
        "철새는 돌아온다",
        "API 게이트웨이를 둔다",
        "대답하지 않는다",
        "응답하는 시간을 본다",
        "실행이 끝난다",
        "실측으로 확인했다",
        "구현을 베낀 기대값",
        "빠르기를 기대한다",
        "불타는 장작을 본다",
        "애타는 마음이 있다",
        "닫지 않은 파일과 소켓",
        "여닫는 문을 본다",
    ],
)
def test_shipped_dictionary_leaves_these_alone(tmp_path: Path, text: str) -> None:
    assert scan_text(tmp_path, text) == []


@pytest.fixture
def hook_env(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """플러그인 사전과 프로젝트 루트를 tmp에 만든다. 홈 사전은 없다."""
    plugin_root = tmp_path / "plugin"
    write_dictionary(plugin_root / "hooks" / ko_style.DICTIONARY_NAME, [CONSUMER])
    project = tmp_path / "repo"
    (project / "docs").mkdir(parents=True)
    home = tmp_path / "home"
    home.mkdir()
    monkeypatch.setenv("CLAUDE_PLUGIN_ROOT", str(plugin_root))
    monkeypatch.setenv("CLAUDE_PROJECT_DIR", str(project))
    monkeypatch.setattr(ko_style.Path, "home", staticmethod(lambda: home))
    return project


def run_hook(payload: dict[str, object], monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> str:
    monkeypatch.setattr("sys.stdin", io.StringIO(json.dumps(payload)))
    ko_style.main()
    return capsys.readouterr().out


def test_dictionary_paths_are_read_plugin_home_project(hook_env: Path, tmp_path: Path) -> None:
    assert ko_style.dictionary_paths() == [
        tmp_path / "plugin" / "hooks" / "ko-style-dictionary.json",
        tmp_path / "home" / ".claude" / "ko-style-dictionary.json",
        hook_env.resolve() / ".claude" / "ko-style-dictionary.json",
    ]


def test_main_reports_the_findings_as_additional_context(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    target = hook_env / "docs" / "queue.md"
    target.write_text("소비자 큐를 만든다\n", encoding="utf-8")
    transcript = write_transcript(tmp_path, [user_input(), edit(str(target))])

    out = run_hook({"transcript_path": str(transcript), "stop_hook_active": False}, monkeypatch, capsys)

    payload = json.loads(out)
    assert payload["hookSpecificOutput"]["hookEventName"] == "Stop"
    assert payload["hookSpecificOutput"]["additionalContext"] == (
        f'{ko_style.PREAMBLE}\n\ndocs/queue.md:1  "소비자"가 컴퓨터 용어에서 consumer의 직역으로 쓰였다면 "컨슈머"로 수정한다.'
    )


def test_main_says_nothing_on_the_second_stop(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    target = hook_env / "docs" / "queue.md"
    target.write_text("소비자 큐를 만든다\n", encoding="utf-8")
    transcript = write_transcript(tmp_path, [user_input(), edit(str(target))])

    assert run_hook({"transcript_path": str(transcript), "stop_hook_active": True}, monkeypatch, capsys) == ""


def test_main_says_nothing_when_the_turn_edited_nothing(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    transcript = write_transcript(tmp_path, [user_input(), tool_use("Bash", {"command": "ls"})])

    assert run_hook({"transcript_path": str(transcript)}, monkeypatch, capsys) == ""


def test_main_does_not_judge_the_dictionary_itself(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    project_dictionary = write_dictionary(hook_env / ".claude" / ko_style.DICTIONARY_NAME, [CONSUMER])
    transcript = write_transcript(tmp_path, [user_input(), edit(str(project_dictionary))])

    assert run_hook({"transcript_path": str(transcript)}, monkeypatch, capsys) == ""


def test_main_does_not_judge_a_dictionary_off_the_read_paths(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    """배포용 사전의 원본은 리포 안에 있고 읽는 경로 셋 어디에도 없다. 이름으로 걸러야 한다."""
    source = write_dictionary(hook_env / "plugins" / "ko-style" / "hooks" / ko_style.DICTIONARY_NAME, [CONSUMER])
    assert source not in ko_style.dictionary_paths()
    transcript = write_transcript(tmp_path, [user_input(), edit(str(source))])

    assert run_hook({"transcript_path": str(transcript)}, monkeypatch, capsys) == ""


def test_main_does_not_judge_its_own_test_file(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    """이 파일은 등재된 표현을 케이스로 담고 있어 줄마다 탐지된다. 이름으로 제외해야 한다."""
    target = hook_env / "tests" / ko_style.SELF_TEST_NAME
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text('("소비자 큐를 만든다", "소비자"),\n', encoding="utf-8")
    transcript = write_transcript(tmp_path, [user_input(), edit(str(target))])

    assert run_hook({"transcript_path": str(transcript)}, monkeypatch, capsys) == ""


def test_main_still_judges_the_other_files_of_the_turn(
    hook_env: Path, tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    dictionary = write_dictionary(hook_env / ".claude" / ko_style.DICTIONARY_NAME, [CONSUMER])
    target = hook_env / "docs" / "queue.md"
    target.write_text("소비자 큐를 만든다\n", encoding="utf-8")
    transcript = write_transcript(tmp_path, [user_input(), edit(str(dictionary)), edit(str(target))])

    out = run_hook({"transcript_path": str(transcript)}, monkeypatch, capsys)

    assert json.loads(out)["hookSpecificOutput"]["additionalContext"] == (
        f'{ko_style.PREAMBLE}\n\ndocs/queue.md:1  "소비자"가 컴퓨터 용어에서 consumer의 직역으로 쓰였다면 "컨슈머"로 수정한다.'
    )
