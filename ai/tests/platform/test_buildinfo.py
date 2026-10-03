import sys
from types import ModuleType

import pytest

from aisvc.platform.buildinfo import BuildInfo, read


def test_read_reports_dev_without_a_stamped_build() -> None:
    assert read() == BuildInfo(version="dev", commit="dev")


def test_read_reports_the_stamped_build(monkeypatch: pytest.MonkeyPatch) -> None:
    stamped = ModuleType("aisvc.platform._buildinfo")
    stamped.__dict__.update(VERSION="1.2.3", COMMIT="abc1234")
    monkeypatch.setitem(sys.modules, "aisvc.platform._buildinfo", stamped)

    assert read() == BuildInfo(version="1.2.3", commit="abc1234")
