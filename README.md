## Behkad automation :D
این اسکریپت حضور و گزارش کارآموزی شما را به صورت خودکار ثبت میکند و از طریق یک ربات تلگرام که از قبل مشخص کردین نتیجه رو براتون میفرسته.
<br>
نحوه استفاده از برنامه هم خیلی ساده است کافیه به صورت کامند لاین چه در ویندوز و چه در لینوکس اجرا کرده و آپشن های مورد نیاز را به برنامه پاس بدید.

البته در گوشی های اندرویدی هم میتونید نسخه لینوکس را توسط برنامه ترموکس اجرا کنید.

**FOR SUPPORT PLEASE GIVE A STAR TO REPO :)**

## Usage
قبل از اجرا حتما فایل `conf.json` را با یک تکست ادیتور ویرایش کنید.

#### مقادیری که باید ویرایش کنید:
- latitude: عرض جغرافیایی
- longitude: طول جغرافیایی
- codeCollage: (نام کاربری)کد دانشجو
- codeMelli: کد ملی (رمزعبور)
- address: آدرس محل کارآموزی
- gozareshat: حداقل 10 تا گزارش داخل این لیست قرار بدین

#### اگر روی سرور آیپی خارج اجرا میکنین میتونین از تلگرام برای ارسال نتایج استفاده کنین:
- useBot: فعال یا غیرفعال بودن ارسال تلگرام
- ApiKey: توکن رباتی که توسط @botfather توی تلگرام ساختین
- telegramUserId: آیدی عددی تلگرام خودتون


### linux:
```bash
bash run_script.sh
```

### windows:
```bash
behkad_windows.exe
```

<hr>

## run in linux server with crontab(optional)

برای اینکه خودتون هر روز دستی این اسکریپت رو ران نکنین، میتونین با استفاده از یک سرور لینوکس این کارو کاملا اتومات کنین :)

با استفاده از این دستور وارد تنظیمات ابزار می‌شوید

```bash
crontab -e
```

اگر برای اولین بار این ابزار رو باز میکنین ازتون میپرسه با کدوم ادیتور بازش کنم
<br>
عدد 1 رو بزنین و وارد ادیتور بشین و در آخرین خط آن فایل این دستور رو قرار بدین:

```bash
0 7 * * * bash /path/to/behkad_hozor/run_script.sh
```

به جای `/path/to/behkad_hozor/script.sh` مسیری که پروژه را کلون کردین رو وارد کنین

و به ترتیب کلید های
<br>
`Ctrl + X`
<br>
`Y`
<br>
`Enter`
<br>
را وارد کنین

با این کار هر روز ساعت ۷ صبح اسکریپت اجرا شده و حضور رو ثبت خواهد کرد.