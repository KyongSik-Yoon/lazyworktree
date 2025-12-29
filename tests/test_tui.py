from __future__ import annotations

import asyncio
from typing import Callable, TypeVar

import pytest
from textual.app import App
from textual.widgets import DataTable, Input

from lazyworktree.app import GitWtStatus
from lazyworktree.config import AppConfig
from lazyworktree.screens import CommitScreen

T = TypeVar("T")


async def wait_for(
    condition: Callable[[], T],
    timeout: float = 5.0,
    interval: float = 0.1,
) -> T:
    start = asyncio.get_event_loop().time()
    while True:
        if res := condition():
            return res
        if asyncio.get_event_loop().time() - start > timeout:
            pytest.fail("timeout waiting for condition")
        await asyncio.sleep(interval)


async def wait_for_workers(app: App, timeout: float = 10.0) -> None:
    start = asyncio.get_event_loop().time()
    while app.workers:
        if asyncio.get_event_loop().time() - start > timeout:
            pytest.fail("timeout waiting for workers")
        await asyncio.sleep(0.1)


@pytest.mark.asyncio
async def test_tui_basic(fake_repo, monkeypatch) -> None:
    monkeypatch.chdir(fake_repo.root)

    config = AppConfig(worktree_dir=str(fake_repo.worktree_root.parent))
    app = GitWtStatus(config=config)
    async with app.run_test():
        await wait_for_workers(app)
        table = app.query_one("#worktree-table", DataTable)
        await wait_for(lambda: table.row_count > 0)
        # 2 feature worktrees + main
        assert table.row_count == 3


@pytest.mark.asyncio
async def test_tui_commit_view_and_create_worktree(fake_repo, monkeypatch) -> None:
    monkeypatch.chdir(fake_repo.root)

    config = AppConfig(worktree_dir=str(fake_repo.worktree_root.parent))
    app = GitWtStatus(config=config)
    async with app.run_test() as pilot:
        await wait_for_workers(app)
        table = app.query_one("#worktree-table", DataTable)
        initial_rows = table.row_count

        await pilot.press("3")
        await wait_for(lambda: getattr(app.focused, "id", None) == "log-pane")
        log_table = app.query_one("#log-pane", DataTable)
        await wait_for(lambda: log_table.row_count > 0)
        await pilot.press("enter")
        await wait_for_workers(app)
        await wait_for(lambda: isinstance(app.screen, CommitScreen))
        await pilot.press("q")
        await wait_for(lambda: not isinstance(app.screen, CommitScreen))

        await pilot.press("c")
        await wait_for(lambda: isinstance(app.focused, Input))
        await pilot.press(
            "a",
            "n",
            "o",
            "t",
            "h",
            "e",
            "r",
            "-",
            "b",
            "r",
            "a",
            "n",
            "c",
            "h",
            "enter",
        )

        # Wait for the next screen to be ready
        await pilot.pause(0.5)
        await wait_for(lambda: isinstance(app.focused, Input))
        await pilot.press("enter")

        await pilot.pause(1.0)
        await wait_for_workers(app)

        assert table.row_count == initial_rows + 1
    assert (fake_repo.worktree_root / "another-branch").exists()
