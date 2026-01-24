# Minimal Rootless Container

This project demonstrates how to create a **minimal isolated Linux process** using **multiple namespaces** in Go.
It launches a child process that can run **any command** inside newly created namespaces, similar to how container runtimes (like Docker or runc) bootstrap containers.

---

## What This Program Does

This Go program:

* Acts as a **parent process** that sets up Linux namespaces
* Re-executes itself to spawn a **child process** in the new namespaces
* Creates **new Linux namespaces** for isolation
* Maps the current user to **root inside the namespace** using user namespaces
* Sets a **hostname** inside the child namespace
* Changes the root filesystem of the child using **`chroot`**

The result is a command (e.g., shell) that behaves as **root inside the isolated namespaces** but is safely contained from the host system.

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
* Create namespaces **without requiring sudo**

> ⚠️ Some kernel features (e.g., networking, certain mounts) still require `CAP_SYS_ADMIN` or elevated privileges.

---

## Filesystem Isolation (chroot vs pivot_root)

### Why `chroot` is used

This project uses **`chroot`** instead of `pivot_root` because:

* `pivot_root` requires **CAP_SYS_ADMIN**, which is not available in rootless mode
* It avoids permission errors when running rootless

### Limitation of `chroot`

* A privileged process can escape chroot using open file descriptors (if not careful)
* `chroot` does not change mount points - it only changes the root directory
* For full container isolation, **`pivot_root` + mount namespace** is preferred when CAP_SYS_ADMIN is available

---

## How It Works (High-Level Flow)

```
Parent Go Process
└─ exec /proc/self/exe "child" with namespace flags
   └─ Child sets hostname
   └─ Performs chroot to rootfs
   └─ Executes intended command (e.g., /bin/sh)
```

* **Parent process:** sets up namespace flags and UID/GID mappings
* **Child process:** runs inside isolated namespaces, sets hostname, changes root, and executes the command

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
* Host filesystem remains isolated from the child

---

## Requirements

* Linux kernel with `unprivileged_userns_clone=1`
* Go installed (Go 1.18+ recommended)
* No root privileges required for this version

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
./mynamespaces parent /bin/sh
```

> Use `/bin/sh` (BusyBox)- Bash may not exist in minimal root filesystems.

### 🍎 Running on macOS
Since macOS doesn't support Linux namespaces, use Colima/Docker with this alias to cross-compile and run your code in a privileged Linux environment:

```bash
# Add this alias to your shell
alias run-linux-go='GOOS=linux GOARCH=arm64 go build -o mynamespaces main.go && docker run --privileged -it -v $(pwd):/app -w /app golang:1.23'
```

```bash
# Execute your container code
run-linux-go ./mynamespaces parent /bin/sh
```

---

## Features Demonstrated

* Linux namespaces for process and resource isolation
* User namespace with UID/GID mapping for rootless containers
* Child process hostname isolation using UTS namespace
* Filesystem isolation using **chroot**
* Running a shell or arbitrary command inside isolated namespaces

---

## What This Program Does NOT Do

This program **does not**:

* Set up networking namespaces
* Apply cgroup-based resource limits
* Provide a full container runtime

---

## Next Steps

To extend this into a **full container runtime**, we will add:

* Networking namespace and virtual interfaces (veth, bridges)
* PID 1 signal handling and reaping
* Resource control using cgroups (CPU, memory, I/O)
