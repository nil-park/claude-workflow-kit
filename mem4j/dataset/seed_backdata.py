#!/usr/bin/env python3
"""Seed mem4j back-data (entities, hub links, statements, events) into Neo4j."""

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
USER_PLACEHOLDER = "{user}"


def resolve_text(value: str, user: str) -> str:
    return value.replace(USER_PLACEHOLDER, user)


def flag_if_set(flag: str, path: Path | None) -> list[str]:
    return [flag, str(path)] if path is not None else []


def run_cli(cmd: list[str], key: str, ignore_exists: bool = False) -> tuple[str, int, str]:
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", check=False)
    out = (proc.stdout + proc.stderr).strip()
    if ignore_exists and proc.returncode != 0 and "already exists" in out:
        return key, 0, out
    return key, proc.returncode, out


def seed_entity(
    exe: Path, cred: Path | None, gcp_cred: Path | None, item: dict[str, Any], user: str
) -> tuple[str, int, str]:
    name = resolve_text(str(item["name"]), user)
    desc = resolve_text(str(item["description"]), user)
    cmd = [
        str(exe),
        "entity",
        "create",
        *flag_if_set("-credentials", cred),
        *flag_if_set("-gcp-credentials", gcp_cred),
        "-name",
        name,
        "-description",
        desc,
    ]
    return run_cli(cmd, name, ignore_exists=True)


def seed_hub_link(
    exe: Path, cred: Path | None, resource: str, item: dict[str, Any], user: str
) -> tuple[str, int, str]:
    src = resolve_text(str(item["from"]), user)
    dst = resolve_text(str(item["to"]), user)
    rel_type = str(item["type"])
    cmd = [
        str(exe),
        resource,
        "link",
        *flag_if_set("-credentials", cred),
        "-from",
        src,
        "-to",
        dst,
        "-type",
        rel_type,
    ]
    return run_cli(cmd, f"{resource}:{src}-{rel_type}->{dst}")


def seed_fact(
    exe: Path,
    cred: Path | None,
    gcp_cred: Path | None,
    resource: str,
    item: dict[str, Any],
    user: str,
) -> tuple[str, int, str]:
    key = str(item["key"])
    cmd = [
        str(exe),
        resource,
        "create",
        *flag_if_set("-credentials", cred),
        *flag_if_set("-gcp-credentials", gcp_cred),
        "-key",
        key,
        "-title",
        str(item["title"]),
        "-description",
        resolve_text(str(item["description"]), user),
        "-body",
        resolve_text(str(item["body"]), user),
    ]
    if resource == "event":
        cmd.extend(["-occurred-at", str(item["occurredAt"])])
    for raw_ent in item["entities"]:
        cmd.extend(["-entity", resolve_text(str(raw_ent), user)])
    return run_cli(cmd, key, ignore_exists=True)


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


def seed_all_backdata(exe: Path, cred: Path | None, gcp_cred: Path | None, user: str, workers: int) -> None:
    entities: list[dict[str, Any]] = json.loads((DATASET_DIR / "entities.json").read_text(encoding="utf-8"))
    hub_links: dict[str, list[dict[str, Any]]] = json.loads(
        (DATASET_DIR / "hub_links.json").read_text(encoding="utf-8")
    )
    statements: list[dict[str, Any]] = json.loads((DATASET_DIR / "statements.json").read_text(encoding="utf-8"))
    events: list[dict[str, Any]] = json.loads((DATASET_DIR / "events.json").read_text(encoding="utf-8"))

    fail_if(
        "entity",
        run_parallel(workers, entities, lambda e: seed_entity(exe, cred, gcp_cred, e, user)),
    )
    fail_if(
        "scope link",
        run_parallel(workers, hub_links["scopeLinks"], lambda x: seed_hub_link(exe, cred, "scope", x, user)),
    )
    fail_if(
        "entity link",
        run_parallel(workers, hub_links["entityLinks"], lambda x: seed_hub_link(exe, cred, "entity", x, user)),
    )
    fail_if(
        "statement",
        run_parallel(workers, statements, lambda s: seed_fact(exe, cred, gcp_cred, "statement", s, user)),
    )
    fail_if(
        "event",
        run_parallel(workers, events, lambda ev: seed_fact(exe, cred, gcp_cred, "event", ev, user)),
    )
    print(
        f"Seeded {len(entities)} entities, {len(hub_links['scopeLinks'])} scope links, "
        f"{len(hub_links['entityLinks'])} entity links, {len(statements)} statements, and {len(events)} events."
    )


def main() -> None:
    parser = argparse.ArgumentParser(description="Seed mem4j back-data into Neo4j")
    parser.add_argument("--user", required=True, help="user name replacing {user}")
    parser.add_argument("--workers", type=int, default=4, help="concurrent CLI invocations")
    parser.add_argument("--credentials", type=Path, default=None, help="Neo4j credentials file")
    parser.add_argument("--gcp-credentials", type=Path, default=None, help="GCP service account JSON")
    args = parser.parse_args()

    user = args.user.strip()
    if not user:
        parser.error("--user must be non-empty")
    if args.workers < 1:
        parser.error("--workers must be at least 1")

    with tempfile.TemporaryDirectory() as tmp:
        exe = Path(tmp) / ("mem4j.exe" if sys.platform == "win32" else "mem4j")
        subprocess.run(["go", "build", "-o", str(exe), "."], cwd=CLI_DIR, check=True)
        seed_all_backdata(exe, args.credentials, args.gcp_credentials, user, args.workers)


if __name__ == "__main__":
    main()
