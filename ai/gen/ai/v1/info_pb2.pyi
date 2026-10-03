from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class GetInfoRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetInfoResponse(_message.Message):
    __slots__ = ("service", "version")
    SERVICE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    service: str
    version: str
    def __init__(self, service: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...
