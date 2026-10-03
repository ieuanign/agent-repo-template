import threading

import grpc
import pytest
from ai.v1 import info_pb2, info_pb2_grpc

from aisvc.platform.config import Settings
from aisvc.rpc.server import assemble, stop


def test_stop_drains_in_flight_rpcs_and_refuses_new_ones() -> None:
    entered, release = threading.Event(), threading.Event()

    def slow(request: bytes, context: grpc.ServicerContext) -> bytes:
        entered.set()
        release.wait(5)
        return request

    # Owned by this test, since stopping it must not affect the shared fixture.
    assembled = assemble(Settings(APP_ENV="dev"), ())
    server = assembled.server
    handler: grpc.RpcMethodHandler[bytes, bytes] = grpc.unary_unary_rpc_method_handler(slow)
    server.add_generic_rpc_handlers((grpc.method_handlers_generic_handler("test.Slow", {"Call": handler}),))
    port = server.add_insecure_port("127.0.0.1:0")
    server.start()
    with grpc.insecure_channel(f"127.0.0.1:{port}") as channel:
        call: grpc.UnaryUnaryMultiCallable[bytes, bytes] = channel.unary_unary("/test.Slow/Call")
        in_flight = call.future(b"ping", timeout=10)
        assert entered.wait(5)

        terminated = stop(assembled)

        with pytest.raises(grpc.RpcError) as refused:
            info_pb2_grpc.InfoServiceStub(channel).GetInfo(info_pb2.GetInfoRequest(), timeout=5)
        assert refused.value.code() == grpc.StatusCode.UNAVAILABLE

        release.set()
        assert in_flight.result(timeout=5) == b"ping"
        assert terminated.wait(5)
