"""grpc.health.v1: "" is readiness, beside one name per registered service, all drained together."""

import grpc
from grpc_health.v1 import health, health_pb2

# Inside the container checks' 5 s timeout.
PROBE_TIMEOUT_SECONDS = 2


class Health:
    def __init__(self) -> None:
        # The servicer's constructor already sets "" SERVING.
        self.servicer = health.HealthServicer()
        self._names = [""]

    def register(self, name: str) -> None:
        self.servicer.set(name, health_pb2.HealthCheckResponse.SERVING)
        self._names.append(name)

    def drain(self) -> None:
        """Turn every name NOT_SERVING, so connected clients see it before the stop."""
        # Not enter_graceful_shutdown: it silently ignores every later set.
        for name in self._names:
            self.servicer.set(name, health_pb2.HealthCheckResponse.NOT_SERVING)


def probe(port: int) -> str | None:
    """Check "" on localhost; None when SERVING, else why not."""
    with grpc.insecure_channel(f"localhost:{port}") as channel:
        # typeshed's HealthStub declares no methods, so Check goes through the typed unary_unary.
        check = channel.unary_unary(
            f"/{health.SERVICE_NAME}/Check",
            request_serializer=health_pb2.HealthCheckRequest.SerializeToString,
            response_deserializer=health_pb2.HealthCheckResponse.FromString,
        )
        try:
            status = check(health_pb2.HealthCheckRequest(service=""), timeout=PROBE_TIMEOUT_SECONDS).status
        except grpc.RpcError as e:
            return f"{e.code()}: {e.details()}"
    if status != health_pb2.HealthCheckResponse.SERVING:
        return f"status {health_pb2.HealthCheckResponse.ServingStatus.Name(status)}"
    return None
