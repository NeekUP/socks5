## Supported only TCP and username/password authorization.
Config example
```yaml
# socks5.yaml
network: "tcp"
address: "127.0.0.1"
port:    7788
auth:    "NO" #NO, PASS
user:    ""
pass:    ""
mtu:     1400
```

RFC:
* [SOCKS Protocol Version 5](https://tools.ietf.org/html/rfc1928)
* [Username/Password Authentication for SOCKS V5](https://tools.ietf.org/html/rfc1929)