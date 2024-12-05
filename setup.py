import os
import pathlib
import sys
from glob import glob
from typing import List, Tuple

from setuptools import Distribution, setup

sys.path.insert(0, pathlib.Path(__file__).parent.joinpath("mlflow_go").as_posix())
from lib import build_lib


def _prune_go_files(path: str):
    for root, dirnames, filenames in os.walk(path, topdown=False):
        for filename in filenames:
            if filename.endswith(".go"):
                os.unlink(os.path.join(root, filename))
        for dirname in dirnames:
            try:
                os.rmdir(os.path.join(root, dirname))
            except OSError:
                pass


def get_platform():
    goos = os.getenv("TARGET_GOOS")
    goarch = os.getenv("TARGET_GOARCH")
    plat = f"{goos}_{goarch}"
    if plat == "darwin_amd64":
        return "macosx_10_13_x86_64"
    elif plat == "darwin_arm64":
        return "macosx_11_0_arm64"
    elif plat == "linux_amd64":
        return "manylinux_2_17_x86_64.manylinux2014_x86_64.musllinux_1_1_x86_64"
    elif plat == "linux_arm64":
        return "manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64"
    elif plat == "windows_amd64":
        return "win_amd64"
    elif plat == "windows_arm64":
        return "win_arm64"
    else:
        raise ValueError("not supported platform.")


def finalize_distribution_options(dist: Distribution) -> None:
    dist.has_ext_modules = lambda: True

    # this allows us to set the tag for the wheel without the python version
    bdist_wheel_base_class = dist.get_command_class("bdist_wheel")

    class bdist_wheel_go(bdist_wheel_base_class):
        def get_tag(self) -> Tuple[str, str, str]:
            return "py3", "none", get_platform()

    dist.cmdclass["bdist_wheel"] = bdist_wheel_go

    # this allows us to build the go binary and add the Go source files to the sdist
    build_base_class = dist.get_command_class("build")

    class build_go(build_base_class):
        def initialize_options(self) -> None:
            self.editable_mode = False
            self.build_lib = None

        def finalize_options(self) -> None:
            self.set_undefined_options("build_py", ("build_lib", "build_lib"))

        def run(self) -> None:
            if not self.editable_mode:
                _prune_go_files(self.build_lib)
                build_lib(
                    pathlib.Path("."),
                    pathlib.Path(self.build_lib).joinpath("mlflow_go"),
                )

        def get_source_files(self) -> List[str]:
            return ["go.mod", "go.sum", *glob("pkg/**/*.go", recursive=True)]

    dist.cmdclass["build_go"] = build_go
    build_base_class.sub_commands.append(("build_go", None))


Distribution.finalize_options = finalize_distribution_options

setup()
