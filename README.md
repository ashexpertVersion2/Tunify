### Tunify

this small project is a simple tool to bypass linux routing table and run an isolated program in separated network
namespace in a way that every traffic goes through a given interface.

#### Build and Run

```bash
make tunify
```

```bash
./bin/tunify <interface_name(wlo1, eth0 or ...)> <application>
