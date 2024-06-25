import os
import shutil
import subprocess
from glob import glob
from typing import List, Tuple

from setuptools import Distribution, setup


def _is_go_installed() -> bool:
    try:
        subprocess.check_call(
            ["go", "version"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
        return True
    except Exception:
        return False


def finalize_distribution_options(dist: Distribution) -> None:
    go_installed = _is_go_installed()

    dist.has_ext_modules = lambda: super(Distribution, dist).has_ext_modules() or go_installed

    # this allows us to set the tag for the wheel based on GOOS and GOARCH
    if go_installed:
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
                shutil.rmtree(os.path.join(self.build_lib, "mlflow_go", "go"), ignore_errors=True)
                if go_installed:
                    env = os.environ.copy()
                    env.update(
                        {
                            "CGO_ENABLED": "1",
                        }
                    )
                    subprocess.check_call(
                        [
                            "go",
                            "build",
                            "-trimpath",
                            "-ldflags",
                            "-w -s",
                            "-o",
                            os.path.join(self.build_lib, "mlflow_go", "go", "libmlflow.so"),
                            "-buildmode",
                            "c-shared",
                            "./mlflow_go/go/extension",
                        ],
                        env=env,
                    )

        def get_source_files(self) -> List[str]:
            return ["go.mod", "go.sum", *glob("mlflow_go/go/**/*.go", recursive=True)]

    dist.cmdclass["build_go"] = build_go
    build_base_class.sub_commands.append(("build_go", None))


Distribution.finalize_options = finalize_distribution_options

setup()
