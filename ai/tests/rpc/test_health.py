import grpc
from conftest import Running
from grpc_health.v1 import health, health_pb2

from aisvc.rpc.server import MAX_WORKERS

SERVING = health_pb2.HealthCheckResponse.SERVING
NOT_SERVING = health_pb2.HealthCheckResponse.NOT_SERVING


def check(channel: grpc.Channel, name: str) -> health_pb2.HealthCheckResponse.ServingStatus.ValueType:
    # typeshed's HealthStub declares no methods, so Check goes through the typed unary_unary.
    call = channel.unary_unary(
        f"/{health.SERVICE_NAME}/Check",
        request_serializer=health_pb2.HealthCheckRequest.SerializeToString,
        response_deserializer=health_pb2.HealthCheckResponse.FromString,
    )
    return call(health_pb2.HealthCheckRequest(service=name), timeout=5).status


def test_readiness_and_info_are_serving_once_started(channel: grpc.Channel) -> None:
    assert check(channel, "") == SERVING
    assert check(channel, "ai.v1.InfoService") == SERVING


def test_drain_turns_every_name_not_serving(running: Running, channel: grpc.Channel) -> None:
    running.assembled.health.drain()

    assert check(channel, "") == NOT_SERVING
    assert check(channel, "ai.v1.InfoService") == NOT_SERVING


def test_open_watch_streams_leave_the_workers_free(channel: grpc.Channel) -> None:
    watch = channel.unary_stream(
        f"/{health.SERVICE_NAME}/Watch",
        request_serializer=health_pb2.HealthCheckRequest.SerializeToString,
        response_deserializer=health_pb2.HealthCheckResponse.FromString,
    )
    streams = [watch(health_pb2.HealthCheckRequest(service=""), timeout=10) for _ in range(MAX_WORKERS)]
    try:
        assert [next(stream).status for stream in streams] == [SERVING] * MAX_WORKERS
        assert check(channel, "") == SERVING
    finally:
        for stream in streams:
            stream.cancel()
