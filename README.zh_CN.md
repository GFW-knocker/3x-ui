# MX-UI
搭载 [Mahsa-Core](https://github.com/GFW-knocker/Xray-core) 的 [3X-UI](https://github.com/MHSanaei/3x-ui)<br><br>
- 所有常规协议<br>
- <b>MVLESS</b>（为 [MahsaNG](https://github.com/GFW-knocker/MahsaNG) 设计）<br>
- <b>QUIC</b>（基于 UDP 的协议，2024 年第四季度已从 xray-core 移除，但在 [Mahsa-Core](https://github.com/GFW-knocker/Xray-core) 中保留）<br>
（注意：高流量下 QUIC 会占用大量内存）
- <b>隐私</b>：出于隐私考虑移除了赞助商横幅和多余的 ip-api 调用，公网 IP 查询改为可选；除了用于更新的 GitHub 和你自行启用的集成之外，面板不会联系任何第三方域名<br>



## 在 Linux VPS 上安装最新版本

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## 在 Linux VPS 上安装 v2.8.4-mahsa-r2

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## 安装指定版本

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## 在 Windows 上安装
从 release 下载并运行 .exe 文件<br>
按照 readme 的说明在浏览器中打开面板

## 遇到问题？
- 在浏览器中<b>清除 Cookie 和网站数据</b>（地址栏左侧）
- 在 VPS 上重新安装面板，而不是升级
