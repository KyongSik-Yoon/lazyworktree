#!/usr/bin/env python3
from __future__ import annotations

import shutil
import subprocess
import tempfile
from pathlib import Path


def _ignore(path: str, names: list[str]) -> set[str]:
    ignored = {
        ".git",
        ".venv",
        "__pycache__",
        ".pytest_cache",
    }
    return {name for name in names if name in ignored}


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    temp_root = Path(tempfile.mkdtemp(prefix="lazyworktree-precommit-"))
    try:
        shutil.copytree(root, temp_root, dirs_exist_ok=True, ignore=_ignore)
        subprocess.run(
            ["uv", "run", "--extra", "dev", "pytest"],
            cwd=temp_root,
            check=True,
        )
    finally:
        shutil.rmtree(temp_root, ignore_errors=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
