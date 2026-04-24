# Encanto — Kiro Specs

هذا المجلد يحتوي على جميع الـ specs والـ steering files الخاصة بإصلاح فجوات مشروع Encanto.

## كيفية الاستخدام
1. انسخ مجلد `.kiro/` بالكامل إلى root الـ repo
2. افتح المشروع في Kiro IDE
3. افتح الـ Specs panel وهتلاقي كل الـ specs جاهزة
4. ابدأ بـ `01-fix-build-tags` لأنه blocking لباقي الـ specs

## الأولويات

| # | Spec | النوع | الأولوية |
|---|---|---|---|
| 01 | `01-fix-build-tags` | Bugfix | 🔴 حرج أولاً |
| 02 | `02-workers-package` | Feature | 🔴 حرج |
| 06 | `06-websocket-security` | Bugfix | 🔴 أمان |
| 03 | `03-password-reset` | Feature | 🟠 مهم |
| 04 | `04-rate-limiting` | Feature | 🟠 مهم |
| 08 | `08-migration-conflict` | Bugfix | 🟠 مهم |
| 05 | `05-campaign-executor` | Feature | 🟡 متوسط |
| 07 | `07-qr-realtime` | Feature | 🟡 متوسط |

## Steering Files
- `product.md` — وصف المنتج وأهدافه
- `tech.md` — الـ tech stack والـ conventions
- `structure.md` — هيكل المشروع والقواعد
