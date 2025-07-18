# Cache Service (`services/cache`)

Bu paket, uygulama genelinde kullanılabilen esnek bir önbellekleme (caching) katmanı sağlar. Temel amacı, sık erişilen verileri (API yanıtları, veritabanı sorgu sonuçları vb.) hafızada veya Redis gibi hızlı bir depoda tutarak, veritabanı yükünü azaltmak ve yanıt sürelerini iyileştirmektir.

## Temel Çalışma Prensibi: Değiştirilebilir Arka Plan (Swappable Backend)

Bu servisin en önemli özelliği, ortam değişkenlerine (`environment variables`) bağlı olarak farklı önbellekleme stratejileri arasında geçiş yapabilmesidir:

1.  **In-Memory Cache (Hafıza İçi Önbellek):**
    * **Ne Zaman Kullanılır?:** `REDIS_IS_ACTIVE` ortam değişkeni `true` olarak ayarlanmadığında varsayılan olarak bu modda çalışır.
    * **Nasıl Çalışır?:** Tüm önbellek verilerini uygulamanın kendi hafızasında (RAM) bir `map` içinde tutar. Herhangi bir dış bağımlılık gerektirmez, bu nedenle geliştirme ortamları veya küçük ölçekli uygulamalar için idealdir.
    * **Dezavantajı:** Uygulama yeniden başlatıldığında tüm önbellek silinir. Birden fazla sunucu (instance) varsa, her birinin kendi ayrı önbelleği olur.

2.  **Redis Cache:**
    * **Ne Zaman Kullanılır?:** `REDIS_IS_ACTIVE=true` olarak ayarlandığında ve `REDIS_ADDR`, `REDIS_PASSWORD` gibi bağlantı bilgileri sağlandığında bu modda çalışır.
    * **Nasıl Çalışır?:** Tüm önbellek verilerini harici bir Redis sunucusunda saklar.
    * **Avantajı:** Önbellek, uygulama yeniden başlasa bile kalıcıdır. Birden fazla sunucu aynı Redis veritabanını kullanarak paylaşımlı bir önbelleğe sahip olabilir, bu da veri tutarlılığını artırır. Üretim (production) ortamları için önerilen yöntemdir.

## Fonksiyonlar (Arayüz Metotları)

`CacheService` bir arayüzdür (`interface`), bu da hem `InMemoryCache` hem de `RedisCache`'in aynı metotlara sahip olduğu anlamına gelir.

---

### `TryCache`

Bir isteği yerine getirmeden önce, yanıtın önbellekte olup olmadığını kontrol eder.

-   **Ne Yapar?:** Verilen `group` ve `identifier` ile bir önbellek anahtarı (key) oluşturur (örn: `content:about-us`). Bu anahtarla önbellekte veri olup olmadığını kontrol eder. Eğer veri varsa, bu veriyi doğrudan HTTP yanıtı olarak yazar ve isteği sonlandırır.
-   **Ne Alır?:** `*gin.Context`, `string` (grup), `string` (tanımlayıcı)
-   **Ne Döndürür?:** `bool`. Eğer önbellekten yanıt verildiyse `true`, veri bulunamadıysa `false` döner.

```go
func (c *CacheService) TryCache(ctx *gin.Context, group, identifier string) bool
```

---

### `SaveCache` / `SaveCacheTTL`

Bir yanıtı önbelleğe kaydeder.

-   **Ne Yapar?:** Bir işlem sonucu elde edilen veriyi (genellikle bir `struct` veya `map`) JSON formatına çevirerek önbelleğe yazar. `SaveCache` varsayılan TTL (Time-To-Live, Yaşam Süresi) ile kaydederken, `SaveCacheTTL` özel bir yaşam süresi belirtme imkanı sunar.
-   **Ne Alır?:** `any` (kaydedilecek veri), `string` (grup), `string` (tanımlayıcı), opsiyonel `time.Duration` (TTL)
-   **Ne Döndürür?:** `error`.

```go
func (c *CacheService) SaveCache(response any, group, identifier string) error
```

---

### `ClearGroup` / `ClearAll`

Önbelleği temizler (invalidate eder).

-   **Ne Yapar?:**
    * `ClearGroup`: Belirli bir gruba ait tüm önbellek kayıtlarını siler (örn: `jobs` grubundaki tüm iş ilanları).
    * `ClearAll`: Tüm önbelleği tamamen boşaltır.
-   **Ne Alır?:** `string` (grup adı) veya hiçbir şey.

```go
func (c *CacheService) ClearGroup(group string)
```
