# Sistem Otomasyon Servisi

Uygulamanın tüm otomasyon ihtiyaçlarını karşılayan merkezi ve kararlı bir çözüm. `robfig/cron` kütüphanesi üzerine inşa edilmiştir.

## 📋 İçindekiler

- [Genel Bakış](#genel-bakış)
- [Görev Türleri](#görev-türleri)
- [Temel Felsefe ve Güvenlik](#temel-felsefe-ve-güvenlik)
- [API Fonksiyonları](#api-fonksiyonları)
  - [Add](#add)
  - [Schedule](#schedule)
  - [Trigger](#trigger)
  - [CancelSchedule](#cancelschedule)
- [Tarih ve Saat Formatı](#tarih-ve-saat-formatı)
- [Örnekler](#örnekler)

## 🎯 Genel Bakış

Bu servis, kararlılık (stability) ve güvenilirlik (reliability) prensiplerine dayalı olarak tasarlanmış, iki farklı türde görevi yönetebilen bir otomasyon sistemidir.

## 📝 Görev Türleri

### 1. Periyodik Görevler
Sistem genelinde düzenli olarak tekrar eden işler.
- **Örnek**: "Her gün veritabanını yedekle"
- **Kullanım**: Sistem bakımı, cache temizleme, rapor oluşturma

### 2. Tek Seferlik Görevler
Kullanıcı aksiyonlarına bağlı, dinamik ve tek seferlik işler.
- **Örnek**: "Bu blog yazısını Cuma günü saat 14:30'da yayınla"
- **Kullanım**: İçerik zamanlama, bildirim gönderme

## 🔒 Temel Felsefe ve Güvenlik

### Çakışma Önleme (Race Condition Prevention)
- Bir iş çalışırken, aynı işin ikinci kez tetiklenmesini aktif olarak engeller
- Dahili "kilitleme (locking)" mekanizması ile sağlanır
- Sistem performansını ve veri tutarlılığını korur

### Panik Kurtarma (Panic Recovery)
- Beklenmedik çökmeler (panic) durumunda sistemi korur
- Hataları yakalar ve loglar
- İşin kilitli kalmasını önler
- Sistemin kendini onarmasını sağlar

### Merkezi Kayıt (Central Registry)
- Tüm görevler tek bir kayıt defterinde tutulur
- Durum takibi ve zamanlama yönetimi
- Basitleştirilmiş yönetim arayüzü

## 🛠 API Fonksiyonları

### Add

Sisteme periyodik olarak çalışacak statik bir iş ekler.

**Parametreler:**
- `key` (string): Görevi tanımlayan benzersiz anahtar
- `spec` (string): Standart cron formatı
- `job` (JobFunc): Çalıştırılacak fonksiyon

**Dönüş:** `error`

**Kullanım Zamanı:** Genellikle sunucu başlangıcında (`main.go`)

```go
// main.go içinde
cacheRevalidationJob := func() {
    log.Println("Clearing 'permissions' cache group...")
    cacheService.ClearGroup("permissions")
}

err := automationService.Add("cache:revalidate-perms", "@every 6h", cacheRevalidationJob)
if err != nil {
    log.Fatalf("Could not schedule permission revalidation: %v", err)
}
```

### Schedule

Dinamik ve tek seferlik bir işi gelecekteki belirli bir zamanda çalışmak üzere zamanlar.

**Parametreler:**
- `key` (string): Görevi tanımlayan benzersiz anahtar
- `at` (time.Time): İşin çalışacağı kesin zaman
- `job` (JobFunc): Çalıştırılacak fonksiyon

**Dönüş:** `error`

**Özel Özellikler:**
- Aynı key ile mevcut zamanlama varsa, eskisini siler ve yenisiyle değiştirir
- İş tamamlandığında kendini kayıtlardan otomatik siler

```go
// Blog handler içinde
publishJob := func() {
    h.blogRepo.SetStatus(postID, "published")
}
scheduleKey := fmt.Sprintf("blog:publish:%s", postID)

err := h.automationService.Schedule(scheduleKey, scheduledTime, publishJob)
```

### Trigger

Add ile eklenmiş periyodik bir görevi anında ve manuel olarak tetikler.

**Parametreler:**
- `key` (string): Tetiklenecek işin anahtarı

**Dönüş:** `error`

**Kullanım Alanı:** Admin panelleri, manuel işlem tetikleme

```go
// Admin handler içinde
err := h.automationService.Trigger("system:db-backup")
if err != nil {
    // Hata yönetimi...
}
```

### CancelSchedule

Schedule ile planlanmış tek seferlik bir görevi çalışma zamanı gelmeden iptal eder.

**Parametreler:**
- `key` (string): İptal edilecek işin anahtarı

```go
// Blog handler içinde
scheduleKey := fmt.Sprintf("blog:publish:%s", postID)
h.automationService.CancelSchedule(scheduleKey)
```

## 🕐 Tarih ve Saat Formatı

Zamanlama işlemlerinde belirsizlik ve zaman dilimi hatalarını önlemek için **ISO 8601 (UTC)** standardı kullanılmalıdır.

### Frontend (JavaScript)

Kullanıcıdan alınan Date nesnesi, backend'e gönderilmeden önce `toISOString()` metodu ile string'e çevrilmelidir.

```javascript
// Kullanıcının bir date picker ile seçtiği tarih
const userSelectedDate = new Date('2025-12-20T15:00:00+03:00'); // Türkiye saati (GMT+3)

// Backend'e yollanacak JSON verisi
const payload = {
  postId: "post-123",
  // toISOString() metodu, tarihi UTC'ye çevirir ve standart formata getirir
  scheduleAt: userSelectedDate.toISOString()
};

// payload.scheduleAt değeri: "2025-12-20T12:00:00.000Z"
// Bu format, tüm sunucular için aynı anı ifade eder

fetch('/api/schedule-post', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(payload)
});
```

### Backend (Go)

Backend, standart string'i `time.Parse` ve `time.RFC3339` sabiti ile `time.Time` nesnesine çevirir.

```go
type ScheduleRequest struct {
    PostID     string `json:"postId" binding:"required"`
    ScheduleAt string `json:"scheduleAt" binding:"required"` // Frontend'den string olarak alınır
}

func (h *MyHandler) HandleSchedule(c *gin.Context) {
    var req ScheduleRequest
    if err := c.ShouldBindJSON(&req); err != nil { /*...*/ }

    // Gelen ISO 8601 string'ini Go'nun time.Time nesnesine çevir
    scheduledTime, err := time.Parse(time.RFC3339, req.ScheduleAt)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601."})
        return
    }

    // scheduledTime artık automationService.Schedule fonksiyonunda kullanılmaya hazır
    h.automationService.Schedule("some-key", scheduledTime, someJob)
}
```

## 📚 Örnekler

### Cache Temizleme (Periyodik)
```go
// Her 6 saatte bir permissions cache'ini temizle
cacheJob := func() {
    log.Println("Clearing permissions cache...")
    cacheService.ClearGroup("permissions")
}

automationService.Add("cache:permissions", "@every 6h", cacheJob)
```

### Blog Post Zamanlama (Tek Seferlik)
```go
// Blog postunu belirli bir zamanda yayınla
publishJob := func() {
    blogService.Publish(postID)
    log.Printf("Blog post %s published", postID)
}

scheduleKey := fmt.Sprintf("blog:publish:%s", postID)
automationService.Schedule(scheduleKey, publishTime, publishJob)
```

### Veritabanı Yedeği (Manuel Tetikleme)
```go
// Admin panelinden manuel yedek başlat
func (h *AdminHandler) TriggerBackup(c *gin.Context) {
    err := h.automationService.Trigger("system:db-backup")
    if err != nil {
        c.JSON(500, gin.H{"error": "Backup could not be triggered"})
        return
    }
    c.JSON(200, gin.H{"message": "Backup started successfully"})
}
```

## 🔧 Gereksinimler

- Go 1.19+
- `github.com/robfig/cron/v3` paketi
- PostgreSQL (veritabanı işlemleri için)

## ⚠️ Önemli Notlar

1. **Benzersiz Anahtarlar**: Her görev için benzersiz bir anahtar kullanın
2. **Zaman Formatı**: Daima ISO 8601 (UTC) formatını kullanın
3. **Hata Yönetimi**: Tüm API çağrılarından dönen hataları kontrol edin
4. **Performans**: Uzun süren işleri goroutine içinde çalıştırmayı düşünün
5. **Loglama**: Kritik işlemler için uygun loglama ekleyin

---

Bu servis, modern web uygulamalarının otomasyon ihtiyaçlarını karşılamak üzere tasarlanmıştır ve production ortamında güvenle kullanılabilir.
