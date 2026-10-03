import grpc
import probe
import pytest
import structlog
from ai.v1 import info_pb2
from structlog.testing import capture_logs
from structlog.typing import EventDict

from aisvc.rpc.interceptors import GENERIC_MESSAGE


def call(
    channel: grpc.Channel, method: str, metadata: tuple[tuple[str, str], ...] = ()
) -> info_pb2.GetInfoResponse:
    stub = channel.unary_unary(
        method,
        request_serializer=info_pb2.GetInfoRequest.SerializeToString,
        response_deserializer=info_pb2.GetInfoResponse.FromString,
    )
    return stub(info_pb2.GetInfoRequest(), timeout=5, metadata=metadata)


def raise_and_capture(channel: grpc.Channel, metadata: tuple[tuple[str, str], ...] = ()) -> list[EventDict]:
    """Call the probe's Raise and return the error events written during it."""
    method = f"/{probe.SERVICE}/Raise"
    with capture_logs(processors=[structlog.contextvars.merge_contextvars]) as logs:
        with pytest.raises(grpc.RpcError):
            call(channel, method, metadata)
    return [log for log in logs if log["log_level"] == "error" and log.get("method") == method]


def test_an_unhandled_exception_answers_internal_with_the_generic_message(channel: grpc.Channel) -> None:
    with pytest.raises(grpc.RpcError) as raised:
        call(channel, f"/{probe.SERVICE}/Raise")

    error = raised.value
    assert isinstance(error, grpc.Call)
    assert error.code() == grpc.StatusCode.INTERNAL
    assert error.details() == GENERIC_MESSAGE
    assert probe.MARKER not in (error.details() or "")


def test_an_unhandled_exception_is_logged_once_with_its_traceback(channel: grpc.Channel) -> None:
    method = f"/{probe.SERVICE}/Raise"
    with capture_logs(processors=[structlog.processors.format_exc_info]) as logs:
        with pytest.raises(grpc.RpcError):
            call(channel, method)

    errors = [log for log in logs if log["log_level"] == "error" and log.get("method") == method]
    assert len(errors) == 1
    assert "ProbeError" in errors[0]["exception"]


def test_an_unregistered_method_still_answers_unimplemented(channel: grpc.Channel) -> None:
    with pytest.raises(grpc.RpcError) as raised:
        call(channel, f"/{probe.SERVICE}/Missing")

    error = raised.value
    assert isinstance(error, grpc.Call)
    assert error.code() == grpc.StatusCode.UNIMPLEMENTED


def test_the_error_line_carries_the_received_request_id(channel: grpc.Channel) -> None:
    errors = raise_and_capture(channel, (("x-request-id", "req-from-backend"),))

    assert len(errors) == 1
    assert errors[0]["request_id"] == "req-from-backend"


def test_a_call_without_a_request_id_is_assigned_one(channel: grpc.Channel) -> None:
    errors = raise_and_capture(channel)

    assert len(errors) == 1
    assert isinstance(errors[0]["request_id"], str)
    assert errors[0]["request_id"] != ""


def test_a_request_id_does_not_reach_a_later_call(channel: grpc.Channel) -> None:
    raise_and_capture(channel, (("x-request-id", "first-call-id"),))
    later = raise_and_capture(channel)

    assert len(later) == 1
    assert later[0]["request_id"] != "first-call-id"
