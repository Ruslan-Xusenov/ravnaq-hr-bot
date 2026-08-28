# HR Telegram Bot & Admin API

Ushbu loyiha ko'p tilli (Uzbek, Russian, English) HR Telegram boti va Admin REST APIning Go (Golang) da yozilgan backend qismidir.

## Asosiy Xususiyatlar (Features)
- **Foydalanuvchini ro'yxatga olish:** Til tanlash, telefon raqam orqali tasdiqlash.
- **Rezyume shakllantirish:** Bot ichida bosqichma-bosqich rezyume yaratish (Ism, manzili, ish haqi talabi).
- **Vakansiyalar:** Faol vakansiyalarni ko'rish va ularga ariza (apply) yuborish.
- **Admin REST API:** JWT orqali himoyalangan API. Vakansiyalar qo'shish, o'chirish, tahrirlash va arizalarni tasdiqlash/rad etish.
- **Ommaviy xabarlar (Broadcast):** Asynq + Redis orqali fon rejimida (background task) limitlarga tushmasdan foydalanuvchilarga xabar yuborish.

## Texnologiyalar
- **Til:** Go (Golang 1.21+)
- **Ma'lumotlar bazasi:** PostgreSQL 15 (pgxpool)
- **Kesh va Navbat (Queue):** Redis 7 (go-redis, hibiken/asynq)
- **Telegram:** telebot.v3
- **HTTP Server:** Gin (gin-gonic)
- **Migratsiya:** Goose
- **Deployment:** Docker & Docker Compose

## Mahalliy kompyuterda ishga tushirish (Local Setup)

1. **Repozitoriyni yuklab olish va .env faylini yaratish:**
   ```bash
   cp .env.example .env
   ```
   `.env` faylidagi `TELEGRAM_BOT_TOKEN` ni o'z bot tokeningizga almashtiring (BotFather orqali olinadi).

2. **Docker Compose orqali ishga tushirish (tavsiya etiladi):**
   ```bash
   docker-compose up -d
   ```
   *Eslatma: Docker compose barcha bazalarni (Postgres, Redis) ko'taradi, migratsiyalarni (migrate konteyneri) bajaradi va ilovani ishga tushiradi.*

3. **Oddiy (Docker'siz) ishga tushirish:**
   - PostgreSQL va Redis serverlarini yoqing.
   - Migratsiyalarni o'rnating: `make migrate-up`
   - Ilovani ishga tushiring: `make run`

## Admin API va Autentifikatsiya

Admin panel orqali ishlash uchun sizda JWT token bo'lishi kerak. Tizimda birlamchi super-admin avtomatik yaratiladi:
- **Email:** `admin@example.com`
- **Parol:** `password`

### API Endpointlar qisqacha
- `POST /api/v1/auth/login` - Tizimga kirish va token olish.
- `GET, POST, PUT, DELETE /api/v1/vacancies` - Vakansiyalarni boshqarish.
- `GET, PATCH /api/v1/applications` - Arizalarni ko'rish va maqomini o'zgartirish.
- `POST /api/v1/broadcast` - Ommaviy xabarlar yuborish.

> *Endpointlar faqat "Bearer <token>" yuborilgandagina ishlaydi.*

## Papkalar Strukturasi
- `cmd/server/main.go` - Ilovaning asosiy ishga tushish nuqtasi.
- `internal/delivery/telegram` - Telegram bot buyruqlari va mantiqi.
- `internal/delivery/http` - Gin HTTP serveri, Middleware va Controllerlar.
- `internal/domain` - Har bir biznes obyekt uchun Model va DB Repository interfeyslari.
- `internal/worker` - Asynq orqali ishlovchi fon vazifalari (Broadcast).
- `migrations` - Ma'lumotlar bazasini shakllantiruvchi SQL fayllar.
