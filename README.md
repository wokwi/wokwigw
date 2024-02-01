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
wokwigw --forward udp:8888:10.37.37.2:1234
```

You can repeat the `--forward` flag multiple times to forward multiple ports.

### Connecting from the simulation to your local machine

To connect from the simulation to your local machine (that is the machine running wokwigw), use the host `host.wokwi.internal`. For example, if you are running an HTTP server on port 1234 on your computer, you can connect to it from within the simulator using the URL http://host.wokwi.internal:1234/.

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
