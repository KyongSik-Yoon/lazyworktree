import pytest

from lazyworktree.app import GitWtStatus

from tests.utils import FakeGit


class FakeApp(GitWtStatus):
    def __init__(self, fake: FakeGit) -> None:
        super().__init__()
        self._fake = fake

    async def run_git(self, args, cwd=None, ok_returncodes=(0,), strip=True) -> str:
        return await self._fake(args, cwd=cwd, ok_returncodes=ok_returncodes, strip=strip)

    async def _apply_delta(self, diff_text: str) -> tuple[str, bool]:
        return diff_text, False


@pytest.mark.asyncio
async def test_build_commit_view_returns_info_and_diff() -> None:
    fake = FakeGit()
    sha = "abc123"
    fmt = "%H%n%an <%ae>%n%ad%n%s%n%b"
    fake.set(
        ["git", "show", "-s", f"--format={fmt}", sha],
        "\n".join(
            [
                sha,
                "Ada Lovelace <ada@example.com>",
                "Mon Jan 1 00:00:00 2024 +0000",
                "Add feature",
                "Body line 1",
                "Body line 2",
            ]
        ),
        cwd="/repo/wt1",
    )
    fake.set(
        ["git", "show", "--patch", "--no-color", "--pretty=format:", sha],
        "diff --git a/file.txt b/file.txt\n+change\n",
        cwd="/repo/wt1",
    )

    app = FakeApp(fake)
    info, diff_text, use_delta = await app._build_commit_view("/repo/wt1", sha)

    assert use_delta is False
    assert info is not None
    assert info["sha"] == sha
    assert "Ada Lovelace" in info["author"]
    assert info["subject"] == "Add feature"
    assert "Body line 1" in info["body"]
    assert "diff --git" in diff_text
