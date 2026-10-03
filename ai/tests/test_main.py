import json

import pytest
from conftest import Running

from aisvc.__main__ import main


@pytest.mark.parametrize(
    ("environ", "names"),
    [
        ({}, ["APP_ENV"]),
        ({"APP_ENV": ""}, ["APP_ENV"]),
        ({"APP_ENV": "prod"}, ["APP_ENV"]),
        ({"APP_ENV": "dev", "PORT": "x"}, ["PORT"]),
        ({"APP_ENV": "dev", "PORT": "0"}, ["PORT"]),
        ({"APP_ENV": "dev", "PORT": "65536"}, ["PORT"]),
        ({"APP_ENV": "prod", "PORT": "x"}, ["APP_ENV", "PORT"]),
    ],
)
def test_serve_exits_1_naming_every_bad_variable(
    environ: dict[str, str], names: list[str], capsys: pytest.CaptureFixture[str]
) -> None:
    assert main(["serve"], environ) == 1

    lines = capsys.readouterr().out.splitlines()
    assert len(lines) == 1
    # JSON even when APP_ENV is what failed.
    msg = json.dumps(json.loads(lines[0]))
    for name in names:
        assert name in msg
    assert "app_env" not in msg
    assert "port" not in msg.replace("PORT", "")


def test_unknown_subcommand_is_a_usage_error() -> None:
    with pytest.raises(SystemExit) as exc:
        main(["nope"], {"APP_ENV": "dev"})

    assert exc.value.code == 2


def test_healthcheck_exits_0_silently_against_a_serving_server(
    running: Running, capsys: pytest.CaptureFixture[str]
) -> None:
    assert main(["healthcheck"], {"APP_ENV": "dev", "PORT": str(running.port)}) == 0

    assert capsys.readouterr().out == ""


def test_healthcheck_exits_1_once_drained(running: Running, capsys: pytest.CaptureFixture[str]) -> None:
    running.assembled.health.drain()

    assert main(["healthcheck"], {"APP_ENV": "dev", "PORT": str(running.port)}) == 1
    assert "healthcheck failed" in capsys.readouterr().out


def test_healthcheck_exits_1_when_nothing_answers(
    running: Running, capsys: pytest.CaptureFixture[str]
) -> None:
    running.assembled.server.stop(None).wait(5)

    assert main(["healthcheck"], {"APP_ENV": "dev", "PORT": str(running.port)}) == 1
    assert "healthcheck failed" in capsys.readouterr().out


def test_healthcheck_exits_1_not_2_on_invalid_configuration(capsys: pytest.CaptureFixture[str]) -> None:
    assert main(["healthcheck"], {"APP_ENV": "dev", "PORT": "x"}) == 1

    out = capsys.readouterr().out
    assert "healthcheck failed" in out
    assert "PORT" in out
