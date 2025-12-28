from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable, Callable
import asyncio


class FakeGit:
    def __init__(self) -> None:
        self._outputs: dict[tuple[tuple[str, ...], str | None], str] = {}
        self.calls: list[tuple[tuple[str, ...], str | None]] = []

    def set(self, args: list[str], output: str, cwd: str | None = None) -> None:
        self._outputs[(tuple(args), cwd)] = output

    async def __call__(
        self,
        args: list[str],
        *,
        cwd: str | None = None,
        ok_returncodes: Iterable[int] = (0,),
        strip: bool = True,
    ) -> str:
        self.calls.append((tuple(args), cwd))
        out = self._outputs.get((tuple(args), cwd), "")
        return out.strip() if strip else out


async def wait_for_workers(app) -> None:
    while True:
        workers = list(app.workers)
        if not workers:
            return
        try:
            await app.workers.wait_for_complete(workers)
        except Exception as exc:
            if exc.__class__.__name__ != "WorkerCancelled":
                raise


async def wait_for(predicate: Callable[[], bool], *, timeout: float = 1.0, interval: float = 0.01) -> None:
    loop = asyncio.get_running_loop()
    start = loop.time()
    while True:
        if predicate():
            return
        if loop.time() - start > timeout:
            raise AssertionError("timed out waiting for condition")
        await asyncio.sleep(interval)
