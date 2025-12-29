import pytest

from textual.widgets import DataTable, Input, RichLog

from lazyworktree.app import GitWtStatus
from lazyworktree.config import AppConfig
from lazyworktree.screens import CommitScreen, HelpScreen

from tests.utils import wait_for, wait_for_workers


@pytest.mark.asyncio
async def test_tui_keyboard_flow(fake_repo, monkeypatch) -> None:
    monkeypatch.chdir(fake_repo.root)

    config = AppConfig(worktree_dir=str(fake_repo.worktree_root.parent))
    app = GitWtStatus(config=config)
    async with app.run_test() as pilot:
        await wait_for_workers(app)
        table = app.query_one("#worktree-table", DataTable)
        status_log = app.query_one("#status-pane", RichLog)
        assert table.row_count == 3
        await wait_for(lambda: getattr(app.focused, "id", None) == "worktree-table")

        await pilot.press("tab")
        await wait_for(lambda: getattr(app.focused, "id", None) == "status-pane")
        await pilot.press("tab")
        await wait_for(lambda: getattr(app.focused, "id", None) == "log-pane")
        await pilot.press("1")
        await wait_for(lambda: getattr(app.focused, "id", None) == "worktree-table")

        start_row = table.cursor_row
        await pilot.press("j")
        await wait_for(
            lambda: table.cursor_row is not None and table.cursor_row != start_row
        )

        await pilot.press("?")
        await wait_for(lambda: isinstance(app.screen, HelpScreen))
        await pilot.press("escape")
        await wait_for(lambda: not isinstance(app.screen, HelpScreen))

        await pilot.press("/")
        await wait_for(
            lambda: app.query_one("#filter-container").styles.display == "block"
        )
        filter_input = app.query_one("#filter-input", Input)
        await wait_for(lambda: app.focused is filter_input)
        await pilot.press("f", "e", "a", "t", "u", "r", "e", "1")
        await wait_for(lambda: filter_input.value == "feature1")
        await pilot.press("enter")
        await wait_for(
            lambda: app.query_one("#filter-container").styles.display == "none"
        )
        await wait_for(lambda: getattr(app.focused, "id", None) == "worktree-table")
        assert table.row_count == 1

        await pilot.press("d")
        await wait_for_workers(app)
        await wait_for(lambda: getattr(app.focused, "id", None) == "status-pane")
        assert len(status_log.lines) > 0


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
        await pilot.press("n", "e", "w", "-", "b", "r", "a", "n", "c", "h", "enter")
        await wait_for_workers(app)

        assert table.row_count == initial_rows + 1
        assert (fake_repo.worktree_root / "new-branch").exists()
