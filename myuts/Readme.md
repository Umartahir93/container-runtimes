# myuts - Creating a UTS Namespace in Go

## Overview

`myuts` is a simple Go program that demonstrates how **Linux namespaces** work, specifically the **UTS (UNIX Time-sharing System) namespace**, using the Go standard library.

This program launches a new `/bin/bash` shell inside a **new UTS namespace**, allowing the process to have an isolated hostname and domain name, separate from the host system.

This is a **foundational building block** for understanding how containers (like Docker) achieve isolation.

---

## What This Program Demonstrates

* Creating a **new Linux UTS namespace**
* Executing a child process (`/bin/bash`) from Go
* Connecting parent and child I/O streams
* Setting environment variables for a child process
* Verifying namespace isolation using the `/proc` filesystem

---

## Prerequisites

* Linux (namespaces are a Linux kernel feature)
* Go installed (Go 1.16+ recommended)
* Root privileges (required for creating namespaces)

---

## Source Code Explanation

### Key Imports

```go
import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
)
```

* `os/exec` - used to create and run external processes
* `syscall` - used to interact directly with Linux kernel features
* `os` - for environment variables and I/O streams

---

### Creating the Child Process

```go
cmd := exec.Command("/bin/bash")
```

This creates a command structure that will execute `/bin/bash` as a **child process**.

---

### Connecting I/O Streams

```go
cmd.Stdin  = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
```

These lines connect:

* Parent stdin → child stdin
* Child stdout → parent stdout
* Child stderr → parent stderr

This allows you to **interact with the shell normally**, as if you launched it directly.

---

### Setting an Environment Variable

```go
cmd.Env = []string{"name=shashank"}
```

This sets an environment variable inside the child process only.

You can verify it inside the shell:

```bash
echo $name
```

---

### Creating a New UTS Namespace

```go
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS,
}
```

This is the **most important line** in the program.

* `CLONE_NEWUTS` tells the Linux kernel:

  > “Create a new UTS namespace for this process”

This isolates:

* Hostname
* Domain name

The child process will **not share** the UTS namespace with the host.

---

### Running the Process

```go
if err := cmd.Run(); err != nil {
    fmt.Printf("running the /bin/bash command - %s\n", err)
    os.Exit(1)
}
```

This executes the command and waits until the shell exits.

---

## Building the Program

From the project directory:

```bash
go build myuts.go
```

This creates a binary named `myuts`.

---

## Running the Program

Run as root:

```bash
sudo ./myuts
```

You will be dropped into a `/bin/bash` shell running inside a **new UTS namespace**.

---

## Verifying Namespace Creation

### Before running `myuts`

```bash
ls -li /proc/self/ns/uts
```

Example output:

```
uts:[4026531838]
```

### Inside `myuts` shell

```bash
ls -li /proc/self/ns/uts
```

Example output:

```
uts:[4026532505]
```

### Interpretation

* The inode number changed
* This proves a **new UTS namespace was created**
* Different inode = different namespace

---

## Why This Matters

This program demonstrates the **exact kernel mechanism** used by container runtimes:

`myuts` is essentially a **minimal container runtime experiment**.

---

## Key Takeaways

* Linux namespaces are kernel-level isolation primitives
* Go can interact directly with the Linux kernel using `syscall`
* Containers are built from simple, composable kernel features
* Namespace isolation can be verified using `/proc`

---

## Next Steps

Possible extensions to this project covered in seperate package and at the end of the project
you will have a container runtime like `Docker`

---

## License

This project is for educational purposes.

---