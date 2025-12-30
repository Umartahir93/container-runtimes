# Linux Namespace Sandbox (Go)

This project demonstrates how to create an isolated Linux process using **multiple namespaces** in Go.
It launches an interactive `/bin/bash` shell inside newly created namespaces, similar to how container runtimes (like Docker) bootstrap containers.

---

## What This Program Does

This Go program:

* Spawns a child process (`/bin/bash`)
* Creates **new Linux namespaces** for that process
* Maps the current user to **root inside the namespace**
* Provides process-level isolation without requiring host root access

The result is a shell that appears to run as **root**, but is safely isolated from the host system.

---

## Namespaces Used

The program creates the following namespaces:

| Namespace                  | Purpose                                                  |
| -------------------------- | -------------------------------------------------------- |
| **UTS** (`CLONE_NEWUTS`)   | Isolates hostname and domain name                        |
| **PID** (`CLONE_NEWPID`)   | Provides a new process ID space (PID 1 inside container) |
| **Mount** (`CLONE_NEWNS`)  | Gives a separate mount table                             |
| **IPC** (`CLONE_NEWIPC`)   | Isolates System V IPC and POSIX message queues           |
| **User** (`CLONE_NEWUSER`) | Allows UID/GID remapping (root inside namespace)         |

---

## User Namespace & UID Mapping

The program maps:

```text
Container UID 0 (root)
→ Host UID <current user>
```

This allows the process to:

* Act as **root inside the namespace**
* Remain **unprivileged on the host**
* Safely create other namespaces without compromising system security

This is a key mechanism used by modern container runtimes.

---

## How It Works (High-Level Flow)

```
Go Parent Process
└─ fork + clone (with namespace flags)
   └─ execve("/bin/bash")
      └─ Bash runs in isolated namespaces
```

The Go process acts as the **parent**, and `/bin/bash` runs as a **child process** inside new namespaces.

---

## What This Program Does NOT Do

This program **does not**:

* Change the root filesystem (`chroot` or `pivot_root`)
* Isolate the filesystem contents
* Provide network isolation
* Act as a full container runtime

Because of this, files on the host filesystem are still visible and accessible (subject to permissions).

---

## Expected Behavior

Inside the namespace shell:

```bash
whoami        # root
id            # uid=0
ps            # PID 1 is bash
hostname      # can be changed without affecting host
```

On the host:

* Host processes and hostname remain unchanged
* Filesystem is shared unless further isolation is added

---

## Why This Is Useful

This program is a **minimal educational example** showing:

* How Linux namespaces work
* How containers begin their lifecycle
* How user namespaces enable rootless containers
* The building blocks behind Docker, containerd, and runc

---

## Requirements

* Linux kernel with `unprivileged_userns_clone=1`
* Go installed (Go 1.18+ recommended)
* No root privileges required

---

## Build & Run

```bash
go build -o mynamespaces
./mynamespaces
```

---

## Next Steps

To turn this into a real container runtime, you would add:

* Mount propagation isolation
* `pivot_root` or `chroot`
* `/proc` mounting
* Network namespace setup
* Cgroups for resource limits

---

## Disclaimer

This project is for **learning purposes only**.
It is **not a secure container runtime** and should not be used in production.