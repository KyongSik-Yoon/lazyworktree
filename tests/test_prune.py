import shutil
import os
import pytest
from lazyworktree.app import GitWtStatus
from lazyworktree.config import AppConfig
from lazyworktree.models import PRInfo
from lazyworktree.screens import ConfirmScreen


@pytest.mark.asyncio
async def test_prune_merged_success(fake_repo, tmp_path, monkeypatch):
    monkeypatch.chdir(fake_repo.root)
    if shutil.which("git") is None:
        pytest.skip("git not found")

    config = AppConfig(
        worktree_dir=str(fake_repo.worktree_root.parent),
    )
    app = GitWtStatus(config=config)

    # Mock worktrees with PR status
    # We need to manually set the worktrees because PR fetching is async/external
    # and we want to control the state.

    # We will let the app initialize, then override the worktrees list.

    async with app.run_test() as pilot:
        await pilot.pause()

        # Override worktrees with mocked data
        app.worktrees[0]  # feature1 or main depending on sort
        # Ensure we find the non-main ones
        features = [w for w in app.worktrees if not w.is_main]
        if not features:
            pytest.fail("No feature worktrees found")

        target_wt = features[0]
        # Attach a MERGED PR to it
        target_wt.pr = PRInfo(
            number=123, state="MERGED", title="Merged Feature", url="http://test"
        )

        # Another one OPEN
        if len(features) > 1:
            features[1].pr = PRInfo(
                number=124, state="OPEN", title="Open Feature", url="http://test"
            )

        # Trigger prune
        await pilot.press("X")

        await pilot.pause(0.5)
        assert isinstance(app.screen, ConfirmScreen)
        assert "Merged Feature" in str(app.screen.message) or target_wt.branch in str(
            app.screen.message
        )

        # Confirm
        # ConfirmScreen has Cancel (primary) and Confirm (error) buttons.
        # We need to click "Confirm" or navigate to it.
        # Confirm is the second button usually, or we can use id="confirm"
        await pilot.click("#confirm")

        await pilot.pause(1.0)

        # Verify target path is gone
        assert not os.path.exists(target_wt.path)

        # Verify open one is still there
        if len(features) > 1:
            assert os.path.exists(features[1].path)


@pytest.mark.asyncio
async def test_prune_merged_no_candidates(fake_repo, tmp_path, monkeypatch):
    monkeypatch.chdir(fake_repo.root)
    config = AppConfig(worktree_dir=str(fake_repo.worktree_root.parent))
    app = GitWtStatus(config=config)

    async with app.run_test() as pilot:
        await pilot.pause()

        # Ensure no PRs or only OPEN PRs
        for wt in app.worktrees:
            wt.pr = None

        await pilot.press("X")
        await pilot.pause(0.5)

        # Should NOT be confirm screen
        assert not isinstance(app.screen, ConfirmScreen)
        # Should have notification "No merged worktrees found"
        # Accessing private notifications for test
        if hasattr(app, "_notifications"):
            msgs = [str(n.message) for n in app._notifications]
            assert any("No merged worktrees found" in m for m in msgs)
