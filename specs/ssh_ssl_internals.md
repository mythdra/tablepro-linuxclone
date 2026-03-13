# SSH & SSL Implementation Algorithms

This document describes exactly how TablePro handles complex network tunneling and security, outlining the precise mechanisms that the Qt/C++ rewrite must replicate.

## 1. SSH Tunneling (`SSHTunnelManager`)

TablePro does **not** link against `libssh2`. Instead, it invokes the host system's native `/usr/bin/ssh` binary using `Process` (`QProcess` equivalent). This provides maximum compatibility with the user's local `~/.ssh/config` and existing agent setups.

### Execution Algorithm
1. **Port Selection**: Randomly selects a local port candidate between `60,000` and `65,000`.
2. **Process Arguments**:
   - `["-N"]`: Do not execute a remote command (port forwarding only).
   - `["-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"]`: Disables strict key checking.
   - `["-o", "ServerAliveInterval=60", "-o", "ServerAliveCountMax=3"]`: Keep-alive logic.
   - `["-L", "127.0.0.1:{localPort}:{remoteHost}:{remotePort}"]`: The actual port mapping.
   - `["-J", "user1@jump1:22,user2@jump2:2222"]`: Binds chained jump hosts natively.
3. **Authentication Overrides:**
   - If using **Password Auth** or **Encrypted Private Key** (requiring a passphrase), the system native `ssh` command normally blocks stdin waiting for a user prompt. 
   - *The Askpass Hack:* TablePro generates a temporary bash script in `/tmp/` named `ssh_askpass_<UUID>` with executable (`0o700`) permissions. The script simply echoes the password: `#!/bin/bash\necho 'the_password'`.
   - TablePro then injects `{ "SSH_ASKPASS": "/tmp/askpass...", "SSH_ASKPASS_REQUIRE": "force", "DISPLAY": ":0" }` into the `QProcess` environment. The `ssh` binary natively calls this script to get the password without hanging the UI.
4. **Health Checking (Startup):**
   - The app polls `127.0.0.1:{localPort}` every 250ms using a raw TCP Socket (`connect`). 
   - It also runs `/usr/sbin/lsof -nP -iTCP:{localPort} -sTCP:LISTEN -t` to ensure the listening process ID specifically belongs to the spawned `ssh` child process tree (prevents race conditions if another app stole the port).
5. **Session Death Tracking:**
   - A background task wakes up every 90 seconds to verify the `Process.isRunning`. If false, it broadcasts a tunnel death event, forcing the UI to display a disconnected state.

## 2. SSL/TLS Configurations

Unlike SSH (which proxies the entire TCP connection), SSL/TLS is handled strictly by the individual Database Drivers (libpq, libmysqlclient). 

### Storage & Serialization
- The UI exposes `caCertPath`, `clientCertPath`, and `clientKeyPath`.
- The `DatabaseConnection` model passes these absolute strings directly down to the plugin driver.
- The C++ library (e.g., `libpq`) handles reading the actual files and validating the certificates.
- The only logic natively handled by the core application is ensuring the user has granted file system read permissions to those paths in highly-sandboxed environments (macOS Sandbox / App Store).

## 3. Keychain Storage (`ConnectionStorage`)
- Passwords (for DB, SSH, and SSH Keys) are not saved in plaintext JSON.
- They are securely stored in the macOS Keychain (`SecItemAdd`, `SecItemCopyMatching`).
- The Qt rewrite must utilize `qtkeychain` (`QKeychain::WritePasswordJob`, `QKeychain::ReadPasswordJob`) to replicate this cross-platform behavior (mapping to Windows Credential Manager, Linux Secret Service, macOS Keychain).
