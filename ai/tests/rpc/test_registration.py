import grpc
import probe
from ai.v1 import info_pb2
from test_health import NOT_SERVING, SERVING, check

from aisvc.platform.config import Settings
from aisvc.rpc.server import assemble


def answer(channel: grpc.Channel) -> info_pb2.GetInfoResponse:
    call = channel.unary_unary(
        f"/{probe.SERVICE}/Answer",
        request_serializer=info_pb2.GetInfoRequest.SerializeToString,
        response_deserializer=info_pb2.GetInfoResponse.FromString,
    )
    return call(info_pb2.GetInfoRequest(), timeout=5)


def test_a_registered_module_answers_its_rpc(channel: grpc.Channel) -> None:
    assert answer(channel).service == "probe"


def test_a_registered_module_health_entry_is_serving(channel: grpc.Channel) -> None:
    assert check(channel, probe.SERVICE) == SERVING


def test_drain_turns_a_module_health_entry_not_serving_with_the_rest() -> None:
    # Owned by this test, since draining it must not affect the shared fixture.
    assembled = assemble(Settings(APP_ENV="dev"), (probe,))
    port = assembled.server.add_insecure_port("127.0.0.1:0")
    assembled.server.start()
    try:
        with grpc.insecure_channel(f"127.0.0.1:{port}") as channel:
            assembled.health.drain()

            for name in ("", "ai.v1.InfoService", probe.SERVICE):
                assert check(channel, name) == NOT_SERVING
    finally:
        assembled.server.stop(None).wait(5)
