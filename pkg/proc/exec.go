package proc

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const unixSocketFile = "/tmp/tunify-temp"

func Exec(executable string, args []string) (*os.Process, error) {
	executablePath, err := exec.LookPath(executable)
	if err != nil {
		return nil, fmt.Errorf("can not find executable: %w", err)
	}
	pwd, _ := os.Getwd()
	childPID, err := syscall.ForkExec(executablePath, args, &syscall.ProcAttr{
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

func ExecSC(port int, remoteIP, inType, outType string) (*os.Process, error) {
	//socat UDP-RECVFROM:53,bind=127.0.0.1,fork UNIX-SENDTO:/tmp/your_unix_socket

	//socat UDP-LISTEN:5553,reuseaddr,fork UNIX-CLIENT:/tmp/dns_socke
	//socat UNIX-LISTEN:/tmp/dns_socke,unlink-early,fork UDP-SENDTO:127.0.0.53:53
	var scSource, scSink string

	if inType == "UDP" {
		scSource = fmt.Sprintf("UDP-LISTEN:%d,bind=%s,reuseaddr,fork", port, remoteIP)
	} else if inType == "UNIX" {
		scSource = fmt.Sprintf("UNIX-LISTEN:%s%d,unlink-early,fork", unixSocketFile, port)
	} else {
		return nil, fmt.Errorf("invalid protocol %s", inType)
	}

	if outType == "UDP" {
		scSink = fmt.Sprintf("UDP-SENDTO:%s:%d", remoteIP, port)
	} else if outType == "UNIX" {
		scSink = fmt.Sprintf("UNIX-CLIENT:%s%d", unixSocketFile, port)
	} else {
		return nil, fmt.Errorf("invalid protocol %s", outType)
	}

	return Exec("socat", []string{
		"socat",
		"-T5",
		scSource,
		scSink,
	})

	// _, err := Exec("socat", []string{
	// 	fmt.Sprintf("UDP-LISTEN:%d,reuseaddr,fork", portFrom),
	// 	fmt.Sprintf("UNIX-CLIENT:%s%d", unixSocketFile, portFrom),
	// })
	// if err != nil {
	// 	return err
	// }
	// _, err = Exec("socat", []string{
	// 	fmt.Sprintf("UDP-LISTEN:%d,reuseaddr,fork", portFrom),
	// 	fmt.Sprintf("UNIX-CLIENT:%s%d", unixSocketFile, portFrom),
	// })

}
