import atexit
import os
import pathlib
import subprocess
import sys
import tempfile

import cffi


def _get_lib():
    ffi = cffi.FFI()
    ffi.cdef("""
    extern int64_t LaunchServer(void* configData, int configSize);
    """)

    pkg = pathlib.Path(__file__).parent
    ext = ".so"
    if sys.platform == "win32":
        ext = ".dll"
    elif sys.platform == "darwin":
        ext = ".dylib"

    name = "libmlflow-go" + ext
    path = pkg.joinpath(name)
    if path.is_file():
        return ffi.dlopen(path.as_posix())

    # create temporary directory
    tmpdir = tempfile.TemporaryDirectory()
    atexit.register(tmpdir.cleanup)
    path = pathlib.Path(tmpdir.name).joinpath(name)
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
            path.as_posix(),
            "-buildmode",
            "c-shared",
            pkg.joinpath("extension").as_posix(),
        ],
        env=env,
    )
    return ffi.dlopen(path.as_posix())


lib = _get_lib()
