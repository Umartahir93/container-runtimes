package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: go run main.go parent <command> <args>")
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

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"name=shashank"}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting child: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parent: Child PID is %d. Setting up slirp4netns...\n", cmd.Process.Pid)
	go setupNetworking(cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Child exited with error: %v\n", err)
	}
}

func setupNetworking(pid int) {
	args := []string{
		"--configure",
		"--mtu=1500",
		fmt.Sprintf("%d", pid),
		"tap0",
	}

	netCmd := exec.Command("slirp4netns", args...)
	if err := netCmd.Run(); err != nil {
		fmt.Printf("slirp4netns error: %v\n", err)
	}
}

func child() {
	err := enableCgroup()
	if err != nil {
		fmt.Printf("Error enabling cgroup: %v\n", err)
	}

	syscall.Sethostname([]byte("myhost"))

	rootfs, _ := filepath.Abs("./rootfs_arm")
	if err := setupRoot(rootfs); err != nil {
		fmt.Printf("Root setup error: %v\n", err)
		os.Exit(1)
	}

	if err := mountProc(); err != nil {
		fmt.Printf("Mount proc error: %v\n", err)
		os.Exit(1)
	}

	exec.Command("ip", "link", "set", "lo", "up").Run()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"PS1=\\u@\\h:\\w# ", "PATH=/bin:/usr/bin:/sbin"}

	err = waitForNetwork()
	if err != nil {
		fmt.Printf("Network wait error: %v\n", err)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command execution error: %v\n", err)
		os.Exit(1)
	}
}

func setupRoot(newRoot string) error {
	if err := syscall.Chroot(newRoot); err != nil {
		return err
	}
	return os.Chdir("/")
}

func mountProc() error {
	target := "/proc"
	os.MkdirAll(target, 0755)
	return syscall.Mount("proc", target, "proc", 0, "")
}

func waitForNetwork() error {
	maxWait := 15 * time.Second
	timeStarted := time.Now()

	for {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
		if len(interfaces) > 1 {
			return nil
		}
		if time.Since(timeStarted) > maxWait {
			return fmt.Errorf("timeout waiting for network interface to become available")
		}
		time.Sleep(1 * time.Second)
	}
}

func enableCgroup() error {
	// path where your cgorup is mounted
	cgroups := "/home/umar-tahir/Development/mygrp"
	pids := filepath.Join(cgroups, "child")

	if err := ioutil.WriteFile(filepath.Join(pids, "memory.max"), []byte("2M"), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(pids, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700); err != nil {
		return err
	}
	return nil
}
