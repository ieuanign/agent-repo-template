"""The one structlog setup: console lines in dev, JSON everywhere else."""

import sys

import structlog
from structlog.typing import Processor

from aisvc.platform.config import AppEnv


def configure(app_env: AppEnv) -> None:
    rendering: list[Processor] = (
        [structlog.dev.ConsoleRenderer()]
        if app_env == "dev"
        else [structlog.processors.dict_tracebacks, structlog.processors.JSONRenderer()]
    )
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso", utc=True),
            *rendering,
        ],
        # Bound to the stdout current at this call, so each main call (and each test) captures its own.
        logger_factory=structlog.PrintLoggerFactory(file=sys.stdout),
        cache_logger_on_first_use=False,
    )
