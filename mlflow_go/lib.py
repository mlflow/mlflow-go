import atexit
import logging
import os
import pathlib
import re
import subprocess
import sys
import tempfile


def _get_lib_name() -> str:
    ext = ".so"
    if sys.platform == "win32":
        ext = ".dll"
    elif sys.platform == "darwin":
        ext = ".dylib"
    return "libmlflow-go" + ext


def build_lib(src_dir: pathlib.Path, out_dir: pathlib.Path) -> pathlib.Path:
    out_path = out_dir.joinpath(_get_lib_name())
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
            out_path.resolve().as_posix(),
            "-buildmode",
            "c-shared",
            src_dir.joinpath("pkg", "lib").resolve().as_posix(),
        ],
        cwd=src_dir.resolve().as_posix(),
        env=env,
    )
    return out_path


def _get_lib():
    # check if the library exists and load it
    path = pathlib.Path(
        os.environ.get("MLFLOW_GO_LIBRARY_PATH", pathlib.Path(__file__).parent.as_posix())
    ).joinpath(_get_lib_name())
    if path.is_file():
        return _load_lib(path)

    logging.getLogger(__name__).warn("Go library not found, building it now")

    # create temporary directory
    tmpdir = tempfile.TemporaryDirectory()
    atexit.register(tmpdir.cleanup)

    # build the library and load it
    return _load_lib(
        build_lib(
            pathlib.Path(__file__).parent.parent,
            pathlib.Path(tmpdir.name),
        )
    )


def _load_lib(path: pathlib.Path):
    import cffi

    ffi = cffi.FFI()

    # load from header file
    ffi.cdef(_parse_header(path.with_suffix(".h")))

    # load the library
    lib = ffi.dlopen(path.as_posix())

    # make sure the library is closed when the program exits
    atexit.register(ffi.dlclose, lib)

    return lib


def _parse_header(path: pathlib.Path):
    with open(path) as file:
        content = file.read()

    # Find all matches in the header
    functions = re.findall(r"extern\s+\w+\s+\w+\s*\([^)]*\);", content, re.MULTILINE)

    # Replace GoInt64 with int64_t in each function
    transformed_functions = [func.replace("GoInt64", "int64_t") for func in functions]

    return "\n".join(transformed_functions)


_lib = None


def get_lib():
    global _lib
    if _lib is None:
        _lib = _get_lib()
    return _lib


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser("build_lib", description="Build Go library")
    parser.add_argument("src", help="the Go source directory")
    parser.add_argument("out", help="the output directory")
    args = parser.parse_args()

    build_lib(pathlib.Path(args.src), pathlib.Path(args.out))
