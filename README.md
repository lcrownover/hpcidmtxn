# hpcidmtxn

## Installation

Building both client and server:

```bash
make
sudo make install
```

Building only the client:

```bash
make client
sudo make install_client
```

Building only the server:

```bash
make server
sudo make install_server
```

## Running the server with Systemd

In the `systemd` directory, copy the hpcidmtxn.service file to `/etc/systemd/system/hpcidmtxn.service`.
Ensure that you have a local user named `hpcidmtxn`, or use another user if you modify the service file.

To get it started:

```bash
systemctl daemon-reload
systemctl start hpcidmtxn
```

To make it run on boot:

```bash
systemctl enable hpcidmtxn
```
