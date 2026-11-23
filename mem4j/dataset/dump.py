#!/usr/bin/env python3
"""Dump merged mem4j recalls to .out/mem4j/merged, and optionally seed Neo4j.

Usage:
    uv run python mem4j/dataset/dump.py --user "<user-scope>" [--seed] [--workers 4] [--credentials <path>]
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

DATASET_DIR = Path(__file__).resolve().parent
CLI_DIR = DATASET_DIR.parent
REPO_ROOT = CLI_DIR.parent
DEFAULT_MERGED_DIR = REPO_ROOT / ".out" / "mem4j" / "merged"
USER_PLACEHOLDER = "{user}"


def resolve_text(value: str, user: str) -> str:
    return value.replace(USER_PLACEHOLDER, user)


def cred_args(cred: Path | None) -> list[str]:
    if cred is None:
        return []
    return ["-credentials", str(cred)]


def build_cli(dest: Path) -> Path:
    exe = dest / ("mem4j.exe" if sys.platform == "win32" else "mem4j")
    subprocess.run(["go", "build", "-o", str(exe), "."], cwd=CLI_DIR, check=True)
    return exe


def create_scope(exe: Path, cred: Path | None, scope: dict[str, Any], user: str) -> tuple[str, int, str]:
    name = resolve_text(str(scope["name"]), user)
    desc = resolve_text(str(scope["description"]), user)
    cmd = [
        str(exe),
        "scope",
        "create",
        *cred_args(cred),
        "-name",
        name,
        "-description",
        desc,
    ]
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
        *cred_args(cred),
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


def dump_merged(
    exe: Path,
    cred: Path | None,
    entry: dict[str, Any],
    user: str,
    merged_dir: Path,
) -> tuple[str, int, str]:
    name = str(entry["name"])
    cmd = [str(exe), "instruction", "recall", *cred_args(cred)]
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


def main() -> None:
    parser = argparse.ArgumentParser(description="Dump merged mem4j recalls and optionally seed Neo4j")
    parser.add_argument("--user", required=True, help="user scope name replacing {user}")
    parser.add_argument("--workers", type=int, default=4, help="concurrent CLI invocations")
    parser.add_argument(
        "--credentials",
        type=Path,
        default=None,
        help="path to Neo4j credentials file (defaults to mem4j CLI default)",
    )
    parser.add_argument(
        "--merged-dir",
        type=Path,
        default=DEFAULT_MERGED_DIR,
        help="output directory for merged recall files (default: .out/mem4j/merged)",
    )
    parser.add_argument(
        "--seed",
        action="store_true",
        help="create scopes and instructions in Neo4j before dumping merged recalls",
    )
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
            for scope in scopes:
                name, rc, out = create_scope(exe, args.credentials, scope, user)
                if rc != 0:
                    print(f"FAIL scope {name} (exit {rc}): {out}", file=sys.stderr)
                    sys.exit(1)

            unit_failures = run_parallel(
                args.workers,
                units,
                lambda u: seed_unit(exe, args.credentials, u, user),
            )
            if unit_failures:
                for key, rc, out in unit_failures:
                    print(f"FAIL instruction {key} (exit {rc}): {out}", file=sys.stderr)
                sys.exit(1)
            print(f"Replayed {len(scopes)} scopes and {len(units)} instruction units via mem4j CLI.")

        args.merged_dir.mkdir(parents=True, exist_ok=True)
        merge_failures = run_parallel(
            args.workers,
            coverage,
            lambda c: dump_merged(exe, args.credentials, c, user, args.merged_dir),
        )
        if merge_failures:
            for name, rc, out in merge_failures:
                print(f"FAIL recall {name} (exit {rc}): {out}", file=sys.stderr)
            sys.exit(1)
        print(f"Wrote {len(coverage)} merged recall files to {args.merged_dir}.")


if __name__ == "__main__":
    main()
