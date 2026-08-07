"""번역투를 탐지해 턴이 끝나기 전에 알리는 Stop 훅.

설계는 docs/architecture/ko-style.md, 전제는 docs/development/ko-style.md에 있다.
"""

from __future__ import annotations

import json
import os
import re
import sys
import traceback
import unicodedata
from collections.abc import Iterable, Iterator
from pathlib import Path
from typing import NamedTuple, cast

MAX_FILE_BYTES = 1024 * 1024
EDIT_TOOLS = frozenset({"Write", "Edit", "MultiEdit", "NotebookEdit"})
PATH_KEYS = ("file_path", "notebook_path")
ENCODINGS = ("utf-8", "cp949")
DICTIONARY_NAME = "ko-style-dictionary.json"
HANGUL_FIRST = 0xAC00
HANGUL_LAST = 0xD7A3
JONGSEONG_COUNT = 28
JONGSEONG_RIEUL = 8


class Entry(NamedTuple):
    """사전 한 항목."""

    term: str
    judged_as: str
    use: str
    pattern: re.Pattern[str]


class Finding(NamedTuple):
    """탐지 한 건."""

    path: Path
    line: int
    matched: str
    entry: Entry


def _mapping(value: object) -> dict[str, object]:
    return cast("dict[str, object]", value) if isinstance(value, dict) else {}


def _sequence(value: object) -> list[object]:
    return cast("list[object]", value) if isinstance(value, list) else []


def _string(value: object) -> str:
    return value if isinstance(value, str) else ""


def _warn(message: str) -> None:
    """훅은 턴을 막지 않으므로 진단은 stderr로만 남긴다. `claude --debug`에서 보인다."""
    print(f"ko-style: {message}", file=sys.stderr)


def _resolved(path: Path) -> Path:
    try:
        return path.resolve()
    except OSError:
        return path


def read_text(path: Path) -> str | None:
    """텍스트로 읽히면 NFC로 정규화해 돌려주고, 아니면 None. 사전과 판정 대상이 함께 쓴다."""
    try:
        if path.stat().st_size > MAX_FILE_BYTES:
            return None
        data = path.read_bytes()
    except OSError:
        return None
    for encoding in ENCODINGS:
        try:
            return unicodedata.normalize("NFC", data.decode(encoding))
        except UnicodeDecodeError:
            continue
    return None


def project_root() -> Path | None:
    root = os.environ.get("CLAUDE_PROJECT_DIR")
    return _resolved(Path(root)) if root else None


def dictionary_paths() -> list[Path]:
    """읽는 순서대로 돌려준다. 뒤에 읽은 것이 같은 `term`을 이긴다."""
    plugin_root = os.environ.get("CLAUDE_PLUGIN_ROOT")
    hooks_dir = Path(plugin_root) / "hooks" if plugin_root else Path(__file__).parent
    paths = [hooks_dir / DICTIONARY_NAME, Path.home() / ".claude" / DICTIONARY_NAME]
    root = project_root()
    if root is not None:
        paths.append(root / ".claude" / DICTIONARY_NAME)
    return paths


def _read_json(path: Path) -> object:
    text = read_text(path)
    if text is None:
        return None
    try:
        return json.loads(text)
    except ValueError as error:
        _warn(f"{path}가 JSON으로 읽히지 않는다 — {error}")
        return None


def _compile(term: str) -> re.Pattern[str] | None:
    try:
        return re.compile(term)
    except re.error as error:
        _warn(f"{term!r}이 컴파일되지 않는다 — {error}")
        return None


def load_dictionary(paths: Iterable[Path]) -> tuple[list[Entry], list[re.Pattern[str]]]:
    """사전 파일들을 합쳐 탐지 항목과 `ok` 필터로 가른다."""
    merged: dict[str, tuple[str, str]] = {}
    for path in paths:
        data = _read_json(path)
        if data is None:
            continue
        items = _sequence(data)
        if not items and data != []:
            _warn(f"{path}의 최상위가 배열이 아니다. 항목 배열이어야 한다")
        for item in items:
            fields = _mapping(item)
            term = unicodedata.normalize("NFC", _string(fields.get("term")))
            if term:
                merged[term] = (_string(fields.get("as")), _string(fields.get("use")))

    entries: list[Entry] = []
    ok: list[re.Pattern[str]] = []
    for term, (judged_as, use) in merged.items():
        pattern = _compile(term)
        if pattern is None:
            continue
        if judged_as == "ok":
            ok.append(pattern)
        else:
            entries.append(Entry(term, judged_as, use, pattern))
    return entries, ok


def _is_user_input(record: dict[str, object]) -> bool:
    """사용자가 직접 넣은 입력이면 참. 훅이 끼워 넣은 것과 서브에이전트 프롬프트는 뺀다."""
    if record.get("type") != "user" or record.get("isMeta") or record.get("isSidechain"):
        return False
    content = _mapping(record.get("message")).get("content")
    if isinstance(content, str):
        return True
    return any(_mapping(block).get("type") == "text" for block in _sequence(content))


def _edited_paths(record: dict[str, object]) -> Iterator[str]:
    if record.get("type") != "assistant":
        return
    for block in _sequence(_mapping(record.get("message")).get("content")):
        fields = _mapping(block)
        if fields.get("type") != "tool_use" or fields.get("name") not in EDIT_TOOLS:
            continue
        args = _mapping(fields.get("input"))
        for key in PATH_KEYS:
            path = _string(args.get(key))
            if path:
                yield path
                break


def edited_files(transcript: Path) -> list[Path]:
    """마지막 사용자 입력 이후 편집 도구가 건드린 경로를 편집 순서대로 모은다."""
    collected: list[str] = []
    try:
        with transcript.open(encoding="utf-8", errors="replace") as stream:
            for line in stream:
                try:
                    record = _mapping(json.loads(line))
                except ValueError:
                    continue
                if _is_user_input(record):
                    collected.clear()
                else:
                    collected.extend(_edited_paths(record))
    except OSError as error:
        _warn(f"{transcript}를 읽지 못했다 — {error}")
        return []

    unique: dict[Path, None] = {}
    for path in collected:
        unique.setdefault(Path(path), None)
    return list(unique)


def scan(path: Path, entries: Iterable[Entry], ok: Iterable[re.Pattern[str]]) -> list[Finding]:
    """파일 하나를 훑어 나온 순서대로 탐지 결과를 돌려준다."""
    text = read_text(path)
    if text is None:
        return []
    located: list[tuple[int, Finding]] = []
    for entry in entries:
        for match in entry.pattern.finditer(text):
            matched = match.group(0)
            if not matched or any(pattern.search(matched) for pattern in ok):
                continue
            line = text.count("\n", 0, match.start()) + 1
            located.append((match.start(), Finding(path, line, matched, entry)))
    located.sort(key=lambda item: item[0])
    return [finding for _, finding in located]


def _jongseong(word: str) -> int | None:
    """마지막 글자의 받침. 한글 음절이 아니면 None."""
    if not word:
        return None
    code = ord(word[-1])
    if not HANGUL_FIRST <= code <= HANGUL_LAST:
        return None
    return (code - HANGUL_FIRST) % JONGSEONG_COUNT


def _i_ga(word: str) -> str:
    return "가" if _jongseong(word) in (None, 0) else "이"


def _ro(word: str) -> str:
    jongseong = _jongseong(word)
    return "로" if jongseong in (None, 0, JONGSEONG_RIEUL) else "으로"


def _display(path: Path, root: Path | None) -> str:
    """프로젝트 안이면 상대경로, 밖이면 절대경로."""
    if root is not None:
        try:
            return _resolved(path).relative_to(root).as_posix()
        except ValueError:
            pass
    return str(path)


def describe(finding: Finding, root: Path | None) -> str:
    entry = finding.entry
    replacement = f'"{entry.use}"{_ro(entry.use)} ' if entry.use else ""
    return (
        f"{_display(finding.path, root)}:{finding.line}  "
        f'"{finding.matched}"{_i_ga(finding.matched)} '
        f"{entry.judged_as}{_ro(entry.judged_as)} 쓰였다면 {replacement}수정한다."
    )


def report(lines: list[str]) -> None:
    """stdout은 UTF-8로 직접 쓴다. Windows의 기본 stdout 인코딩은 cp949라 그대로 두면 깨진다."""
    payload = {"hookSpecificOutput": {"hookEventName": "Stop", "additionalContext": "\n".join(lines)}}
    sys.stdout.flush()
    sys.stdout.buffer.write(json.dumps(payload, ensure_ascii=False).encode("utf-8") + b"\n")
    sys.stdout.buffer.flush()


def main() -> None:
    stdin = sys.stdin.read().strip()
    payload = _mapping(json.loads(stdin)) if stdin else {}
    if payload.get("stop_hook_active"):
        return
    transcript = _string(payload.get("transcript_path"))
    if not transcript:
        return

    dictionaries = dictionary_paths()
    skipped = {_resolved(path) for path in dictionaries}
    targets = [path for path in edited_files(Path(transcript)) if _resolved(path) not in skipped]
    if not targets:
        return

    entries, ok = load_dictionary(dictionaries)
    if not entries:
        return

    root = project_root()
    lines = [describe(finding, root) for path in targets for finding in scan(path, entries, ok)]
    if lines:
        report(lines)


if __name__ == "__main__":
    try:
        main()
    # 훅의 실패가 턴을 막지 않는다. 무엇이 터졌든 stderr로만 알리고 0으로 끝낸다.
    except Exception:  # noqa: BLE001
        _warn(traceback.format_exc())
    sys.exit(0)
