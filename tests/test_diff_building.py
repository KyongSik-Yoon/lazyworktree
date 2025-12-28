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
async def test_build_diff_text_combines_sections_and_limits_untracked() -> None:
    fake = FakeGit()
    fake.set(["git", "diff", "--cached", "--patch", "--no-color"], "staged diff\n", cwd="/repo/wt1")
    fake.set(["git", "diff", "--patch", "--no-color"], "unstaged diff\n", cwd="/repo/wt1")
    untracked_files = [f"file{i}.txt" for i in range(12)]
    fake.set(
        ["git", "ls-files", "--others", "--exclude-standard"],
        "\n".join(untracked_files),
        cwd="/repo/wt1",
    )
    for name in untracked_files:
        fake.set(
            ["git", "diff", "--no-index", "--no-color", "--", "/dev/null", name],
            f"diff --git /dev/null b/{name}\n+{name}\n",
            cwd="/repo/wt1",
        )

    app = FakeApp(fake)
    diff_text, use_delta = await app._build_diff_text("/repo/wt1")

    assert use_delta is False
    assert "# Staged" in diff_text
    assert "# Unstaged" in diff_text
    assert "# Untracked" in diff_text
    assert "Showing first 10 untracked files (total: 12)" in diff_text
    assert diff_text.count("diff --git /dev/null") == 10
