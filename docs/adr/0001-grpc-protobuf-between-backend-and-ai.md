# gRPC and Protobuf between backend and ai

backend calls ai over gRPC from day one rather than starting with HTTP/JSON and migrating later. The contract lives in the root `proto/` directory and is built by buf: Go stubs go into backend and Python stubs into ai. Generated code is committed so a contract change shows up as a diff in review. Lint runs `buf lint`, `buf breaking` and a regenerate-and-diff check, so a stale or backward-incompatible contract fails before merge.
