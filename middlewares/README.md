# Authentication & Authorization Middleware

Bu klasör, uygulamanın kimlik doğrulama (authentication) ve yetkilendirme (authorization) katmanlarını içerir. Her middleware'in belirli bir sorumluluğu vardır ve birlikte güvenli bir sistem oluştururlar.

## 📋 İçindekiler

- [AuthMiddleware](#authmiddleware)
- [OptionalAuthMiddleware](#optionalauthmiddleware)
- [RequireRoleMiddleware](#requirerolemiddleware)
- [PermissionMiddleware](#permissionmiddleware)
- [RateLimitMiddleware](#ratelimitmiddleware)
- [TurnstileCaptchaMiddleware](#turnstilecaptchamiddleware)
- [TimeoutMiddleware](#timeoutmiddleware)

---

## AuthMiddleware

**Ana kimlik doğrulama katmanı** - Korumalı endpoint'lere sadece kimliği doğrulanmış kullanıcıların erişmesini sağlar.

### 🎯 Amaç
Sadece geçerli oturuma sahip kullanıcıların korumalı rotalara erişebilmesini garantiler.

### ⚙️ Çalışma Mantığı

1. **Access Token Kontrolü**
   - Cookie'lerde `access_token` arar
   - Token geçerliyse → kullanıcı bilgilerini context'e ekler

2. **Refresh Token Fallback**
   - Access token yoksa/geçersizse → `refresh_token` kontrol eder
   - Refresh token geçerliyse → yeni access token oluşturur

3. **Yetkisiz Erişim**
   - Her iki token da geçersizse → `401 Unauthorized` döndürür

### 🚫 Önemli Not
"Ya hep ya hiç" prensibiyle çalışır - misafir erişimine izin vermez.

---

## OptionalAuthMiddleware

**Esnek kimlik doğrulama** - Hem misafir hem de üye kullanıcıların erişebildiği rotalar için.

### 🎯 Amaç
Kullanıcı durumuna göre farklı içerik sunulabilecek endpoint'ler için kimlik tespiti.

### ⚙️ Çalışma Mantığı

- **Başarılı doğrulama:** `is_authenticated: true` + kullanıcı bilgileri
- **Başarısız doğrulama:** `is_authenticated: false`
- **Her durumda:** İstek devam eder, handler içeriği belirler

### 💡 Kullanım Örneği
Blog yazısının misafire özet, üyeye tam metin gösterilmesi.

---

## RequireRoleMiddleware

**Rol tabanlı yetkilendirme** - Belirli rollere sahip kullanıcıları filtreler.

### 🎯 Amaç
Rotaları sadece belirli kullanıcı rollerine (Admin, Moderator vb.) açmak.

### ⚙️ Çalışma Mantığı

1. Context'ten `user_role` okur (AuthMiddleware'den gelir)
2. Rolün izin verilen roller listesinde olup olmadığını kontrol eder
3. Uygun değilse → `403 Forbidden` döndürür

### 📌 Gereksinim
AuthMiddleware'den **sonra** çalışmalıdır.

---

## PermissionMiddleware

**Granüler yetkilendirme** - Spesifik eylem izinlerini kontrol eder.

### 🎯 Amaç
Rol bağımsız, detaylı izin kontrolü (`files:upload`, `posts:delete` gibi).

### ⚙️ Çalışma Mantığı

1. Context'ten `user_id` okur
2. **Cache kontrolü:** Redis'te kullanıcı izinlerini arar
3. **Database fallback:** Cache'te yoksa veritabanından çeker
4. **İzin kontrolü:** Gerekli iznin varlığını doğrular
5. İzin yoksa → `403 Forbidden` döndürür

### 🚀 Performans
Redis cache kullanarak veritabanı yükünü azaltır.

---

## RateLimitMiddleware

**Hız sınırlama** - DoS saldırılarını ve kötüye kullanımı önler.

### 🎯 Amaç
İstemci başına belirli zaman aralığında maksimum istek sayısı sınırı.

### ⚙️ Çalışma Mantığı

1. **IP tespiti:** İsteği yapan gerçek IP adresini bulur
2. **Sayaç kontrolü:** Redis'te IP bazlı istek sayacını kontrol eder
3. **Limit kontrolü:** Belirlenen sınır (örn: 60 istek/dakika) aşılmış mı?
4. Limit aşıldıysa → `429 Too Many Requests` döndürür

### ⚡ Önlem
Sunucu kaynaklarını korur ve servis kalitesini garanti eder.

---

## TurnstileCaptchaMiddleware

**Bot koruması** - Cloudflare Turnstile ile insan/bot ayrımı.

### 🎯 Amaç
Otomatik botları ve spam'i engelleyerek form güvenliği sağlar.

### ⚙️ Çalışma Mantığı

1. **Token okuma:** Request body'den `cf-turnstile-response` token'ını alır
2. **Cloudflare doğrulama:** Token'ı secret key ile Cloudflare API'sine gönderir
3. **Sonuç değerlendirme:**
   - Başarılı → İstek devam eder
   - Başarısız → `403 Forbidden` döndürür

### 🔒 Kullanım Alanları
Kayıt, giriş, iletişim formları gibi kritik işlemler.

---

## TimeoutMiddleware

**Zaman aşımı koruması** - Uzun süren istekleri sonlandırır.

### 🎯 Amaç
Takılıp kalan veya çok yavaş isteklerin sunucu kaynaklarını tüketmesini engeller.

### ⚙️ Çalışma Mantığı

1. **Timeout tanımlama:** Her istek için context'e zaman sınırı koyar
2. **Süre kontrolü:** Handler'ın işlem süresi takip edilir
3. **Zaman aşımı:** Belirlenen süre aşılırsa context iptal edilir
4. **Hata yanıtı:** `503 Service Unavailable` döndürür

### ⏱️ Sonuç
Sunucu stabilitesi ve kaynak yönetimi sağlanır.

---

## 🔗 Middleware Zinciri Örneği

```
Request → TimeoutMiddleware → RateLimitMiddleware → AuthMiddleware → RequireRoleMiddleware → PermissionMiddleware → Handler
```

## 📝 Notlar

- Middleware'ler belirli bir sırayla çalışmalıdır
- Her middleware bir sonraki katmana geçmeden önce kendi kontrollerini yapar
- Hata durumunda istek zinciri kesilir ve uygun HTTP status kodu döndürülür
- Cache kullanımı (Redis) performans için kritiktir
