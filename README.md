# Wokwi IoT Network Gateway

> Connect your Wokwi simulated IoT Devices (e.g. ESP32) to you local network!

For installation and usage instructions, check out the [Wokwi ESP32 WiFi Guide](https://docs.wokwi.com/guides/esp32-wifi#the-private-gateway)

## Build

```
make
```

The compiled binaries go into the `bin` directory, as follows:

- wokwigw.exe - Windows, x86_64
- wokwigw-darwin - Mac OS X, x86_64
- wokwigw-darwin_arm64 - Mac OS X, arm64
- wokwigw-linux - Linux, x86_64
- wokwigw-linux_arm64 - Linux, arm64

## Test

```
make test
```

## Cloud build environment (Gitpod)

Gitpod allows you to edit the code, build the project in the cloud, and then download the compiled binary. Here are the instructions:

1. [Open this project in Gitpod](https://gitpod.io/#https://github.com/wokwi/wokwigw). You may need to authenticate using your GitHub account.
2. You'll get an online code editor where you can make changes to the source code. Make your code changes.
3. Go to Gitpod's built-in terminal and type `make` to compile the project (you can also type `make test` to run the tests).

You can download the compiled binaries from the `bin` directory by locating them in the file explorer, right-clicking the binary you want to download, and selecting "Download...".
