import pytest

from lazyworktree.app import GitWtStatus

from tests.utils import FakeGit


class FakeApp(GitWtStatus):
    def __init__(self, fake: FakeGit) -> None:
        super().__init__()
        self._fake = fake

    async def run_git(self, args, cwd=None, ok_returncodes=(0,), strip=True) -> str:
        return await self._fake(args, cwd=cwd, ok_returncodes=ok_returncodes, strip=strip)


@pytest.mark.asyncio
async def test_get_worktrees_parses_porcelain_and_status() -> None:
    fake = FakeGit()
    fake.set(
        ["git", "worktree", "list", "--porcelain"],
        "\n".join(
            [
                "worktree /repo",
                "HEAD 111",
                "branch refs/heads/main",
                "worktree /repo/feature1",
                "HEAD 222",
                "branch refs/heads/feature1",
                "worktree /repo/feature2",
                "HEAD 333",
                "branch refs/heads/feature2",
            ]
        ),
    )
    fake.set(
        [
            "git",
            "for-each-ref",
            "--format=%(refname:short)|%(committerdate:relative)|%(committerdate:unix)",
            "refs/heads",
        ],
        "\n".join(
            [
                "main|2 days ago|100",
                "feature1|1 day ago|200",
                "feature2|3 days ago|50",
            ]
        ),
    )
    fake.set(
        ["git", "status", "--porcelain=v2", "--branch"],
        "# branch.ab +1 -2\n",
        cwd="/repo",
    )
    fake.set(
        ["git", "status", "--porcelain=v2", "--branch"],
        "# branch.ab +0 -0\n? new.txt\n1 .M N... 100644 100644 100644 1 1 file.txt\n",
        cwd="/repo/feature1",
    )
    fake.set(
        ["git", "status", "--porcelain=v2", "--branch"],
        "# branch.ab +2 -1\n1 A. N... 100644 100644 100644 1 1 staged.txt\n",
        cwd="/repo/feature2",
    )

    app = FakeApp(fake)
    worktrees = await app.get_worktrees()

    assert [wt.branch for wt in worktrees] == ["main", "feature1", "feature2"]
    assert worktrees[0].is_main is True
    assert worktrees[1].is_main is False
    assert worktrees[0].ahead == 1
    assert worktrees[0].behind == 2
    assert worktrees[1].dirty is True
    assert worktrees[1].untracked == 1
    assert worktrees[1].modified == 1
    assert worktrees[1].staged == 0
    assert worktrees[2].staged == 1
    assert worktrees[2].modified == 0
    assert worktrees[0].last_active == "2 days ago"
    assert worktrees[1].last_active_ts == 200
