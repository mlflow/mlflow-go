import logging
import os
import pathlib
import platform
import re
import subprocess
import sys
import tempfile
import shutil

def _get_lib_name() -> str:
    ext = ".so"
    if sys.platform == "win32":
        ext = ".dll"
    elif sys.platform == "darwin":
        ext = ".dylib"
    return "libmlflow-go" + ext


def get_goarch():
    machine = platform.machine().lower()

    if machine in ["x86_64", "amd64"]:
        return "amd64"
    elif machine in ["aarch64", "arm64"]:
        return "arm64"
    elif machine in ["armv7l", "arm"]:
        return "arm"
    else:
        return "unknown"


def get_target_triple(goos, goarch):
    if goos == "linux":
        if goarch == "amd64":
            return "x86_64-linux-gnu"
        elif goarch == "arm64":
            return "aarch64-linux-gnu"
        else:
            raise (f"Could not determine target triple for {goos}, {goarch}")
    elif goos == "windows":
        if goarch == "amd64":
            return "x86_64-windows-gnu"
        elif goarch == "arm64":
            return "aarch64-windows-gnu"
        else:
            raise (f"Could not determine target triple for {goos}, {goarch}")
    else:
        raise (f"Could not determine target triple for {goos}, {goarch}")


def build_lib(src_dir: pathlib.Path, out_dir: pathlib.Path) -> pathlib.Path:
    out_path = out_dir.joinpath(_get_lib_name()).absolute()
    env = os.environ.copy()
    env.update(
        {
            "CGO_ENABLED": "1",
        }
    )

    current_goos = platform.system().lower()
    current_goarch = get_goarch()

    target_goos = os.getenv("TARGET_GOOS", current_goos)
    target_goarch = os.getenv("TARGET_GOARCH", current_goarch)
    env.update({"GOOS": target_goos, "GOARCH": target_goarch})

    if target_goos == "darwin" and current_goos != "darwin":
        raise "it is unsupported to build a Python wheel on Mac on a non-Mac platform"

    if target_goos != "darwin":
        triple = get_target_triple(target_goos, target_goarch)
        logging.getLogger(__name__).info(f"CC={sys.executable} -mziglang cc -target {triple}")
        env.update({"CC": f"{sys.executable} -mziglang cc -target {triple}"})

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
    path = (
        pathlib.Path(
            os.environ.get("MLFLOW_GO_LIBRARY_PATH", pathlib.Path(__file__).parent.as_posix())
        )
        .joinpath(_get_lib_name())
        .absolute()
    )
    if path.is_file():
        return _load_lib(path)

    logger = logging.getLogger(__name__)
    logger.warning("Go library not found, building it now")

    cache_dir = pathlib.Path(tempfile.gettempdir()).joinpath("mlflow_go_lib_cache")
    if os.path.isdir(cache_dir) and os.listdir(cache_dir):
        logger.info(f"Deleting files in {cache_dir}")
        shutil.rmtree(cache_dir)
    
    cache_dir.mkdir(exist_ok=True)

    # build the library in a temporary directory and load it
    with tempfile.TemporaryDirectory() as tmpdir:
        logger.info(f"Building library in {tmpdir}")
        built_path = build_lib(
            pathlib.Path(__file__).parent.parent,
            pathlib.Path(tmpdir),
        )
        cached_path = cache_dir.joinpath(built_path.name)
        shutil.copy(built_path, cached_path)
        shutil.copy(built_path.with_suffix(".h"), cached_path.with_suffix(".h"))
        return _load_lib(cached_path)


def _load_lib(path: pathlib.Path):
    ffi = get_ffi()

    # load from header file
    ffi.cdef(_parse_header(path.with_suffix(".h")))

    # load the library
    return ffi.dlopen(path.as_posix())

def _parse_header(path: pathlib.Path):
    with open(path) as file:
        content = file.read()

    # Find all matches in the header
    functions = re.findall(
        r"extern\s+(?:__declspec\(dllexport\)\s+)?\w+\s*\*?\s+\w+\s*\([^)]*\);",
        content,
        re.MULTILINE,
    )

    # Replace GoInt64 with int64_t in each function
    transformed_functions = [
        func.replace("GoInt64", "int64_t").replace("__declspec(dllexport) ", "")
        for func in functions
    ]

    return "\n".join(transformed_functions)


def _get_ffi():
    import cffi

    return cffi.FFI()


_ffi = None


def get_ffi():
    global _ffi
    if _ffi is None:
        _ffi = _get_ffi()
        _ffi.cdef("void free(void*);")
    return _ffi


_lib = None


def get_lib():
    global _lib
    if _lib is None:
        _lib = _get_lib()
    return _lib


_clib = None


def _get_clib():
    if sys.platform == "win32":
        return get_ffi().dlopen("msvcrt.dll")
    else:
        return get_lib()


def get_clib():
    global _clib
    if _clib is None:
        _clib = _get_clib()
    return _clib


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser("build_lib", description="Build Go library")
    parser.add_argument("src", help="the Go source directory")
    parser.add_argument("out", help="the output directory")
    args = parser.parse_args()

    build_lib(pathlib.Path(args.src), pathlib.Path(args.out))
