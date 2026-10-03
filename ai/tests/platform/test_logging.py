import io
import json
from contextlib import redirect_stdout

import pytest
import structlog

from aisvc.platform.config import AppEnv
from aisvc.platform.logging import configure


@pytest.mark.parametrize("app_env", ["staging", "production"])
def test_configure_renders_json_outside_dev(app_env: AppEnv, capsys: pytest.CaptureFixture[str]) -> None:
    configure(app_env)

    structlog.get_logger().info("started", service="ai")

    line = json.loads(capsys.readouterr().out)
    assert line["event"] == "started"
    assert line["service"] == "ai"
    assert line["level"] == "info"


def test_configure_renders_console_lines_in_dev(capsys: pytest.CaptureFixture[str]) -> None:
    configure("dev")

    structlog.get_logger().info("started", service="ai")

    out = capsys.readouterr().out
    assert "started" in out
    assert "service" in out and "ai" in out
    with pytest.raises(json.JSONDecodeError):
        json.loads(out)


def test_configure_writes_to_the_stdout_current_at_each_call() -> None:
    first, second = io.StringIO(), io.StringIO()
    with redirect_stdout(first):
        configure("production")
        structlog.get_logger().info("one")
    with redirect_stdout(second):
        configure("production")
        structlog.get_logger().info("two")

    assert json.loads(first.getvalue())["event"] == "one"
    assert json.loads(second.getvalue())["event"] == "two"


@pytest.mark.parametrize("app_env", ["dev", "production"])
def test_configure_renders_bound_context_only_while_bound(app_env: AppEnv) -> None:
    out = io.StringIO()
    with redirect_stdout(out):
        configure(app_env)
        with structlog.contextvars.bound_contextvars(request_id="REQ-7"):
            structlog.get_logger().info("inside")
            structlog.get_logger().info("also inside")
        structlog.get_logger().info("after")

    lines = out.getvalue().splitlines()
    assert len(lines) == 3
    assert "REQ-7" in lines[0] and "REQ-7" in lines[1]
    assert "REQ-7" not in lines[2]
