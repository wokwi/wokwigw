# Wokwi IoT Network Gateway

> Connect your Wokwi simulated IoT Devices (e.g. ESP32) to you local network!

For installation and usage instructions, check out the [Wokwi ESP32 WiFi Guide](https://docs.wokwi.com/guides/esp32-wifi#the-private-gateway)

## Usage

Run `wokwigw` and configure Wokwi to use the Private IoT Server:

- In [wokwi.com](https://wokwi.com) - Press "F1" in the code editor and select "Enable Private Wokwi IoT Gateway".
- In [Wokwi for VS Code](https://docs.wokwi.com/vscode/getting-started) - Add the following line to your `wokwi.toml` file:

```toml
[net]
gateway="ws://localhost:9011"
```

### Port forwarding

The `wokwigw` tool can forward ports from your local machine to the simulated device. For example, if you have a web server running on port 80 on your simulated device, you can forward port 8080 on your local machine to port 80 on the simulated device:

```bash
wokwigw --forward 8080:10.13.37.2:80
```

To forward a UDP port, add the `udp:` prefix. For instance, the following command will forward UDP port 8888 on your local machine to UDP port 1234 on the simulated device:

```bash
wokwigw --forward udp:8888:10.13.37.2:1234
```

You can repeat the `--forward` flag multiple times to forward multiple ports.

### Connecting from the simulation to your local machine

To connect from the simulation to your local machine (that is the machine running wokwigw), use the host `host.wokwi.internal`. For example, if you are running an HTTP server on port 1234 on your computer, you can connect to it from within the simulator using the URL http://host.wokwi.internal:1234/.

### Bridge mode

The bridge mode is an advanced feature that allows you to connect your simulated device to your local network. The simulated device will get an IP address on your local network, and you can connect to it using the IP address.

The bridge mode works on Linux (root required) and Windows (A driver is required, see below).

Port forwarding is not supported in bridge mode, since the simulated device is now connected to your local network and you can directly connect to it using the IP address.

#### Linux setup

1. Run `wokwigw --bridge`
2. This will create a virtual bridge interface on your machine. Run the following commands as root to configure the bridge interface:

```bash
sudo ip link add br0 type bridge
sudo ip link set eth0 master br0
sudo ip link set dev tap0 up
sudo ip link set tap0 master br0
sudo ip link set dev br0 up
```

Replace `eth0` with the name of your local network interface. WiFi interfaces may not work in bridge mode (if you find otherwise, please let us know!).

#### Windows setup

1. Install the [windows TAP driver](https://build.openvpn.net/downloads/releases/tap-windows-9.24.2-I601-Win10.exe)
2. Run `wokwigw --bridge`
3. Go to the "Network & Internet" settings in the Windows Settings app, and select "More network adapter options".
4. Find a "Local Area Connection" interface named "TAP-Windows Adapter V9". Shift-click on it to select it.
5. Find your local network interface (e.g. "Ethernet" or "Wi-Fi") and shift-click on it to select it.
6. Right click, and select "Bridge connections" from the context menu.

## Building

```
make
```

The compiled binaries go into the `bin` directory, as follows:

- wokwigw.exe - Windows, x86_64
- wokwigw-darwin - Mac OS X, x86_64
- wokwigw-darwin_arm64 - Mac OS X, arm64
- wokwigw-linux - Linux, x86_64
- wokwigw-linux_arm64 - Linux, arm64

### Testing

```
make test
```

### Cloud build environment (Gitpod)

Gitpod allows you to edit the code, build the project in the cloud, and then download the compiled binary. Here are the instructions:

1. [Open this project in Gitpod](https://gitpod.io/#https://github.com/wokwi/wokwigw). You may need to authenticate using your GitHub account.
2. You'll get an online code editor where you can make changes to the source code. Make your code changes.
3. Go to Gitpod's built-in terminal and type `make` to compile the project (you can also type `make test` to run the tests).

You can download the compiled binaries from the `bin` directory by locating them in the file explorer, right-clicking the binary you want to download, and selecting "Download...".
