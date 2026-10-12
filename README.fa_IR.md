# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) مجهز به [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- همه‌ی پروتکل‌های معمول<br>
- <b>MVLESS</b> (طراحی‌شده برای [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (پروتکل مبتنی بر UDP که در سه‌ماهه‌ی چهارم ۲۰۲۴ از xray-core حذف شد اما در [Mahsa-Core](https://github.com/GFW-knocker/Xray-core) باقی مانده است)<br>
(توجه: QUIC در ترافیک بالا حافظه‌ی RAM زیادی مصرف می‌کند)



## نصب آخرین نسخه روی VPS لینوکس

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## نصب نسخه‌ی v2.8.4-mahsa-r2 روی VPS لینوکس

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## نصب یک نسخه‌ی مشخص

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## نصب روی ویندوز
فایل اجرایی .exe را از بخش release دانلود و اجرا کنید<br>
برای باز کردن پنل در مرورگر، readme را دنبال کنید

## با مشکل مواجه شدید؟
- <b>کوکی‌ها و داده‌های سایت</b> را در مرورگر پاک کنید (در سمت چپ نوار آدرس)
- به‌جای به‌روزرسانی، پنل را روی VPS دوباره نصب کنید
