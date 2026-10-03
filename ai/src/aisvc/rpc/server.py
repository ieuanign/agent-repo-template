"""The one gRPC server assembly, shared by serve and the tests, and its graceful stop."""

import threading
from collections.abc import Sequence
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from typing import Protocol

import grpc
from ai.v1 import info_pb2, info_pb2_grpc
from grpc_health.v1 import health_pb2_grpc

from aisvc.platform.config import Settings
from aisvc.rpc.health import Health
from aisvc.rpc.info import InfoServicer
from aisvc.rpc.interceptors import ErrorGuard, RequestId

# A thread pool, so a blocking call stalls one worker rather than every RPC; gRPC itself
# answers RESOURCE_EXHAUSTED beyond MAX_CONCURRENT_RPCS.
MAX_WORKERS = 16
MAX_CONCURRENT_RPCS = 64
# Under Compose's 30 s stop_grace_period, so in-flight RPCs finish before SIGKILL.
STOP_GRACE_SECONDS = 20


@dataclass(frozen=True)
class Assembled:
    server: grpc.Server
    health: Health


class Registration:
    """What a module's register receives: the server for its servicer, and its health entry."""

    def __init__(self, server: grpc.Server, health: Health) -> None:
        self.grpc_server = server
        self._health = health

    def add_health(self, name: str) -> None:
        """Name the module's proto service; it is SERVING once started and drains with the rest."""
        self._health.register(name)


class Module(Protocol):
    # Modules declare their own narrower server Protocol, since they may not import rpc.
    def register(self, server: Registration) -> None: ...


def assemble(settings: Settings, modules: Sequence[Module]) -> Assembled:
    """Build the server with its servicers and the modules registered; the caller binds and starts it."""
    server = grpc.server(
        ThreadPoolExecutor(max_workers=MAX_WORKERS),
        # First is outermost, so the guard's error line is written with the ID bound.
        interceptors=[RequestId(), ErrorGuard()],
        maximum_concurrent_rpcs=MAX_CONCURRENT_RPCS,
    )
    health = Health()
    health_pb2_grpc.add_HealthServicer_to_server(health.servicer, server)
    info_pb2_grpc.add_InfoServiceServicer_to_server(InfoServicer(), server)
    health.register(info_pb2.DESCRIPTOR.services_by_name["InfoService"].full_name)
    registration = Registration(server, health)
    for module in modules:
        module.register(registration)
    return Assembled(server, health)


def stop(assembled: Assembled) -> threading.Event:
    """Drain health, then refuse new RPCs and let in-flight ones finish; the event is set on termination."""
    assembled.health.drain()
    return assembled.server.stop(STOP_GRACE_SECONDS)
