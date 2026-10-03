import grpc
from ai.v1 import info_pb2, info_pb2_grpc


def test_get_info_names_the_service_and_its_build_version(channel: grpc.Channel) -> None:
    stub = info_pb2_grpc.InfoServiceStub(channel)

    resp = stub.GetInfo(info_pb2.GetInfoRequest(), timeout=5)

    # No _buildinfo.py exists outside an image build.
    assert resp == info_pb2.GetInfoResponse(service="ai", version="dev")
