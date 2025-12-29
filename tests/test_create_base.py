import shutil
import asyncio
import pytest
from unittest.mock import patch, MagicMock, AsyncMock
from lazyworktree.app import GitWtStatus
from lazyworktree.config import AppConfig
from lazyworktree.screens import InputScreen


@pytest.mark.asyncio
async def test_create_worktree_with_base(fake_repo, tmp_path, monkeypatch):
    monkeypatch.chdir(fake_repo.root)
    if shutil.which("git") is None:
        pytest.skip("git not found")

    config = AppConfig(
        worktree_dir=str(fake_repo.worktree_root.parent),
    )
    app = GitWtStatus(config=config)

    # Mock subprocess to capture arguments
    # We want to verify `git worktree add -b <name> <path> <base>`

    mock_process = MagicMock()
    mock_process.returncode = 0
    mock_process.communicate = AsyncMock(return_value=(b"", b""))

    real_create_subprocess_exec = asyncio.create_subprocess_exec

    captured_args = []

    async def side_effect(program, *args, **kwargs):
        if program == "git" and args[0] == "worktree" and args[1] == "add":
            captured_args.append(args)
            # Create directory manually since we mock the command
            # The app expects it to exist after command
            # Actually app calls git which creates it.
            # But we are mocking it.
            # However, before that app calls os.makedirs(parent).
            # Git creates the leaf.
            # Let's just let it return success.
            return mock_process
        return await real_create_subprocess_exec(program, *args, **kwargs)

    with patch("asyncio.create_subprocess_exec", side_effect=side_effect):
        async with app.run_test() as pilot:
            await pilot.pause()

            await pilot.press("c")
            await pilot.pause(0.1)
            assert isinstance(app.screen, InputScreen)

            # Type name
            await pilot.press("n", "e", "w", "-", "f", "e", "a", "t", "enter")
            await pilot.pause(0.5)

            # Now should be second screen for base
            assert isinstance(app.screen, InputScreen)
            # Default is "main" (selected).
            # Type "feature1" to override
            await pilot.press("f", "e", "a", "t", "u", "r", "e", "1", "enter")

            await pilot.pause(0.5)

            # Verify captured args
            # Expected: ('worktree', 'add', '-b', 'new-feat', '.../new-feat', 'feature1')
            assert len(captured_args) == 1
            args = captured_args[0]
            assert args[2] == "-b"
            assert args[3] == "new-feat"
            assert args[5] == "feature1"


@pytest.mark.asyncio
async def test_create_worktree_default_base(fake_repo, tmp_path, monkeypatch):
    monkeypatch.chdir(fake_repo.root)
    config = AppConfig(worktree_dir=str(fake_repo.worktree_root.parent))
    app = GitWtStatus(config=config)

    mock_process = MagicMock()
    mock_process.returncode = 0
    mock_process.communicate = AsyncMock(return_value=(b"", b""))
    real_create_subprocess_exec = asyncio.create_subprocess_exec
    captured_args = []

    async def side_effect(program, *args, **kwargs):
        if program == "git" and args[0] == "worktree" and args[1] == "add":
            captured_args.append(args)
            return mock_process
        return await real_create_subprocess_exec(program, *args, **kwargs)

    with patch("asyncio.create_subprocess_exec", side_effect=side_effect):
        async with app.run_test() as pilot:
            await pilot.pause()

            await pilot.press("c")
            await pilot.pause(0.1)
            await pilot.press(
                "d", "e", "f", "a", "u", "l", "t", "-", "b", "a", "s", "e", "enter"
            )
            await pilot.pause(0.5)

            # Hit enter again to accept default "main"
            await pilot.press("enter")
            await pilot.pause(0.5)

            assert len(captured_args) == 1
            args = captured_args[0]
            # args[5] should be "main"
            assert args[5] == "main"
