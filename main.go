package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected parent or child")
	}

	switch os.Args[1] {
	case "parent":
		parent()
	case "child":
		child()
	default:
		panic("unknown command")
	}

}

// parent function invoked main program which sets up the namespace
func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	// The statements below refer to the input, output and error streams of the process created (cmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//setting an environment variable
	cmd.Env = []string{"name=shashank"}

	// The | operator is bitwise OR, combining multiple flags into a single number which kernel understands to
	// create process with all new namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	err := cmd.Run()
	if err != nil {
		fmt.Printf("E%s\n", err)
	}
}

// this is the child process w copy of the parent program
func child() {
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// the command below sets the h to myhost. Idea here is to show use of UTS namespace.
	err := syscall.Sethostname([]byte("myhost"))
	if err != nil {
		fmt.Printf("E%s\n", err)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Printf("E%s\n", err)
	}

}
