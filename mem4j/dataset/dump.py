#!/usr/bin/env python3
"""Dump merged mem4j recalls to .out/mem4j/merged, and optionally seed Neo4j.

Usage:
    uv run python mem4j/dataset/dump.py --user "<user-scope>" [--seed] [--workers 4]
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import subprocess
import sys
import tempfile
from collections.abc import Callable
from pathlib import Path
from typing import Any

from seed_backdata import seed_all_backdata

DATASET_DIR = Path(__file__).resolve().parent
CLI_DIR = DATASET_DIR.parent
REPO_ROOT = CLI_DIR.parent
DEFAULT_MERGED_DIR = REPO_ROOT / ".out" / "mem4j" / "merged"
USER_PLACEHOLDER = "{user}"


def resolve_text(value: str, user: str) -> str:
    return value.replace(USER_PLACEHOLDER, user)


def flag_if_set(flag: str, path: Path | None) -> list[str]:
    return [flag, str(path)] if path is not None else []


def build_cli(dest: Path) -> Path:
    exe = dest / ("mem4j.exe" if sys.platform == "win32" else "mem4j")
    subprocess.run(["go", "build", "-o", str(exe), "."], cwd=CLI_DIR, check=True)
    return exe


def create_scope(exe: Path, cred: Path | None, scope: dict[str, Any], user: str) -> tuple[str, int, str]:
    name = resolve_text(str(scope["name"]), user)
    desc = resolve_text(str(scope["description"]), user)
    cmd = [str(exe), "scope", "create", *flag_if_set("-credentials", cred), "-name", name, "-description", desc]
    if scope.get("disableModelInvocation"):
        cmd.append("-disable-model-invocation")
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", check=False)
    out = (proc.stdout + proc.stderr).strip()
    if proc.returncode != 0 and "already exists" in out:
        return name, 0, out
    return name, proc.returncode, out


def seed_unit(exe: Path, cred: Path | None, unit: dict[str, Any], user: str) -> tuple[str, int, str]:
    body_path = DATASET_DIR / str(unit["file"])
    cmd = [
        str(exe),
        "instruction",
        "create",
        *flag_if_set("-credentials", cred),
        "-key",
        str(unit["key"]),
        "-title",
        str(unit["title"]),
        "-file",
        str(body_path),
    ]
    for raw_scope in unit["scopes"]:
        cmd.extend(["-scope", resolve_text(str(raw_scope), user)])
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", check=False)
    return str(unit["key"]), proc.returncode, (proc.stdout + proc.stderr).strip()


def seed_reminder(
    exe: Path, cred: Path | None, gcp_cred: Path | None, item: dict[str, Any], user: str
) -> tuple[str, int, str]:
    cmd = [
        str(exe),
        "reminder",
        "create",
        *flag_if_set("-credentials", cred),
        *flag_if_set("-gcp-credentials", gcp_cred),
        "-key",
        str(item["key"]),
        "-title",
        str(item["title"]),
        "-description",
        str(item["description"]),
        "-body",
        str(item["body"]),
    ]
    for raw_scope in item["scopes"]:
        cmd.extend(["-scope", resolve_text(str(raw_scope), user)])
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", check=False)
    return str(item["key"]), proc.returncode, (proc.stdout + proc.stderr).strip()


def dump_merged(
    exe: Path, cred: Path | None, entry: dict[str, Any], user: str, merged_dir: Path
) -> tuple[str, int, str]:
    name = str(entry["name"])
    cmd = [str(exe), "instruction", "recall", *flag_if_set("-credentials", cred)]
    for raw_scope in entry["scopes"]:
        cmd.extend(["-scope", resolve_text(str(raw_scope), user)])
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", check=False)
    if proc.returncode != 0:
        return name, proc.returncode, (proc.stdout + proc.stderr).strip()
    if not proc.stdout.strip():
        return name, 1, proc.stderr.strip() or "empty recall output"
    (merged_dir / f"{name}.txt").write_text(proc.stdout, encoding="utf-8")
    return name, 0, proc.stderr.strip()


def run_parallel(
    workers: int,
    items: list[dict[str, Any]],
    fn: Callable[[dict[str, Any]], tuple[str, int, str]],
) -> list[tuple[str, int, str]]:
    failures: list[tuple[str, int, str]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        futs = [pool.submit(fn, item) for item in items]
        for fut in concurrent.futures.as_completed(futs):
            key, rc, out = fut.result()
            if rc != 0:
                failures.append((key, rc, out))
    return sorted(failures)


def fail_if(kind: str, failures: list[tuple[str, int, str]]) -> None:
    if not failures:
        return
    for key, rc, out in failures:
        print(f"FAIL {kind} {key} (exit {rc}): {out}", file=sys.stderr)
    sys.exit(1)


def main() -> None:
    parser = argparse.ArgumentParser(description="Dump merged mem4j recalls and optionally seed Neo4j")
    parser.add_argument("--user", required=True, help="user scope name replacing {user}")
    parser.add_argument("--workers", type=int, default=4, help="concurrent CLI invocations")
    parser.add_argument("--credentials", type=Path, default=None, help="Neo4j credentials file")
    parser.add_argument("--gcp-credentials", type=Path, default=None, help="GCP service account JSON")
    parser.add_argument("--merged-dir", type=Path, default=DEFAULT_MERGED_DIR, help="output dir for merged recalls")
    parser.add_argument("--seed", action="store_true", help="seed scopes, instructions, and reminders before dumping")
    args = parser.parse_args()

    user = args.user.strip()
    if not user:
        parser.error("--user must be non-empty")
    if args.workers < 1:
        parser.error("--workers must be at least 1")

    coverage: list[dict[str, Any]] = json.loads((DATASET_DIR / "coverage.json").read_text(encoding="utf-8"))

    with tempfile.TemporaryDirectory() as tmp:
        exe = build_cli(Path(tmp))

        if args.seed:
            scopes: list[dict[str, Any]] = json.loads((DATASET_DIR / "scopes.json").read_text(encoding="utf-8"))
            units: list[dict[str, Any]] = json.loads((DATASET_DIR / "units.json").read_text(encoding="utf-8"))
            reminders: list[dict[str, Any]] = json.loads((DATASET_DIR / "reminders.json").read_text(encoding="utf-8"))
            for scope in scopes:
                name, rc, out = create_scope(exe, args.credentials, scope, user)
                if rc != 0:
                    print(f"FAIL scope {name} (exit {rc}): {out}", file=sys.stderr)
                    sys.exit(1)
            fail_if(
                "instruction", run_parallel(args.workers, units, lambda u: seed_unit(exe, args.credentials, u, user))
            )
            fail_if(
                "reminder",
                run_parallel(
                    args.workers,
                    reminders,
                    lambda r: seed_reminder(exe, args.credentials, args.gcp_credentials, r, user),
                ),
            )
            print(f"Seeded {len(scopes)} scopes, {len(units)} instructions, and {len(reminders)} reminders.")
            seed_all_backdata(exe, args.credentials, args.gcp_credentials, user, args.workers)

        args.merged_dir.mkdir(parents=True, exist_ok=True)
        fail_if(
            "recall",
            run_parallel(
                args.workers, coverage, lambda c: dump_merged(exe, args.credentials, c, user, args.merged_dir)
            ),
        )
        print(f"Wrote {len(coverage)} merged recall files to {args.merged_dir}.")


if __name__ == "__main__":
    main()
