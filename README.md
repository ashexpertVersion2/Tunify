### Tunify

this small project is a simple tool to bypass linux routing table and run an isolated program in separated network
namespace in a way that every traffic goes through a given interface.

#### Build and Run

```bash
make tunify
```

```bash
./bin/tunify <interface_name(wlo1, eth0 or ...)> <application>
```
## Limitations
- `"net.ipv4.ip_forward"` must be enabled.
- Program must have root access or `CAP_NET` to alter iptable rules and create veth.
- `socat` must be installed.

