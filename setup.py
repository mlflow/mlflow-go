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


def finalize_distribution_options(dist: Distribution) -> None:
    dist.has_ext_modules = lambda: True

    # this allows us to set the tag for the wheel without the python version
    bdist_wheel_base_class = dist.get_command_class("bdist_wheel")

    class bdist_wheel_go(bdist_wheel_base_class):
        def get_tag(self) -> Tuple[str, str, str]:
            _, _, plat = super().get_tag()
            return "py3", "none", plat

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
