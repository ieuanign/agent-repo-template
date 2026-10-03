"""A test-only module: test.v1.ProbeService, registered through register(server) as a real module is."""

from collections.abc import Callable
from typing import Protocol

import grpc
from ai.v1 import info_pb2

SERVICE = "test.v1.ProbeService"
MARKER = "probe-secret-7f3a"


class Server(Protocol):
    # Declared here, since a module may not import rpc.
    @property
    def grpc_server(self) -> grpc.Server: ...

    def add_health(self, name: str) -> None: ...


def _answer(request: info_pb2.GetInfoRequest, context: grpc.ServicerContext) -> info_pb2.GetInfoResponse:
    return info_pb2.GetInfoResponse(service="probe")


class ProbeError(Exception):
    pass


def _raise(request: info_pb2.GetInfoRequest, context: grpc.ServicerContext) -> info_pb2.GetInfoResponse:
    raise ProbeError(MARKER)


def _handler(
    behaviour: Callable[[info_pb2.GetInfoRequest, grpc.ServicerContext], info_pb2.GetInfoResponse],
) -> grpc.RpcMethodHandler[info_pb2.GetInfoRequest, info_pb2.GetInfoResponse]:
    return grpc.unary_unary_rpc_method_handler(
        behaviour,
        request_deserializer=info_pb2.GetInfoRequest.FromString,
        response_serializer=info_pb2.GetInfoResponse.SerializeToString,
    )


def register(server: Server) -> None:
    server.grpc_server.add_generic_rpc_handlers(
        (
            grpc.method_handlers_generic_handler(
                SERVICE, {"Answer": _handler(_answer), "Raise": _handler(_raise)}
            ),
        )
    )
    server.add_health(SERVICE)
