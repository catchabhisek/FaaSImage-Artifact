# Setup and Execution Guide

Follow the steps below to build and run the modified kernel module, collect container file access data, refactor images, and run the supporting services.

---

### 1. Build and Load the Modified OverlayFS Kernel Module

1. Download the kernel source for your **currently installed kernel** from the official repository.  
2. Replace the `fs/overlayfs` folder in the kernel source with the contents of `kernel/overlayfs` from this project.  
3. Rebuild and load the modified module:

```bash
make CONFIG_OVERLAY_FS=m -C ./ M=./fs/overlayfs modules
sudo cp ./fs/overlayfs/overlay.ko /lib/modules/$(uname -r)/kernel/fs/overlayfs/overlay.ko
sudo zstd /lib/modules/$(uname -r)/kernel/fs/overlayfs/overlay.ko
sudo modprobe -r overlay
sudo modprobe overlay
```

---

### 2. Collect File Access Data During Function Initialization

1. Start your function container.  
2. All files accessed during initialization are logged to `/var/log/syslog`.  
3. Copy the file paths you need into a text file; this file will be used as input to the image builder notebook.

---

### 3. Build Refactored Images and Library Partitions

1. Go to the `faasimage/builder` directory.  
2. Open `builder.ipynb` in Jupyter Notebook on the **registry node**.  
3. Execute the notebook cells to produce the refactored image and library partition.

---

### 4. Run the Package Repository Server

1. Go to the `faasimage/package_manager` directory.  
2. Run the package repository server to host packages on the registry node:

```bash
go run server.go
```

---

### 5. Run the On-Demand File Fetching User Application

Compile and run the user application responsible for fetching missing files:

```bash
gcc kernel/userapp/app_code/app.c -o app
sudo ./app
```

---

### 6. Start the Remote Storage Server For On-Demand File Fetching

On the remote storage node, run the server that hosts image files so they can be sent to users on request:

```bash
python3 kernel/userapp/server.py
```

---

### 7. Build and Run the Modified Docker Daemon

On the node that will run containers with refactored images:

```bash
sudo bash install_binary.sh
sudo dockerd -s overlay2
```

This starts the Docker service with the modified image pull operation.
