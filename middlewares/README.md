# Middlewares (/middlewares)

Bu dizin, projenin Gin web çatısı üzerine kurulu olan ve gelen HTTP isteklerini işleyen ara katman yazılımlarını (middlewares) içerir. Middleware'ler, ana handler fonksiyonu çalışmadan önce veya sonra belirli görevleri yerine getirmek için bir zincir (chain) halinde çalışır.

## Middleware Zinciri ve Çalışma Sırası

`main.go` dosyasında tanımlanan middleware'ler, eklendikleri sırayla çalışır. Bu sıra, güvenlik ve performans için kritik öneme sahiptir. Projemizdeki genel sıralama şöyledir:

### Global Middleware'ler (Tüm isteklerde çalışır)
- **TimeoutMiddleware**: İsteğin tamamı için bir zaman aşımı belirler.
- **SecureConfig**: Güvenlik başlıkları (headers) ekler.
- **CorsConfig**: Cross-Origin Resource Sharing kurallarını belirler.
- **RateLimiterMiddleware**: IP bazlı istek limiti uygular.

### Protected Route Middleware'leri (Kimlik doğrulaması gerektiren rotalarda çalışır)
- **AuthMiddleware**: Kullanıcının kimliğini doğrular.
- **PermissionMiddleware**: Kimliği doğrulanmış kullanıcının belirli bir işlemi yapma yetkisini kontrol eder.

## Middleware'ler ve Görevleri

### TimeoutMiddleware

- **Dosya**: `timeout.go`
- **Amaç**: Sunucuya gelen her isteğin, `configs/constants.go` içinde tanımlanan `REQUEST_MAX_DURATION` süresini aşmamasını garanti eder. Eğer bir istek bu sürede tamamlanmazsa, 408 Request Timeout hatası ile sonlandırılır. Bu, sunucunun yavaş veya takılmış istekler tarafından meşgul edilmesini önler.

### RateLimiterMiddleware

- **Dosya**: `rate-limit.go`
- **Amaç**: Kötü niyetli veya hatalı çalışan istemcilerin sunucuya çok kısa sürede aşırı sayıda istek göndermesini (Brute Force, DDoS saldırıları) engeller.
- **Nasıl Çalışır?**: Her bir istemci IP adresi için belirli bir zaman aralığında yapılabilecek maksimum istek sayısını takip eder. Limit aşıldığında 429 Too Many Requests hatası döndürür.

### AuthMiddleware

- **Dosya**: `auth.go`
- **Amaç**: Korunmuş rotalara erişmeye çalışan kullanıcının kimliğini doğrulamak. Bu, projenin en kritik güvenlik katmanlarından biridir.
- **Nasıl Çalışır?**:
  - İstekle birlikte gelen `access_token` cookie'sini kontrol eder.
  - **Access Token Geçerliyse**: Token içindeki kullanıcı bilgilerini (`user_id`, `user_role`) ayıklar, Gin context'ine ekler ve isteğin devam etmesine izin verir.
  - **Access Token Geçersiz veya Süresi Dolmuşsa**: Bu kez `refresh_token` cookie'sini kontrol eder.
  - **Refresh Token Geçerliyse**: Veritabanından bu token'ın geçerliliğini doğrular, kullanıcı için yeni bir `access_token` üretir, cookie'leri günceller ve isteğin devam etmesine izin verir.
  - **Her İki Token da Geçersizse**: 401 Unauthorized hatası döndürerek isteği sonlandırır.

### PermissionMiddleware

- **Dosya**: `permission.go`
- **Amaç**: AuthMiddleware tarafından kimliği doğrulanmış bir kullanıcının, erişmeye çalıştığı rota için gerekli izne sahip olup olmadığını kontrol eder. Rol bazlı kontrolden (RequireRole) daha granüler ve esnek bir yetkilendirme sağlar.
- **Nasıl Çalışır?**:
  - Admin rolündeki kullanıcıları her zaman yetkili kabul eder.
  - Gelen isteğin `METHOD:PATH` kombinasyonunu (örn: `DELETE:/v1/files/:id`), `PermissionMap` adındaki merkezi haritada arar.
  - Eğer rota için bir izin gerekiyorsa, kullanıcının izinlerini `CacheService.GetOrSet` metodunu kullanarak cache'den veya veritabanından getirir.
  - Kullanıcının izin listesinde gerekli izin varsa isteğe devam eder, yoksa 403 Forbidden hatası döndürür.

### RequireRole

- **Dosya**: `require-role.go`
- **Amaç**: Belirli bir rotaya sadece belirli bir role (örn: Admin) sahip kullanıcıların erişebilmesini sağlar. PermissionMiddleware'e göre daha basit bir yetkilendirme yöntemidir.

### Captcha Middleware'leri

- **Dosyalar**: `turnstile-captcha.go`, `recaptcha.go`
- **Amaç**: Kayıt olma, giriş yapma gibi kritik form işlemlerini bot saldırılarından korumak.
- **Nasıl Çalışır?**: Frontend'den gelen captcha token'ını alır ve ilgili servise (Cloudflare veya Google) göndererek doğrulamasını yapar. Doğrulama başarısız olursa isteği 403 Forbidden hatası ile engeller.

## Middleware Kullanım Örnekleri

### Global Middleware Tanımlama
```go
// main.go içinde
router.Use(TimeoutMiddleware())
router.Use(SecureConfig())
router.Use(CorsConfig())
router.Use(RateLimiterMiddleware())
```

### Protected Route Middleware Tanımlama
```go
// Kimlik doğrulaması gerektiren route grubu
protected := router.Group("/api/v1")
protected.Use(AuthMiddleware())
protected.Use(PermissionMiddleware())
```

### Rol Bazlı Middleware Tanımlama
```go
// Sadece admin kullanıcıların erişebileceği rotalar
adminOnly := router.Group("/admin")
adminOnly.Use(AuthMiddleware())
adminOnly.Use(RequireRole("admin"))
```
