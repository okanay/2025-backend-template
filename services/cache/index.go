// cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8" // veya github.com/redis/go-redis/v8
)

// Cache grupları - gerektiğinde ekleyebilirsiniz
const (
	PermissionCacheGroup = "permissions"
)

type FallbackFunc func() (any, error)

// CacheService, tüm cache implementasyonları için ortak arayüz
type CacheService interface {
	GetOrSet(group, identifier string, dest any, fallback FallbackFunc) error
	TryCache(ctx *gin.Context, group, identifier string) bool
	SaveCache(response any, group, identifier string) error
	SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error
	ClearGroup(group string)
	ClearAll()
	Stop()
}

// NewCacheService, ortam değişkenlerine göre uygun cache servisini döndürür
func NewCacheService(defaultTTL time.Duration) CacheService {
	// Ortam değişkenlerini başta oku
	useRedis := os.Getenv("REDIS_IS_ACTIVE")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")

	// Varsayılan değerler ata
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 0
	if redisDBStr != "" {
		fmt.Sscanf(redisDBStr, "%d", &redisDB)
	}

	// Redis kullanılacak mı?
	if useRedis == "true" {
		fmt.Println("🚀 [REDIS CACHE] : Starting Redis cache backend")
		fmt.Printf("🔗 Redis address : %s\n", redisAddr)
		return NewRedisCache(redisAddr, redisPassword, redisDB, defaultTTL)
	}

	fmt.Println("💾 [MEMORY CACHE] : Starting in-memory cache backend")
	return NewInMemoryCache(defaultTTL)
}

// ===== IN-MEMORY CACHE IMPLEMENTATION =====

// InMemoryCache in-memory önbellekleme için yapı
type InMemoryCache struct {
	mu   sync.RWMutex
	data map[string]cacheItem
	ttl  time.Duration
}

// cacheItem önbellekteki bir veriyi ve metadata'sını temsil eder
type cacheItem struct {
	value    []byte
	cachedAt time.Time
	ttl      time.Duration // Opsiyonel TTL
}

// NewInMemoryCache yeni bir in-memory cache instance'ı oluşturur
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		data: make(map[string]cacheItem),
		ttl:  ttl,
	}

	return cache
}

func (c *InMemoryCache) GetOrSet(group, identifier string, dest any, fallback FallbackFunc) error {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// 1. Cache'den veriyi almayı dene.
	cachedValue, found := c.get(cacheKey)
	if found {
		// 2. Cache'de bulundu. Veriyi hedefe (dest) Unmarshal etmeyi dene.
		// Burada 'err' değişkenini if bloğunun dışında da kullanabilmek için önce tanımlıyoruz.
		var unmarshalErr error
		if unmarshalErr = json.Unmarshal(cachedValue, dest); unmarshalErr == nil {
			// Başarıyla çözümlendi! İşlem tamam, fonksiyondan çık.
			return nil
		}

		// Eğer buraya ulaştıysak, Unmarshal işlemi başarısız olmuştur.
		// Veri bozuk kabul edilir, loglanır ve taze veri için fallback'e devam edilir.
		log.Printf("CACHE_CORRUPT: Key %s. Re-fetching. Error on unmarshal: %v", cacheKey, unmarshalErr)
	}

	// 3. Cache'de yok ya da veri bozuk. Fallback fonksiyonunu çalıştırarak taze veriyi al.
	fallbackData, err := fallback()
	if err != nil {
		return err // Veritabanı veya ana kaynaktan veri alınamadı.
	}

	// 4. Fallback'ten gelen taze veriyi JSON'a çevir.
	bytes, err := json.Marshal(fallbackData)
	if err != nil {
		return fmt.Errorf("fallback data could not be marshalled to JSON: %w", err)
	}

	// 5. Bu taze veriyi (goroutine ile) arka planda cache'e kaydet.
	go func() {
		if saveErr := c.SaveCache(fallbackData, group, identifier); saveErr != nil {
			log.Printf("CACHE_SAVE_ERROR: Key %s. Error: %v", cacheKey, saveErr)
		}
	}()

	// 6. Son olarak, taze veriyi hedefe (dest) Unmarshal et.
	return json.Unmarshal(bytes, dest)
}

// TryCache önbellekteki veriyi kontrol eder ve varsa yanıt olarak döndürür
func (c *InMemoryCache) TryCache(ctx *gin.Context, group, identifier string) bool {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.RLock()
	item, exists := c.data[cacheKey]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	// TTL kontrolü
	now := time.Now()

	// Özel TTL kontrolü
	if item.ttl > 0 && now.Sub(item.cachedAt) > item.ttl {
		c.mu.Lock()
		delete(c.data, cacheKey)
		c.mu.Unlock()
		return false
	}

	// Genel TTL kontrolü
	if now.Sub(item.cachedAt) > c.ttl {
		c.mu.Lock()
		delete(c.data, cacheKey)
		c.mu.Unlock()
		return false
	}

	// Cache hit - önbellekteki veriyi dön
	ctx.Data(http.StatusOK, "application/json", item.value)
	ctx.Header("X-Cache", "HIT")
	return true
}

// SaveCache yanıtı önbelleğe alır
func (c *InMemoryCache) SaveCache(response any, group, identifier string) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.Lock()
	c.data[cacheKey] = cacheItem{
		value:    jsonData,
		cachedAt: time.Now(),
	}
	c.mu.Unlock()

	return nil
}

// SaveCacheTTL özel TTL ile yanıtı önbelleğe alır
func (c *InMemoryCache) SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.Lock()
	c.data[cacheKey] = cacheItem{
		value:    jsonData,
		cachedAt: time.Now(),
		ttl:      ttl,
	}
	c.mu.Unlock()

	return nil
}

func (c *InMemoryCache) Delete(group, identifier string) error {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, cacheKey)
	return nil
}

// ClearGroup bir grubu önbellekten temizler
func (c *InMemoryCache) ClearGroup(group string) {
	prefix := group + ":"

	c.mu.Lock()
	defer c.mu.Unlock()

	// Belirli bir önekle başlayan tüm anahtarları temizle
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ClearAll tüm önbelleği temizler
func (c *InMemoryCache) ClearAll() {
	c.mu.Lock()
	c.data = make(map[string]cacheItem)
	c.mu.Unlock()
}

// Stop temizleme goroutine'ini durdurur
func (c *InMemoryCache) Stop() {}

// Cleanup, süresi dolmuş tüm öğeleri önbellekten temizler.
func (c *InMemoryCache) Cleanup() {
	now := time.Now()
	expiredKeys := []string{}

	c.mu.RLock()
	for key, item := range c.data {
		// Genel TTL kontrolü
		if now.Sub(item.cachedAt) > c.ttl {
			expiredKeys = append(expiredKeys, key)
		}
	}
	c.mu.RUnlock()

	if len(expiredKeys) > 0 {
		c.mu.Lock()
		for _, key := range expiredKeys {
			delete(c.data, key)
		}
		c.mu.Unlock()
		log.Printf("[CACHE] Cleaned up %d expired items from in-memory cache.", len(expiredKeys))
	}
}

func (c *InMemoryCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists || time.Since(item.cachedAt) > c.ttl {
		return nil, false
	}
	return item.value, true
}

// ===== REDIS CACHE IMPLEMENTATION =====

// RedisCache Redis tabanlı önbellekleme için yapı
type RedisCache struct {
	client     *redis.Client
	ctx        context.Context
	defaultTTL time.Duration
}

// NewRedisCache yeni bir Redis cache instance'ı oluşturur
func NewRedisCache(addr string, password string, db int, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client:     client,
		ctx:        context.Background(),
		defaultTTL: ttl,
	}
}

func (c *RedisCache) GetOrSet(group, identifier string, dest any, fallback FallbackFunc) error {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// 1. Cache'den veriyi almayı dene.
	cachedValue, found := c.get(cacheKey)
	if found {
		// 2. Cache'de bulundu. Veriyi hedefe (dest) Unmarshal etmeyi dene.
		// Burada 'err' değişkenini if bloğunun dışında da kullanabilmek için önce tanımlıyoruz.
		var unmarshalErr error
		if unmarshalErr = json.Unmarshal(cachedValue, dest); unmarshalErr == nil {
			// Başarıyla çözümlendi! İşlem tamam, fonksiyondan çık.
			return nil
		}

		// Eğer buraya ulaştıysak, Unmarshal işlemi başarısız olmuştur.
		// Veri bozuk kabul edilir, loglanır ve taze veri için fallback'e devam edilir.
		log.Printf("CACHE_CORRUPT: Key %s. Re-fetching. Error on unmarshal: %v", cacheKey, unmarshalErr)
	}

	// 3. Cache'de yok ya da veri bozuk. Fallback fonksiyonunu çalıştırarak taze veriyi al.
	fallbackData, err := fallback()
	if err != nil {
		return err // Veritabanı veya ana kaynaktan veri alınamadı.
	}

	// 4. Fallback'ten gelen taze veriyi JSON'a çevir.
	bytes, err := json.Marshal(fallbackData)
	if err != nil {
		return fmt.Errorf("fallback data could not be marshalled to JSON: %w", err)
	}

	// 5. Bu taze veriyi (goroutine ile) arka planda cache'e kaydet.
	go func() {
		if saveErr := c.SaveCache(fallbackData, group, identifier); saveErr != nil {
			log.Printf("CACHE_SAVE_ERROR: Key %s. Error: %v", cacheKey, saveErr)
		}
	}()

	// 6. Son olarak, taze veriyi hedefe (dest) Unmarshal et.
	return json.Unmarshal(bytes, dest)
}

// TryCache önbellekteki veriyi kontrol eder ve varsa yanıt olarak döndürür
func (c *RedisCache) TryCache(ctx *gin.Context, group, identifier string) bool {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'ten veriyi al
	val, err := c.client.Get(c.ctx, cacheKey).Bytes()
	if err != nil {
		// Redis'te yok veya bir hata oluştu
		return false
	}

	// Cache hit - önbellekteki veriyi dön
	ctx.Data(http.StatusOK, "application/json", val)
	ctx.Header("X-Cache", "HIT")
	return true
}

// SaveCache yanıtı önbelleğe alır
func (c *RedisCache) SaveCache(response any, group, identifier string) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'e kaydet
	return c.client.Set(c.ctx, cacheKey, jsonData, c.defaultTTL).Err()
}

// SaveCacheTTL özel TTL ile yanıtı önbelleğe alır
func (c *RedisCache) SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'e özel TTL ile kaydet
	return c.client.Set(c.ctx, cacheKey, jsonData, ttl).Err()
}

func (c *RedisCache) Delete(group, identifier string) error {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)
	// Redis'ten ilgili anahtarı sil
	return c.client.Del(c.ctx, cacheKey).Err()
}

// ClearGroup bir grubu önbellekten temizler
func (c *RedisCache) ClearGroup(group string) {
	prefix := group + ":"

	// Redis'te desen araması yap
	iter := c.client.Scan(c.ctx, 0, prefix+"*", 0).Iterator()

	// Bulunan tüm anahtarları sil
	for iter.Next(c.ctx) {
		c.client.Del(c.ctx, iter.Val())
	}
}

// ClearAll tüm önbelleği temizler
func (c *RedisCache) ClearAll() {
	c.client.FlushAll(c.ctx)
}

// Stop Redis bağlantısını kapatır
func (c *RedisCache) Stop() {
	c.client.Close()
}

func (c *RedisCache) get(key string) ([]byte, bool) {
	val, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return val, true
}
