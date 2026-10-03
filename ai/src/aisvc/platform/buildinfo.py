"""The version and commit the image build stamped, or dev for both outside an image."""

import importlib
from dataclasses import dataclass


@dataclass(frozen=True)
class BuildInfo:
    version: str
    commit: str


def read() -> BuildInfo:
    # Imported by name so Ruff and mypy pass while the gitignored module is absent.
    try:
        stamped = importlib.import_module("aisvc.platform._buildinfo")
    except ModuleNotFoundError:
        return BuildInfo(version="dev", commit="dev")
    return BuildInfo(version=str(stamped.VERSION), commit=str(stamped.COMMIT))
