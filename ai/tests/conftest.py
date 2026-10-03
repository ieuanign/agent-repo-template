from collections.abc import Iterator
from typing import NamedTuple

import grpc
import probe
import pytest

from aisvc.platform.config import Settings
from aisvc.rpc.server import Assembled, assemble


class Running(NamedTuple):
    assembled: Assembled
    port: int


@pytest.fixture
def running() -> Iterator[Running]:
    """A server assembled as serve does, plus the probe module, on a loopback port and owned by this test."""
    assembled = assemble(Settings(APP_ENV="dev"), (probe,))
    port = assembled.server.add_insecure_port("127.0.0.1:0")
    assembled.server.start()
    yield Running(assembled, port)
    assembled.server.stop(None).wait(5)


@pytest.fixture
def channel(running: Running) -> Iterator[grpc.Channel]:
    """A real channel to the running server."""
    with grpc.insecure_channel(f"127.0.0.1:{running.port}") as ch:
        yield ch
