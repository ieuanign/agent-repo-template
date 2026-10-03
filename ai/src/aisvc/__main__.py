"""Composition root: python -m aisvc serve | healthcheck."""

import argparse
import os
import signal
import sys
from collections.abc import Mapping, Sequence

import structlog

from aisvc.platform import buildinfo, logging
from aisvc.platform.config import ConfigError, load
from aisvc.rpc import health
from aisvc.rpc import server as rpc_server


def serve(environ: Mapping[str, str]) -> int:
    try:
        settings = load(environ)
    except ConfigError as e:
        # JSON regardless, since APP_ENV may be the variable that failed.
        logging.configure("production")
        structlog.get_logger().error("invalid configuration", error=str(e))
        return 1
    logging.configure(settings.app_env)
    log = structlog.get_logger()

    assembled = rpc_server.assemble(settings, ())
    server = assembled.server
    server.add_insecure_port(f"[::]:{settings.port}")
    # SIGINT drains like SIGTERM: watchfiles reloads with it.
    for sig in (signal.SIGTERM, signal.SIGINT):
        signal.signal(sig, lambda *_: rpc_server.stop(assembled))
    server.start()

    build = buildinfo.read()
    # The only place the commit is disclosed.
    log.info(
        "started",
        service="ai",
        version=build.version,
        commit=build.commit,
        environment=settings.app_env,
        port=settings.port,
    )
    server.wait_for_termination()
    return 0


def healthcheck(environ: Mapping[str, str]) -> int:
    """Exit 0 only when this host's server answers "" SERVING; silent unless it fails."""
    try:
        settings = load(environ)
    except ConfigError as e:
        logging.configure("production")
        structlog.get_logger().error("healthcheck failed", reason=f"invalid configuration: {e}")
        return 1
    reason = health.probe(settings.port)
    if reason is None:
        return 0
    logging.configure(settings.app_env)
    structlog.get_logger().error("healthcheck failed", reason=reason)
    return 1


def main(argv: Sequence[str], environ: Mapping[str, str]) -> int:
    parser = argparse.ArgumentParser(prog="aisvc")
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("serve", help="run the gRPC server")
    commands.add_parser("healthcheck", help="exit 0 when the local server is ready")
    args = parser.parse_args(argv)
    if args.command == "healthcheck":
        return healthcheck(environ)
    return serve(environ)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:], os.environ))
