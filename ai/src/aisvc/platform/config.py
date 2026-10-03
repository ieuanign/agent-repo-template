"""ai's settings, validated from the environment mapping main receives."""

from collections.abc import Mapping
from typing import Any, Literal

from pydantic import Field, ValidationError
from pydantic_settings import BaseSettings, PydanticBaseSettingsSource, SettingsConfigDict

AppEnv = Literal["dev", "staging", "production"]


class ConfigError(Exception):
    pass


class Settings(BaseSettings):
    model_config = SettingsConfigDict(frozen=True, extra="ignore", case_sensitive=True)

    app_env: AppEnv = Field(alias="APP_ENV")
    port: int = Field(default=50051, ge=1, le=65535, alias="PORT")

    @classmethod
    def settings_customise_sources(
        cls,
        settings_cls: type[BaseSettings],
        init_settings: PydanticBaseSettingsSource,
        env_settings: PydanticBaseSettingsSource,
        dotenv_settings: PydanticBaseSettingsSource,
        file_secret_settings: PydanticBaseSettingsSource,
    ) -> tuple[PydanticBaseSettingsSource, ...]:
        # Only the mapping handed to load; os.environ, .env and secrets dirs would break test hermeticity.
        return (init_settings,)


def load(environ: Mapping[str, str]) -> Settings:
    """Validate environ, reporting every bad variable in one ConfigError."""
    # Any, as values are strings until validated; the typed constructor would reject them.
    values: dict[str, Any] = dict(environ)
    try:
        return Settings(**values)
    except ValidationError as e:
        # loc carries the alias, so messages name the variable rather than the field.
        msgs = [
            f"{err['loc'][0]}: " + ("required" if err["type"] == "missing" else f"invalid: {err['msg']}")
            for err in e.errors()
        ]
        raise ConfigError("; ".join(msgs)) from None
