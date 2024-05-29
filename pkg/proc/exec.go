package proc

import (
	"os"
	"syscall"
)

func Exec(executablePath, pwd string) (*os.Process, error) {
	childPID, err := syscall.ForkExec(executablePath, os.Args[2:], &syscall.ProcAttr{
		Dir:   pwd,
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}
	proc, err := os.FindProcess(childPID)
	if err != nil {
		return nil, err
	}
	return proc, nil
}
