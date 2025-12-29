import hashlib
import json
import os
from enum import Enum, auto
from pathlib import Path
from typing import Dict

# Default path for the trusted database
TRUST_DB_PATH = (
    Path(os.environ.get("XDG_DATA_HOME", "~/.local/share")).expanduser()
    / "lazyworktree"
    / "trusted.json"
)


class TrustStatus(Enum):
    TRUSTED = auto()
    UNTRUSTED = auto()  # Content changed or new file
    NOT_FOUND = (
        auto()
    )  # File does not exist (technically safe effectively, but handled distinctly)


class TrustManager:
    def __init__(self, db_path: Path = TRUST_DB_PATH):
        self.db_path = db_path
        self._trusted_hashes: Dict[str, str] = {}  # Map absolute path -> sha256 hash
        self._load()

    def _load(self):
        if not self.db_path.exists():
            return
        try:
            with self.db_path.open("r", encoding="utf-8") as f:
                self._trusted_hashes = json.load(f)
        except (OSError, json.JSONDecodeError):
            # If corrupt, we start fresh for safety (fail closed-ish)
            self._trusted_hashes = {}

    def _save(self):
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        try:
            with self.db_path.open("w", encoding="utf-8") as f:
                json.dump(self._trusted_hashes, f, indent=2)
        except OSError:
            pass  # TODO: Log warning?

    def _calculate_hash(self, file_path: Path) -> str:
        """Calculates SHA256 of the file content."""
        sha256 = hashlib.sha256()
        try:
            with file_path.open("rb") as f:
                while True:
                    data = f.read(65536)
                    if not data:
                        break
                    sha256.update(data)
        except OSError:
            return ""
        return sha256.hexdigest()

    def check_trust(self, file_path: Path) -> TrustStatus:
        """Checks if the file at file_path is trusted."""
        resolved_path = str(file_path.resolve())

        if not file_path.exists():
            return TrustStatus.NOT_FOUND

        current_hash = self._calculate_hash(file_path)
        stored_hash = self._trusted_hashes.get(resolved_path)

        if stored_hash == current_hash:
            return TrustStatus.TRUSTED

        return TrustStatus.UNTRUSTED

    def trust_file(self, file_path: Path):
        """Marks the current content of file_path as trusted."""
        if not file_path.exists():
            return

        resolved_path = str(file_path.resolve())
        current_hash = self._calculate_hash(file_path)
        self._trusted_hashes[resolved_path] = current_hash
        self._save()
