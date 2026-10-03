"""ai.v1.InfoService: which service answers and the version it was built as."""

import grpc
from ai.v1 import info_pb2, info_pb2_grpc

from aisvc.platform import buildinfo


class InfoServicer(info_pb2_grpc.InfoServiceServicer):
    def GetInfo(
        self, request: info_pb2.GetInfoRequest, context: grpc.ServicerContext
    ) -> info_pb2.GetInfoResponse:
        # The commit stays out: info must reveal nothing an attacker could use.
        return info_pb2.GetInfoResponse(service="ai", version=buildinfo.read().version)
