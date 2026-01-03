# Linux Namespace Sandbox & Minimal RootFS (Go)

This project demonstrates how to create a **minimal isolated Linux process** using **multiple namespaces** in Go.
It launches a child process that can run **any command** inside newly created namespaces, similar to how container runtimes (like Docker or runc) bootstrap containers.
The program also implements a **pivot_root** operation to isolate the filesystem.

---

## What This Program Does

This Go program:

* Acts as a **parent process** that sets up Linux namespaces
* Re-executes itself to spawn a **child process** in the new namespaces
* Creates **new Linux namespaces** for isolation
* Maps the current user to **root inside the namespace** using user namespaces
* Sets a **hostname** inside the child namespace
* Changes the root filesystem of the child using **pivot_root**
  (the child only sees the contents of `rootfs` after pivot)

The result is a command (e.g., shell) that behaves as **root** inside the isolated namespaces but is safely contained from the host system.

---

## Namespaces Used

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

This allows the child process to:

* Act as **root inside the namespace**
* Remain **unprivileged on the host**
* Safely create additional namespaces without compromising host security

---

## Pivot Root (Filesystem Isolation)

The child process performs a **pivot_root** into a `rootfs` directory:

* The child sees `rootfs` as `/`
* The original host root is hidden at `/.pivot_root` temporarily
* After unmounting `/.pivot_root`, the host filesystem is no longer visible from inside the container

> ⚠️ Make sure `rootfs` contains the binaries you want to run (e.g., `/bin/sh` for BusyBox). Bash may not exist in minimal root filesystems.

---

## How It Works (High-Level Flow)

```
Parent Go Process
└─ exec /proc/self/exe "child" with namespace flags
   └─ Child sets hostname
   └─ Performs pivot_root to rootfs
   └─ Executes intended command (e.g., /bin/sh)
```

* **Parent process:** sets up all namespace flags and UID/GID mappings
* **Child process:** runs inside isolated namespaces, sets hostname, pivots root, and executes the command

---

## Example Behavior

Inside the namespace shell:

```bash
whoami        # root
id            # uid=0 (mapped from host user)
ps            # PID 1 is the shell
hostname      # shows "myhost"
```

On the host:

* Host processes and hostname remain unchanged
* Host filesystem remains intact and isolated from child after pivot

---

## Requirements

* Linux kernel with `unprivileged_userns_clone=1`
* Go installed (Go 1.18+ recommended)
* Root privileges are **optional**, but some operations (like mount or certain pivot_root setups) may require root

---

## Build & Run

Build the program:

```bash
go build -o mynamespaces
```

Prepare a minimal root filesystem (example with BusyBox):

```bash
mkdir -p rootfs/{bin,proc,sys,dev,tmp}
cp /usr/bin/busybox rootfs/bin/
cd rootfs/bin && ./busybox --install .
```

Run the program:

```bash
sudo ./mynamespaces parent /bin/sh
```

* `parent` → triggers the parent process logic
* `/bin/sh` → command executed inside isolated namespaces (use `/bin/sh` for BusyBox rootfs)

---

## Features Demonstrated

* Linux namespaces for process and resource isolation
* User namespace with UID/GID mapping for rootless containers
* Child process hostname isolation using UTS namespace
* Filesystem isolation using **pivot_root**
* Running a shell or arbitrary command inside isolated namespaces

---

## What This Program Does NOT Do

This program **does not**:

* Set up networking namespaces
* Mount `/proc` or `/sys` automatically
* Apply cgroup-based resource limits
* Provide a full container runtime

It is intended **only as an educational example** for understanding Linux namespaces and minimal container isolation.

---

## Next Steps

To extend this into a **full container runtime**, we will add:

* Mount propagation and full filesystem setup (`/proc`, `/sys`)
* Networking namespace and virtual interfaces
* PID 1 signal handling for child processes
* Resource control using cgroups (CPU, memory, I/O)
* Integration with a container image (like BusyBox or Alpine)
