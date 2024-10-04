package command

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/mlflow/mlflow-go/pkg/utils"
)

type processGroupCmd struct {
	*exec.Cmd
	job windows.Handle
}

const PROCESS_ALL_ACCESS = 2097151

func newProcessGroupCommand(ctx context.Context, cmd *exec.Cmd) (*processGroupCmd, error) {
	logger := utils.GetLoggerFromContext(ctx)

	// Get the job object handle
	jobHandle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create job object: %w", err)
	}

	// Set the job object to kill processes when the job is closed
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err = windows.SetInformationJobObject(
		jobHandle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info))); err != nil {
		return nil, fmt.Errorf("failed to set job object information: %w", err)
	}

	// Terminate the job object (which will terminate all processes in the job)
	cmd.Cancel = func() error {
		logger.Debug("Closing job object to terminate command process group")

		return windows.CloseHandle(jobHandle)
	}

	// Create the process in a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	return &processGroupCmd{Cmd: cmd, job: jobHandle}, nil
}

func (pgc *processGroupCmd) Start() error {
	// Start the command
	if err := pgc.Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Get the process handle
	hProc, err := windows.OpenProcess(PROCESS_ALL_ACCESS, true, uint32(pgc.Process.Pid))
	if err != nil {
		return fmt.Errorf("failed to open process: %w", err)
	}
	defer windows.CloseHandle(hProc)

	// Assign the process to the job object
	if err := windows.AssignProcessToJobObject(pgc.job, hProc); err != nil {
		return fmt.Errorf("failed to assign process to job object: %w", err)
	}

	return nil
}
