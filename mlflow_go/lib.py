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
        cwd=src_dir.resolve().as_posix(),
        env=env,
    )
    return out_path


def _get_lib():
    import cffi

    # initialize cffi
    ffi = cffi.FFI()
    ffi.cdef("""
        extern int64_t LaunchServer(void* configData, int configSize);

        extern int64_t CreateArtifactsService(void* configData, int configSize);
        extern void DestroyArtifactsService(int64_t id);
        extern int64_t CreateModelRegistryService(void* configData, int configSize);
        extern void DestroyModelRegistryService(int64_t id);
        extern int64_t CreateTrackingService(void* configData, int configSize);
        extern void DestroyTrackingService(int64_t id);

        extern void* ModelRegistryServiceGetLatestVersions(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);

        extern void* TrackingServiceGetExperimentByName(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceCreateExperiment(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceGetExperiment(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceDeleteExperiment(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceCreateRun(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceSearchRuns(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);
        extern void* TrackingServiceLogBatch(int64_t serviceID,
            void* requestData, int requestSize, int* responseSize);

        void free(void*);
    """)

    # check if the library exists and load it
    path = pathlib.Path(
        os.environ.get("MLFLOW_GO_LIBRARY_PATH", pathlib.Path(__file__).parent.as_posix())
    ).joinpath(_get_lib_name())
    if path.is_file():
        return ffi.dlopen(path.as_posix())

    logging.getLogger(__name__).warn("Go library not found, building it now")

    # create temporary directory
    tmpdir = tempfile.TemporaryDirectory()
    atexit.register(tmpdir.cleanup)

    # build the library and load it
    path = build_lib(
        pathlib.Path(__file__).parent.parent.joinpath("pkg", "lib"), pathlib.Path(tmpdir.name)
    )
    return ffi.dlopen(path.as_posix())


_lib = None


def get_lib():
    global _lib
    if _lib is None:
        _lib = _get_lib()
    return _lib


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("src")
    parser.add_argument("out")
    args = parser.parse_args()

    build_lib(pathlib.Path(args.src), pathlib.Path(args.out))
