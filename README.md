# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) equipped with [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- All ordinary protocols<br>
- <b>MVLESS</b> (designed for [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (udp-based protocol removed from xray-core in 2024-Q4 but remain in [Mahsa-Core](https://github.com/GFW-knocker/Xray-core))<br>
(note: QUIC take too much RAM in high traffic)
- <b>allowInsecure</b> supported (removed from xray-core but kept in [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)): switch in the TLS settings, exported in share links and subscriptions<br>
- <b>Privacy</b>: sponsor banner and redundant ip-api calls removed for privacy reason, public-IP lookup is optional; no third-party domain is contacted except GitHub for updates and the integrations you enable<br>



## Install latest version on Linux VPS

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## Install v2.8.4-mahsa-r2 on Linux VPS

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## Install Specific Version

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## Install on Windows
download & run .exe binary from release<br>
follow readme to open panel on browser

## Encounter Bug?
- <b>clear cookies and site data</b> from the browser (on the left side of address-bar)
- reinstall panel on vps instead of upgrading
