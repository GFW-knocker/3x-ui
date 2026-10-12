# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) equipado con [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- Todos los protocolos habituales<br>
- <b>MVLESS</b> (diseñado para [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (protocolo basado en UDP, eliminado de xray-core en el cuarto trimestre de 2024 pero conservado en [Mahsa-Core](https://github.com/GFW-knocker/Xray-core))<br>
(nota: con mucho tráfico QUIC consume demasiada RAM)



## Instalar la última versión en un VPS Linux

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## Instalar v2.8.4-mahsa-r2 en un VPS Linux

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## Instalar una versión concreta

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## Instalar en Windows
descarga y ejecuta el binario .exe desde la sección release<br>
sigue el readme para abrir el panel en el navegador

## ¿Encontraste un error?
- <b>borra las cookies y los datos del sitio</b> en el navegador (a la izquierda de la barra de direcciones)
- reinstala el panel en el VPS en lugar de actualizarlo
