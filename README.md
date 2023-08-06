<div align="center">

<img src="icons/on.png" width="60em">

# Tailscale Systray

This is an unofficial cross-platform system tray application written in Golang for managing Tailscale status, exit nodes, and other related functionalities. Tailscale is a secure and easy-to-use VPN that allows you to securely connect your devices across the internet.

![image](https://github.com/C10udburst/tailscale-systray/assets/18114966/adc69050-a320-4e3d-952b-fcf57185c8d4)

</div>

## Features

- **System Tray Interface**: The application provides a user-friendly system tray interface that allows quick access to Tailscale functionalities directly from the system tray icon.
- **Exit Node Management**: View current exit node and easily switch to a different exit node.
- **Preference Management**: Manage Tailscale preferences like DNS, accept routes, and more.
- **Device List**: View all devices connected to your Tailscale network.
- **Cross-platform**: The application is designed to work seamlessly on major operating systems such as Windows, macOS, and Linux.

## Installation

### Windows

In PowerShell, run the following command:

```powershell
iwr -useb https://raw.githubusercontent.com/C10udburst/tailscale-systray/master/install.txt | iex
```

### macOS

In Terminal, run the following command:

```bash
curl -fsSL https://raw.githubusercontent.com/C10udburst/tailscale-systray/master/install.txt | sh
```

### Linux

In Terminal, run the following command:

```bash
curl -fsSL https://raw.githubusercontent.com/C10udburst/tailscale-systray/master/install.txt | sh
```

## Update

Just run the installation command again to update the application.

## Uninstall

### Windows

Remove the application from the `shell:startup` folder.

### macOS
```bash
rm ~/.local/bin/tailscale-systray
rm ~/Library/LaunchAgents/com.tailscale.tailscale-systray.plist
```

### Linux
```bash
rm ~/.local/bin/tailscale-systray
rm ~/.config/autostart/tailscale-systray.desktop
```

<sub>This application is an unofficial project and is not associated with the official Tailscale project. Use it at your own risk, and the developers are not liable for any potential issues or damages caused by the usage of this application.</sub>

