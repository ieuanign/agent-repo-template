import pytest

from aisvc.platform.config import ConfigError, Settings, load


@pytest.mark.parametrize(
    ("environ", "want"),
    [
        ({"APP_ENV": "dev"}, Settings(APP_ENV="dev", PORT=50051)),
        ({"APP_ENV": "staging", "PORT": "9000"}, Settings(APP_ENV="staging", PORT=9000)),
        (
            {"APP_ENV": "production", "PORT": "65535", "HOME": "/x"},
            Settings(APP_ENV="production", PORT=65535),
        ),
    ],
)
def test_load_reads_only_the_given_environ(environ: dict[str, str], want: Settings) -> None:
    assert load(environ) == want


@pytest.mark.parametrize(
    ("environ", "names"),
    [
        ({}, ["APP_ENV: required"]),
        ({"APP_ENV": ""}, ["APP_ENV: invalid"]),
        ({"APP_ENV": "prod"}, ["APP_ENV: invalid"]),
        ({"APP_ENV": "dev", "PORT": "x"}, ["PORT: invalid"]),
        ({"APP_ENV": "dev", "PORT": "0"}, ["PORT: invalid"]),
        ({"APP_ENV": "dev", "PORT": "65536"}, ["PORT: invalid"]),
        ({"PORT": "x"}, ["APP_ENV: required", "PORT: invalid"]),
    ],
)
def test_load_reports_every_bad_variable_by_its_env_name(environ: dict[str, str], names: list[str]) -> None:
    with pytest.raises(ConfigError) as exc:
        load(environ)

    msg = str(exc.value)
    for name in names:
        assert name in msg
    assert "app_env" not in msg
    assert "port" not in msg.replace("PORT", "")


def test_load_ignores_the_process_environment(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("APP_ENV", "dev")

    with pytest.raises(ConfigError, match="APP_ENV"):
        load({})
