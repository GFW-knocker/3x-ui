# MX-UI
[3X-UI](https://github.com/MHSanaei/3x-ui) مزوّد بـ [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)<br><br>
- جميع البروتوكولات المعتادة<br>
- <b>MVLESS</b> (مصمَّم لـ [MahsaNG](https://github.com/GFW-knocker/MahsaNG))<br>
- <b>QUIC</b> (بروتوكول يعتمد على UDP أُزيل من xray-core في الربع الأخير من 2024 لكنه ما زال موجودًا في [Mahsa-Core](https://github.com/GFW-knocker/Xray-core))<br>
(ملاحظة: يستهلك QUIC ذاكرة RAM كبيرة عند ارتفاع حركة المرور)
- دعم <b>allowInsecure</b> (أُزيل من xray-core لكنه محفوظ في [Mahsa-Core](https://github.com/GFW-knocker/Xray-core)): مفتاح في إعدادات TLS، ويُصدَّر في روابط المشاركة والاشتراكات<br>
- <b>الخصوصية</b>: أُزيل بانر الرعاة والاستدعاءات الزائدة لـ ip-api حفاظًا على الخصوصية، وأصبح البحث عن الـ IP العام اختياريًا؛ لا تتصل اللوحة بأي نطاق خارجي سوى GitHub للتحديثات والتكاملات التي تفعّلها بنفسك<br>



## تثبيت أحدث إصدار على VPS لينكس

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/main/install.sh)
```

## تثبيت الإصدار v2.8.4-mahsa-r2 على VPS لينكس

```bash
bash <(curl -Ls https://raw.githubusercontent.com/GFW-knocker/3x-ui/v2.8.4-mahsa-r2/install.sh) v2.8.4-mahsa-r2
```

## تثبيت إصدار محدد

```bash
VERSION=v2.8.4-mahsa-r2 && bash <(curl -Ls "https://raw.githubusercontent.com/GFW-knocker/3x-ui/$VERSION/install.sh") $VERSION
```

## التثبيت على ويندوز
نزّل ملف .exe من قسم release وشغّله<br>
اتبع ملف readme لفتح اللوحة في المتصفح

## واجهت مشكلة؟
- <b>امسح ملفات تعريف الارتباط وبيانات الموقع</b> من المتصفح (على يسار شريط العنوان)
- أعد تثبيت اللوحة على الـ VPS بدلًا من الترقية
