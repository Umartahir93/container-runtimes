# Minimal Rootless Container with Networking

This project demonstrates how to build a **rootless Linux container from scratch** using Go. It achieves process isolation through namespaces and provides internet connectivity to the unprivileged container using user-mode networking.

---

## 🚀 Key Features

* **Rootless Execution:** Runs entirely without `sudo` by leveraging User Namespaces.
* **Process Isolation:** Uses 6 different Linux namespaces to "jail" the process.
* **User-Mode Networking:** Provides full internet access inside the container using `slirp4netns`.
* **Dynamic Synchronization:** Uses interface polling to ensure the network is ready before the containerized command executes.

---

## 🛠 Namespaces Used

| Namespace | Flag | Purpose |
| --- | --- | --- |
| **User** | `CLONE_NEWUSER` | Maps your host user to `root` inside the container. |
| **Network** | `CLONE_NEWNET` | Isolates the network stack (IPs, routes, etc.). |
| **UTS** | `CLONE_NEWUTS` | Allows the container to have its own hostname. |
| **PID** | `CLONE_NEWPID` | The containerized process thinks it is PID 1. |
| **Mount** | `CLONE_NEWNS` | Provides a private mount table and `chroot` environment. |
| **IPC** | `CLONE_NEWIPC` | Prevents the container from accessing host shared memory. |

---

## 🌐 Networking Architecture

Since an unprivileged user cannot create bridge interfaces on the host, this project uses **slirp4netns**.

1. The **Parent** spawns the child with a private Network Namespace.
2. The **Parent** starts `slirp4netns` on the host, pointing it at the child's PID.
3. `slirp4netns` creates a virtual `tap0` interface inside the container.
4. The **Child** polls for `tap0`, brings up the `lo` (loopback) interface, and finally executes the shell.

---

## 📋 Prerequisites

1. **Linux Kernel:** Must support unprivileged user namespaces.
2. **slirp4netns:** Install via your package manager:
```bash
sudo apt install slirp4netns  # Ubuntu/Debian
sudo dnf install slirp4netns  # Fedora

```


3. **ICMP Permissions (Optional):** To allow `ping` to work rootless, run this on your host:
```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"

```



---

## 🏃 How to Run

1. **Prepare the Rootfs:**
   Ensure you have a directory named `rootfs` with a basic Linux distribution (like Alpine) and a valid DNS config:
```bash
mkdir -p rootfs/etc
echo "nameserver 8.8.8.8" > rootfs/etc/resolv.conf

```


2. **Build and Execute:**
```bash
go build -o gocontainer main.go
./gocontainer parent /bin/sh

```


3. **Test Networking:**
   Inside the container shell:
```bash
ip addr           # See tap0 and lo
ping google.com # Verify internet access

```
---

## ⚠️ Known Limitations

* **Filesystem:** Uses `chroot` for simplicity. For better isolation, `pivot_root` is preferred but requires more complex mount setups in rootless mode.
* **Resource Limits:** Currently, the container can consume unlimited CPU and RAM.

---

## 🔜 Next Step: Resource Control (Cgroups v2)

The next phase of this project is implementing **Cgroups (Control Groups)**. This will allow us to:

* Limit Memory usage (e.g., "This container can only use 100MB RAM").
* Limit CPU shares (e.g., "This container gets only 10% of the CPU").
* Limit the number of processes (PIDs) to prevent fork bombs.