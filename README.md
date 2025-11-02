# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) equipped with [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- All ordinary protocols<br>
- <b>MVLESS</b> (designed for [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (udp-based protocol removed from xray-core in 2024-Q4 but remain in [Mahsa-Core](https://github.com/GFW-knocker/Xray-core))<br>



## Install on Linux VPS

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/master/install.sh)
```

## Install Specific Version

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## Install on Windows
download & run .exe binary from release<br>
follow readme to open panel on browser

## Encounter Bug?
- clear cookies and site data for the panel (on the left side of address-bar)
- reinstall panel on vps instead of upgrading
