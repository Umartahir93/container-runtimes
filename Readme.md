
# Linux Namespace Sandbox (Go)

This project demonstrates how to create a **minimal isolated Linux process** using **multiple namespaces** in Go.
It launches a child process that can run any command (typically `/bin/bash`) inside newly created namespaces, similar to how container runtimes (like Docker) bootstrap containers.

---

## What This Program Does

This Go program:

* Acts as a **parent process** that sets up Linux namespaces
* Re-executes itself to spawn a **child process** in the new namespaces
* Creates **new Linux namespaces** for isolation
* Maps the current user to **root inside the namespace** using user namespaces
* Allows the child process to set its **hostname** independently of the host

The result is a shell or command that behaves as **root** inside the namespaces but is safely isolated from the host system.

---

## Namespaces Used

The program creates the following namespaces:

| Namespace                  | Purpose                                                    |
| -------------------------- | ---------------------------------------------------------- |
| **UTS** (`CLONE_NEWUTS`)   | Isolates hostname and domain name                          |
| **PID** (`CLONE_NEWPID`)   | Provides a new process ID space (child process sees PID 1) |
| **Mount** (`CLONE_NEWNS`)  | Gives a separate mount table                               |
| **IPC** (`CLONE_NEWIPC`)   | Isolates System V IPC and POSIX message queues             |
| **User** (`CLONE_NEWUSER`) | Allows UID/GID remapping (root inside namespace)           |

---

## User Namespace & UID/GID Mapping

The program maps:

```text
Container UID 0 (root)
→ Host UID <current user>
```

This allows the process to:

* Act as **root inside the namespace**
* Remain **unprivileged on the host**
* Create additional namespaces safely

---

## How It Works (High-Level Flow)

```
Parent Go Process
└─ exec /proc/self/exe "child" with namespace flags
   └─ Child sets hostname
   └─ Executes command (e.g., /bin/bash) in isolated namespaces
```

* **Parent process:** sets up all namespace flags and UID/GID mappings
* **Child process:** runs inside the isolated namespaces, sets hostname, and executes the intended command

---

## Example Behavior

Inside the namespace shell:

```bash
whoami        # root
id            # uid=0
ps            # PID 1 is the shell
hostname      # can be changed without affecting host
```

On the host:

* Host processes and hostname remain unchanged
* Filesystem is shared unless further isolation is added

---

## Features Demonstrated

* Linux namespaces for isolation
* Re-executing the current binary (`/proc/self/exe`) to implement **parent/child roles**
* Child process hostname isolation using UTS namespace
* Minimal container-like behavior without full container runtime setup
* Rootless UID/GID mapping using user namespaces

---

## What This Program Does NOT Do

This program **does not**:

* Change the root filesystem (`chroot` or `pivot_root`)
* Provide full filesystem isolation
* Provide network isolation
* Enforce cgroup-based resource limits

It is intended **only as an educational example**.

---

## Requirements

* Linux kernel with `unprivileged_userns_clone=1`
* Go installed (Go 1.18+ recommended)
* Root privileges are **optional** but may be needed for some namespace operations

---

## Build & Run

Build the program:

```bash
go build -o myuts
```

Run the program as root (or unprivileged if your kernel allows user namespaces):

```bash
sudo ./myuts parent /bin/bash
```

* `parent` → triggers the parent process logic
* `/bin/bash` → command executed in the child process inside isolated namespaces

---

## Next Steps

To extend this into a **full container runtime**, we could add:

* Mount propagation and `pivot_root` isolation
* `/proc` and `/sys` filesystem setup
* Network namespace isolation
* Cgroup support for CPU, memory, and I/O limits