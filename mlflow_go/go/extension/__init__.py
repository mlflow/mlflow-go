import atexit
import logging
import os
import pathlib
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
            src_dir.resolve().as_posix(),
        ],
        env=env,
    )
    return out_path


def _get_lib():
    import cffi

    # initialize cffi
    ffi = cffi.FFI()
    ffi.cdef("""
    extern int64_t LaunchServer(void* configData, int configSize);
    """)

    # find Go package path
    pkg = pathlib.Path(__file__).parent

    # check if the library exists and load it
    path = pkg.joinpath(_get_lib_name())
    if path.is_file():
        return ffi.dlopen(path.as_posix())

    logging.getLogger(__name__).warn("Go extension library not found, building it now")

    # create temporary directory
    tmpdir = tempfile.TemporaryDirectory()
    atexit.register(tmpdir.cleanup)

    # build the library and load it
    path = build_lib(pkg, pathlib.Path(tmpdir.name))
    return ffi.dlopen(path.as_posix())


_lib = None


def get_lib():
    global _lib
    if _lib is None:
        _lib = _get_lib()
    return _lib
