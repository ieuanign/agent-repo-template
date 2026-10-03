# A thread-pool gRPC server, not grpc.aio

ai serves gRPC with grpcio's synchronous server over a `ThreadPoolExecutor` of 16 workers, admitting at most 64 concurrent RPCs; beyond that, gRPC itself answers `RESOURCE_EXHAUSTED`. ai hosts AI capabilities of both kinds: I/O-bound modules that wait on a model provider or a vector database, and CPU-bound modules that run inference or image processing in the process. A thread pool serves both without asking every module's author to keep the event loop free: a call that blocks holds one worker while the other 15 keep answering.

## Considered Options

- **`grpc.aio` on asyncio** — rejected: it serves many more concurrent I/O-bound calls per process, but every handler shares one event loop. One forgotten blocking call, whether a synchronous client, a file read or a CPU-bound loop, stalls every RPC in the process until it returns, health checks included, so the deploy would read a busy ai as a dead one.

## Consequences

- CPU-bound work follows one rule, recorded in [`ai/CLAUDE.md`](../../CLAUDE.md): libraries that release the GIL (numpy, onnxruntime, torch) run on the handler thread, and pure-Python CPU-bound work goes to a process pool. The free-threaded build is not an option, because grpcio publishes no free-threaded wheels.
- Volume has a ceiling: 16 workers per process and 64 concurrent RPCs. Serving more means more workers or more copies, not a change of model within a copy.
- Switching to `grpc.aio` later rewrites every servicer as a coroutine, so the choice is made once, before the first module exists.
