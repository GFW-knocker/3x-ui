# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) с ядром [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- Все обычные протоколы<br>
- <b>MVLESS</b> (разработан для [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (протокол на основе UDP, удалённый из xray-core в четвёртом квартале 2024 года, но сохранённый в [Mahsa-Core](https://github.com/GFW-knocker/Xray-core))<br>
(примечание: при высоком трафике QUIC потребляет слишком много оперативной памяти)
- поддержка <b>allowInsecure</b> (удалён из xray-core, но сохранён в [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)): переключатель в настройках TLS, передаётся в ссылках и подписках<br>
- <b>Приватность</b>: из соображений приватности убраны баннер спонсоров и лишние обращения к ip-api, определение публичного IP сделано необязательным; панель не обращается к сторонним доменам, кроме GitHub для обновлений и интеграций, которые вы включили сами<br>



## Установка последней версии на Linux VPS

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## Установка версии v2.8.4-mahsa-r2 на Linux VPS

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## Установка конкретной версии

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## Установка на Windows
скачайте и запустите .exe из раздела release<br>
чтобы открыть панель в браузере, следуйте readme

## Столкнулись с ошибкой?
- <b>очистите cookie и данные сайта</b> в браузере (слева от адресной строки)
- переустановите панель на VPS вместо обновления
