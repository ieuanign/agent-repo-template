"""Request-ID tagging of each unary call's log lines, and the error guard that answers INTERNAL."""

import secrets
from collections.abc import Callable

import grpc
import structlog

GENERIC_MESSAGE = "internal error"
REQUEST_ID_HEADER = "x-request-id"
# The shape of backend's crypto/rand.Text(), so both services' logs carry one ID shape.
_ID_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
_ID_LENGTH = 26

_log = structlog.get_logger()


class ErrorGuard(grpc.ServerInterceptor):
    def intercept_service[Req, Resp](
        self,
        continuation: Callable[[grpc.HandlerCallDetails], grpc.RpcMethodHandler[Req, Resp] | None],
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler[Req, Resp] | None:
        handler = continuation(handler_call_details)
        # Streaming handlers pass untouched: rewrapping drops the non-blocking flag health's Watch needs.
        if handler is None or handler.unary_unary is None:
            return handler
        behaviour = handler.unary_unary
        method = handler_call_details.method

        def guarded(request: Req, context: grpc.ServicerContext) -> Resp:
            try:
                return behaviour(request, context)
            except Exception:
                # Exception text may carry personal data, so it is logged here and never sent.
                _log.exception("rpc failed", method=method)
                context.abort(grpc.StatusCode.INTERNAL, GENERIC_MESSAGE)

        return grpc.unary_unary_rpc_method_handler(
            guarded,
            request_deserializer=handler.request_deserializer,
            response_serializer=handler.response_serializer,
        )


class RequestId(grpc.ServerInterceptor):
    def intercept_service[Req, Resp](
        self,
        continuation: Callable[[grpc.HandlerCallDetails], grpc.RpcMethodHandler[Req, Resp] | None],
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler[Req, Resp] | None:
        handler = continuation(handler_call_details)
        # Streaming handlers pass untouched: rewrapping drops the non-blocking flag health's Watch needs.
        if handler is None or handler.unary_unary is None:
            return handler
        behaviour = handler.unary_unary

        def tagged(request: Req, context: grpc.ServicerContext) -> Resp:
            # Bound on the thread that writes this call's lines, and unbound however the call ends.
            with structlog.contextvars.bound_contextvars(request_id=_request_id(context)):
                return behaviour(request, context)

        return grpc.unary_unary_rpc_method_handler(
            tagged,
            request_deserializer=handler.request_deserializer,
            response_serializer=handler.response_serializer,
        )


def _request_id(context: grpc.ServicerContext) -> str:
    received = next((value for key, value in context.invocation_metadata() if key == REQUEST_ID_HEADER), None)
    if isinstance(received, str) and received:
        return received
    return "".join(secrets.choice(_ID_ALPHABET) for _ in range(_ID_LENGTH))
