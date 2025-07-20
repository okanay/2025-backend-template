# Go & Gin Backend Şablonu (2025)

Bu proje, yüksek performanslı **Gin** web çatısı üzerine inşa edilmiş, modern ve zengin özelliklere sahip bir Go backend başlangıç şablonudur. Amacı, her yeni projede ihtiyaç duyulan temel sistemleri (kimlik doğrulama, cache, dosya yükleme, arka plan görevleri vb.) en iyi pratiklerle standart hale getirerek geliştirme sürecini dramatik bir şekilde hızlandırmaktır.

## Çekirdek Motor: Gin Web Framework

Proje, gücünü ve hızını Go ekosisteminin en popüler ve sevilen web çatılarından biri olan [Gin](https://gin-gonic.com/)'den alır. Gin sayesinde, minimum eforla son derece performanslı, ölçeklenebilir ve yönetimi kolay API'lar geliştirebilirsiniz. Bu şablon, Gin'in esnek middleware altyapısını sonuna kadar kullanır.

## ✨ Şablonun Yetenekleri

Bu şablon, bir backend uygulamasında ihtiyaç duyabileceğiniz birçok kabiliyeti kutudan çıktığı gibi sunar:

### 1. Gelişmiş Kimlik Doğrulama ve Yetkilendirme

Sadece "giriş yapma" değil, tam kapsamlı ve güvenli bir oturum yönetimi altyapısı sunar.

* **JWT & Güvenli Cookie'ler:** Oturumlar, kısa ömürlü `access_token` ve uzun ömürlü `refresh_token` mekanizması ile yönetilir. Token'lar, XSS saldırılarına karşı korumalı `HttpOnly` ve `Secure` cookie'ler içinde saklanır.
* **Sosyal Medya ile Giriş (OAuth):** Google, Apple gibi popüler sağlayıcılar üzerinden kolayca giriş yapma altyapısı `goth` kütüphanesi ile entegre edilmiştir.
* **Detaylı İzin Sistemi (Permissions):** Standart `Admin`, `User` gibi rollerin ötesinde, "dosya silebilir", "blog yayınlayabilir" gibi çok daha granüler yetkileri (`file:delete`, `blog:publish`) kullanıcılara atamanızı sağlayan güçlü bir `PermissionMiddleware` içerir.

### 2. Yüksek Performanslı ve Değiştirilebilir Cache

Uygulamanızın hızını ve veritabanı yükünü dramatik şekilde optimize eden, ortamınıza göre seçebileceğiniz bir cache katmanı sunar.

* **Redis Desteği:** Üretim (production) ortamları için ideal, dağıtık ve kalıcı bir cache altyapısı.
* **In-Memory Cache Desteği:** Geliştirme (development) ortamları için harici bir bağımlılık gerektirmeyen, basit ve hızlı bir hafıza içi cache.
* **Akıllı `GetOrSet` Deseni:** "Veri cache'de var mı? Yoksa veritabanından al, cache'e yaz ve sonra döndür" şeklindeki tekrar eden mantığı tek bir fonksiyonda toplayarak kod tekrarını önler.

### 3. Asenkron ve Zamanlanmış Görevler (Automation)

Uygulamanızın arka planda veya gelecekte bir zamanda işler yapmasını sağlayan güçlü bir otomasyon motoru içerir.

* **Periyodik Görevler (Cron Jobs):** "Her gece 03:00'te veritabanını yedekle" veya "her saat başı eski logları temizle" gibi rutin sistem işlerini `Add` fonksiyonu ile kolayca tanımlayın.
* **Tek Seferlik Planlanmış Görevler:** "Bu blog yazısını Cuma günü saat 14:30'da yayınla" gibi kullanıcıya özel, dinamik görevleri `Schedule` fonksiyonu ile planlayın.
* **Güvenli ve Yönetilebilir:** Tüm görevler, aynı anda iki kez çalışmalarını önleyen bir kilitleme mekanizmasına sahiptir ve `Trigger` komutuyla manuel olarak tetiklenebilirler.

### 4. Modern ve Güvenli Dosya Yükleme (Cloudflare R2 Entegrasyonu)

Büyük dosyaların backend sunucunuzu yormasını engelleyen, modern ve ölçeklenebilir bir dosya yükleme mimarisi sunar.

1.  **Güvenli URL Talebi:** Frontend, backend'den dosyayı yüklemek için süresi kısıtlı ve güvenli bir **"Presigned URL"** talep eder.
2.  **Doğrudan Yükleme:** Frontend, aldığı bu URL ile dosyayı sunucunuzu **bypass ederek** doğrudan Cloudflare R2'ye yükler. Bu, sunucu kaynaklarınızı korur ve performansı artırır.
3.  **Onaylama:** Yükleme bitince frontend backend'e haber verir ve dosya metadatası veritabanına kaydedilir.

### 5. Zengin Middleware Katmanı

Uygulamanızın güvenliğini ve kararlılığını artıran, Gin ile tam uyumlu, kullanıma hazır bir middleware koleksiyonu içerir.

* **Rate Limiter:** Kötü niyetli botlara ve brute-force saldırılarına karşı IP bazlı istek sınırlaması.
* **Timeout:** Bir isteğin sunucuyu çok uzun süre meşgul etmesini önleyen zaman aşımı kontrolü.
* **CORS & Güvenlik Başlıkları:** Tarayıcılar için Cross-Origin Resource Sharing ve diğer temel güvenlik başlıklarını (`XSS-Protection`, `Frame-Options` vb.) yönetir.
* **Captcha Doğrulaması:** Cloudflare Turnstile ve Google reCAPTCHA entegrasyonları ile bot koruması.

## 🚀 Başlarken

1.  **`.env` Dosyasını Oluşturun:** `.env.example` dosyasını kopyalayarak `.env` adında yeni bir dosya oluşturun ve içindeki değerleri kendi yapılandırmanıza göre doldurun.
2.  **Veritabanı Migration'larını Çalıştırın:**
    ```bash
    go run ./cmd/migrate up_all
    ```
3.  **Uygulamayı Başlatın:**
    ```bash
    go run main.go
    ```
