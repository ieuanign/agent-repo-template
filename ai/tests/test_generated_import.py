# Importing runs protobuf's runtime-version check, so this fails when the lock drifts below the generator.
def test_generated_info_contract_imports() -> None:
    from ai.v1 import info_pb2, info_pb2_grpc

    response = info_pb2.GetInfoResponse(service="ai", version="0.1.0")

    assert info_pb2.GetInfoResponse.FromString(response.SerializeToString()) == response
    assert hasattr(info_pb2_grpc, "InfoServiceStub")
