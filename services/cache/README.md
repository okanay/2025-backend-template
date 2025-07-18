# Cache Servisi (`services/cache`)

Bu paket, uygulama genelinde yüksek performanslı ve esnek bir önbellekleme (caching) katmanı sağlar. Temel amacı, sık erişilen verileri (veritabanı sorgu sonuçları, kullanıcı izinleri, ayarlar vb.) hafızada veya Redis gibi hızlı bir depoda tutarak veritabanı yükünü azaltmak ve API yanıt sürelerini dramatik şekilde iyileştirmektir.

## Temel Çalışma Prensibi

### 1. Değiştirilebilir Arka Plan (Swappable Backend)

Servis, ortam değişkenlerine (`environment variables`) bağlı olarak iki farklı modda çalışabilir:

#### In-Memory Cache (Hafıza İçi Önbellek)
- **Ne Zaman Kullanılır?** Varsayılan moddur. `REDIS_IS_ACTIVE=true` olarak ayarlanmadığında devreye girer.
- **Nasıl Çalışır?** Tüm verileri uygulamanın kendi hafızasında (RAM) tutar. Dış bağımlılık gerektirmediği için geliştirme ortamları için idealdir. Uygulama yeniden başladığında tüm önbellek silinir.

#### Redis Cache
- **Ne Zaman Kullanılır?** `REDIS_IS_ACTIVE=true` olarak ayarlandığında ve gerekli Redis bağlantı bilgileri sağlandığında çalışır.
- **Nasıl Çalışır?** Tüm verileri harici bir Redis sunucusunda saklar. Önbellek kalıcıdır ve birden fazla sunucu (instance) arasında paylaşılabilir. Üretim (production) ortamları için şiddetle tavsiye edilen yöntemdir.

### 2. "Cache-Aside" Deseni ve `GetOrSet` Metodu

Servisin en güçlü yeteneği, "Cache-Aside" desenini uygulayan `GetOrSet` metodudur. Bu metot, tekrar eden "cache'de var mı, yoksa veritabanından al" mantığını tamamen soyutlar.

**Nasıl Çalışır?**
1. Önce verinin cache'de olup olmadığını kontrol eder.
2. **Cache Hit (Başarılı):** Veri cache'de varsa, bunu doğrudan hedefe yazar ve işlemi bitirir.
3. **Cache Miss (Başarısız):** Veri cache'de yoksa veya bozuksa, parametre olarak aldığı `FallbackFunc`'ı (genellikle bir veritabanı sorgusu) çalıştırır.
4. Fallback'ten gelen taze veriyi hem hedefe yazar hem de bir sonraki istek için **arka planda cache'e kaydeder.**

## Fonksiyonlar ve Kullanım

### `GetOrSet`

Bir veriyi cache'den getirmeye çalışır. Bulamazsa, `fallback` fonksiyonunu çalıştırır, sonucu hem hedefe yazar hem de cache'e kaydeder. Bu, uygulamanızda kullanacağınız birincil cache metodudur.

**Parametreler:**
- `group` (string): Cache grubu
- `identifier` (string): Benzersiz anahtar
- `dest` (pointer): Sonucun yazılacağı hedef
- `fallback` (function): Veri bulunamazsa çalışacak fonksiyon

**Dönüş Değeri:** `error`

#### Örnek Kullanım:

```go
// Bir kullanıcının izinlerini cache'den veya veritabanından getirme
var userPermissions []types.Permission

// Cache'de veri yoksa çalışacak olan veritabanı sorgusu
dbFallback := func() (any, error) {
    return authRepo.SelectPermissionsByUserID(ctx, userID)
}

// GetOrSet ile tüm cache mantığını tek satırda hallet
err := cacheService.GetOrSet(
    "permissions",      // Cache Grubu
    userID.String(),    // Benzersiz Anahtar
    &userPermissions,   // Sonucun yazılacağı hedef
    dbFallback,         // Fallback Fonksiyonu
)

if err != nil {
    // Hata yönetimi...
}

// Bu noktadan sonra 'userPermissions' dolu ve kullanıma hazır.
```

### `TryCache`

**Sadece tam bir HTTP yanıtını** cache'lemek için kullanılır. Veriyi bulursa, yanıtı doğrudan `gin.Context`'e yazar ve isteği sonlandırır.

**Parametreler:**
- `*gin.Context`: Gin context
- `group` (string): Cache grubu
- `identifier` (string): Benzersiz anahtar

**Dönüş Değeri:** `bool` (cache'den yanıt verildiyse `true`)

#### Örnek Kullanım:

```go
// HTTP yanıtını cache'den sunmaya çalış
if cacheService.TryCache(c, "api_responses", "user_profile_"+userID) {
    return // Cache'den yanıt verildi, işlem tamamlandı
}

// Cache'de yoksa normal işlemi devam ettir
// ... API logic ...
```

### `SaveCache` / `SaveCacheTTL`

Bir veriyi manuel olarak cache'e kaydetmek için kullanılır. `GetOrSet` bu işlemi otomatik yaptığı için nadiren ihtiyaç duyulur.

- **`SaveCache`**: Varsayılan TTL (ömür) ile kaydeder.
- **`SaveCacheTTL`**: Özel bir TTL ile kaydeder.

#### Örnek Kullanım:

```go
// Varsayılan TTL ile kaydet
err := cacheService.SaveCache("users", userID.String(), userData)

// Özel TTL ile kaydet (1 saat)
err := cacheService.SaveCacheTTL("temp_data", sessionID, tempData, time.Hour)
```

### `ClearGroup` / `ClearAll`

Cache'i geçersiz kılmak (invalidate) için kullanılır.

- **`ClearGroup`**: Belirli bir gruba ait tüm kayıtları siler (örn: bir kullanıcının izinleri güncellendiğinde `"permissions"` grubunu temizlemek).
- **`ClearAll`**: Tüm cache'i tamamen boşaltır.

#### Örnek Kullanım:

```go
// Belirli bir grubu temizle
err := cacheService.ClearGroup("permissions")

// Tüm cache'i temizle
err := cacheService.ClearAll()
```

## Cache Grupları ve Organizasyon

Cache'i organize etmek için grup sistemi kullanılır:

```go
// Farklı veri tiplerini farklı gruplarda organize etme
cacheService.GetOrSet("users", userID.String(), &user, getUserFromDB)
cacheService.GetOrSet("permissions", userID.String(), &perms, getPermsFromDB)
cacheService.GetOrSet("settings", "app_config", &config, getConfigFromDB)
```

## Yapılandırma

### Ortam Değişkenleri

```bash
# Redis kullanımını aktifleştir
REDIS_IS_ACTIVE=true

# Redis bağlantı bilgileri
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password
REDIS_DB=0
```

### Geliştirme vs Üretim

```go
// Geliştirme ortamında
REDIS_IS_ACTIVE=false  // In-memory cache kullanır

// Üretim ortamında
REDIS_IS_ACTIVE=true   // Redis cache kullanır
```

## Performans Optimizasyonu

### Cache Hit Oranını Artırma

```go
// Sık kullanılan verileri cache'le
cacheService.GetOrSet("popular_posts", "trending", &posts, getPopularPosts)

// Uzun süreli veriler için daha uzun TTL
cacheService.SaveCacheTTL("static_content", "about_page", content, 24*time.Hour)
```

### Cache Invalidation Stratejileri

```go
// Veri güncellendiğinde ilgili cache'i temizle
func UpdateUserPermissions(userID uuid.UUID, newPermissions []Permission) error {
    // Önce veritabanını güncelle
    err := repo.UpdatePermissions(userID, newPermissions)
    if err != nil {
        return err
    }

    // Sonra cache'i temizle
    return cacheService.ClearGroup("permissions")
}
```
