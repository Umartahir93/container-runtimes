# Minimal Rootless Container (Go)

A lightweight, **rootless container runtime** written in Go. It demonstrates how to build a container from scratch using Linux Namespaces, Cgroups v2, and user-mode networking (`slirp4netns`).

---

## 🚀 Key Features

* **100% Rootless:** Runs as a standard user (UID 1000) mapped to `root` (UID 0) inside the container.
* **Networking:** Full internet access using `slirp4netns` (no root bridges required).
* **Resource Control:** Enforces a **2MB Memory Limit** using Cgroups v2.
* **Architecture:** Pre-configured for ARM64 (Apple Silicon/Raspberry Pi) using the included `rootfs_arm`.

---

## 🛠 Dependencies

Ensure you have these installed on your host before running:

1. **slirp4netns** (For networking)
```bash
sudo apt install slirp4netns

```


2. **Go** (Golang 1.18+)
3. **Cgroup v2** enabled kernel (Standard on Ubuntu 20.04+).

---

## ⚙️ Configuration & Setup

### 1. The Filesystem

The project includes a `rootfs_arm` directory containing the Alpine Linux filesystem. **No download is required.**

### 2. Cgroup Mount & Setup (Critical)

You must manually mount the Cgroup v2 filesystem first, and then create the child directory inside it.

Run these commands in your terminal (ensure the path matches the `cgroups` variable in your `main.go`):

```bash
# 1. Create the mount point (Parent folder only)
mkdir -p ~/Development/mygrp

# 2. Mount Cgroup v2 (Requires sudo once)
sudo mount -t cgroup2 none ~/Development/mygrp

# 3. Enable Memory Controller (Crucial)
# This allows sub-folders to actually restrict memory.
echo "+memory" > ~/Development/mygrp/cgroup.subtree_control

# 4. Create the 'child' folder inside the mount
# The Go code expects this folder to exist!
mkdir ~/Development/mygrp/child

# 5. Take Ownership (Crucial for Rootless)
# This ensures your Go program (running as User) can write to these files.
sudo chown -R $USER:$USER ~/Development/mygrp

```

---

## 🏃 Usage

Once the Cgroup directories are created, run the container:

```bash
# Syntax: go run main.go parent <command>
go run main.go parent /bin/sh

```

You should see:

```text
Parent: Child PID is [PID]. Setting up slirp4netns...
/ # 

```

---

## 🧪 Verification

### 1. Test Networking

Inside the container shell, try to reach the internet:

```bash
/ # ping -c 2 google.com
PING google.com (142.250.x.x): 56 data bytes
64 bytes from ...

```

## 🧩 How It Works

1. **Parent Process:**
* Sets up Namespaces (`CLONE_NEWUSER`, `CLONE_NEWNET`, etc.).
* Maps Host UID -> Container Root.
* Starts `slirp4netns` to provide internet connectivity.


2. **Child Process:**
* **Enables Cgroups:** Writes the limit to `.../child/memory.max`.
* **Sets Hostname:** Changes hostname to `myhost`.
* **Chroot:** Jails the file system into `./rootfs_arm`.
* **Mounts /proc:** Required for process visibility.
* **Waits for Network:** Polls `tap0` interface before running the user command.