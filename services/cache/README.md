# Sistem Otomasyon Servisi (services/automation)

Bu servis, uygulamanızın **"zamanlama beynidir"**. Belirli işlerin gelecekte bir zamanda veya düzenli aralıklarla otomatik olarak yapılmasını sağlar. Tıpkı bir ajanda gibi, "şu işi her gün yap" veya "bu özel görevi Cuma günü saat 14:30'da yap" gibi komutları hatırlar ve zamanı geldiğinde uygular.

## 📋 İçindekiler

- [Temel Kavramlar: Mutfaktaki Şef Analojisi](#1-temel-kavramlar-mutfaktaki-şef-analojisi)
- [Dört Ana Komut](#2-dört-ana-komut-servis-ile-nasıl-konuşulur)
- [Önemli Uygulama Detayları](#3-önemli-uygulama-detayları)
- [Kod Örnekleri](#kod-örnekleri)

## 1. Temel Kavramlar: Mutfaktaki Şef Analojisi

Bu servisi anlamanın en kolay yolu, onu bir **restoran mutfağındaki şef** gibi düşünmektir.

### 🍳 İki Tür Görev Vardır: Rutin İşler ve Özel Siparişler

#### Rutin İşler (`Add` metodu ile eklenir)

**🔄 Analoji:** Bu, şefin her gün yapması gereken standart işlerdir.
- "Her sabah saat 8'de kahve makinesini çalıştır"
- "Her gece mutfağı temizle"

**⚙️ Teknik Karşılığı:** Bunlar, uygulamanın sağlığı için gerekli olan, tekrar eden sistem görevleridir.
- `Add` ile eklenirler
- Kalıcıdırlar ve uygulama çalıştığı sürece var olurlar
- Adminler bu rutin işleri `Trigger` ile manuel olarak da başlatabilir

**📝 Örnekler:**
- Veritabanı yedekleme
- Eski logları temizleme
- Saatlik cache yenileme

#### Özel Siparişler (`Schedule` metodu ile eklenir)

**🎂 Analoji:** Bu, müşteriden gelen tek seferlik özel bir istektir.
- "Masa 7 için Cuma günü saat 20:00'de özel bir pasta hazırla"
- Şef bu pastayı yapar, teslim eder ve bu siparişi ajandasından siler
- Bu sipariş bir daha tekrarlanmaz

**⚙️ Teknik Karşılığı:** Bunlar, genellikle kullanıcı aksiyonlarına bağlı olan, dinamik ve tek seferlik görevlerdir.
- `Schedule` ile planlanırlar
- Zamanı geldiğinde çalışırlar
- İşleri bittiğinde sistemden otomatik olarak silinirler

**📝 Örnekler:**
- Bir blog yazısını 2 gün sonra yayınlamak
- Kullanıcının deneme süresi bittiğinde e-posta göndermek

### 🛡️ Altın Kural: Mutfakta Asla Kaos Olmaz

**👨‍🍳 Analoji:** Yetenekli bir şef:
- Aynı anda iki tane aynı pastayı yapmaya başlamaz
- Kahve makinesi zaten çalışırken onu tekrar çalıştırmayı denemez

**⚙️ Teknik Karşılığı:** Bu servis:
- Bir işin (örneğin "veritabanı yedekleme") aynı anda birden fazla kez çalışmasını aktif olarak engeller
- Sistemin kararlılığını korur ve veri bozulması gibi felaket senaryolarının önüne geçer
- Bir iş çalışırken çökerse (panik), servis kendini toparlayabilir
- İşin "kilitli" kalmasını önler

## 2. Dört Ana Komut: Servis ile Nasıl Konuşulur?

Servisi yönetmek için **dört basit ve net komutunuz** var:

### 🔄 `Automation.Add()`

Sisteme **kalıcı ve periyodik** bir görev ekler. Genellikle uygulama başlarken `main.go` içinde kullanılır.

**📅 Ne Zaman Kullanılır?** Sistem genelindeki rutin, tekrarlanan işler için.

**📋 Parametreler:**
- `key` (string): Görevi tanımlayan benzersiz anahtar
  *Örn: `"system:db-backup:daily"`*
- `spec` (string): Ne zaman çalışacağını belirten standart cron ifadesi
  - `@daily` veya `0 0 * * *`: Her gece yarısı
  - `@hourly` veya `0 * * * *`: Her saat başı
  - `@every 1h30m`: Her 1 saat 30 dakikada bir
  - `0 4 * * SUN`: Her Pazar sabah 4'te
- `job` (func()): Çalıştırılacak olan Go fonksiyonu

```go
// main.go -> setupScheduledJobs
backupJob := func() {
    log.Println("Yedekleme başlıyor...")
}
automationService.Add("system:db-backup:daily", "@daily", backupJob)
```

### ⏰ `Automation.Schedule()`

Sisteme **tek seferlik ve dinamik** bir görev planlar.

**📅 Ne Zaman Kullanılır?** Bir kullanıcı eylemine bağlı olarak gelecekte bir defa yapılması gereken işler için.

**📋 Parametreler:**
- `key` (string): Genellikle ilgili veriyle ilişkili benzersiz bir anahtar
  *Örn: `"blog:publish:post-123"`*
- `at` (time.Time): Görevin çalışacağı kesin zaman
- `job` (func()): Çalıştırılacak olan Go fonksiyonu

**⚠️ Önemli Davranış:**
- Bu görev, zamanı geldiğinde çalıştıktan sonra kendini sistemden otomatik olarak siler
- Eğer aynı key ile yeni bir Schedule çağrısı yapılırsa, eski planlama silinir ve yenisiyle değiştirilir

```go
// Bir handler içinde
publishJob := func() {
    repository.PublishPost(postID)
}
scheduleKey := fmt.Sprintf("blog:publish:%s", postID)

// scheduledTime: Frontend'den gelen ve parse edilmiş zaman
automationService.Schedule(scheduleKey, scheduledTime, publishJob)
```

### ⚡ `Automation.Trigger()`

`Add` ile eklenmiş kalıcı bir görevi, **zamanını beklemeden anında** çalıştırmak için kullanılır.

**📅 Ne Zaman Kullanılır?**
- Bir adminin acil bir yedekleme başlatması
- Bir sistem görevini test etmek istemesi

**📋 Parametreler:**
- `key` (string): `Add` ile eklenmiş görevin anahtarı

**⚠️ Kısıtlama:** Bu fonksiyon, `Schedule` ile planlanmış tek seferlik görevleri tetikleyemez.

```go
// Bir admin API endpoint'i içinde
err := automationService.Trigger("system:db-backup:daily")
```

### ❌ `Automation.CancelSchedule()`

`Schedule` ile planlanmış bir görevi, **çalışma zamanı gelmeden iptal** etmek için kullanılır.

**📅 Ne Zaman Kullanılır?** Bir editörün, yayınlamayı planladığı bir blog yazısından vazgeçmesi gibi durumlar için.

**📋 Parametreler:**
- `key` (string): `Schedule` ile eklenmiş görevin anahtarı

```go
// Bir handler içinde
scheduleKey := fmt.Sprintf("blog:publish:%s", postID)
automationService.CancelSchedule(scheduleKey)
```

## 3. Önemli Uygulama Detayları

### 🔗 Görevlere Servis ve Repository'leri Nasıl Paslarım?

Bir görevin veritabanına erişmesi gerekebilir. `job` fonksiyonu `func()` tipinde olduğu için doğrudan parametre alamaz. **Çözüm: Go'nun "closure" özelliğini kullanmak.**

```go
// main.go
func setupScheduledJobs(auto *AutomationService.Service, userRepo *AuthRepository.Repository) {

    // inactiveUserCleanupJob, userRepo'ya ihtiyaç duyuyor.
    inactiveUserCleanupJob := func() {
        // Bu fonksiyon, bir üst kapsamdaki (scope) `userRepo` değişkenine
        // erişebilir ve onu kullanabilir.
        log.Println("Eski kullanıcılar temizleniyor...")
        userRepo.DeleteOldInactiveUsers()
    }

    // `userRepo`'yu parametre olarak vermedik, ama closure sayesinde
    // `inactiveUserCleanupJob` onu "hatırlıyor".
    auto.Add("user:cleanup", "@weekly", inactiveUserCleanupJob)
}
```

### 🌍 Frontend'den Tarih/Saat Nasıl Alınmalı? (ISO 8601 Standardı)

Zamanlama işlemlerinde hataları önlemek için backend ve frontend arasında **tek ve standart bir dil** konuşulmalıdır. Bu dil, **ISO 8601 (UTC)** formatıdır.

#### Frontend (JavaScript)
Kullanıcıdan alınan tarih, `toISOString()` ile evrensel formata çevrilip yollanmalıdır.

```javascript
const userSelectedDate = new Date('2025-12-20T18:00:00+03:00'); // Türkiye saati

const payload = {
  // .toISOString() -> "2025-12-20T15:00:00.000Z"
  // Bu string, tüm dünya için aynı anı ifade eder.
  scheduleAt: userSelectedDate.toISOString()
};

// fetch(..., { body: JSON.stringify(payload) });
```

#### Backend (Go)
Gelen bu standart string, `time.Parse` ile kolayca Go'nun anladığı `time.Time` nesnesine çevrilir.

```go
// Bir handler içinde
var req struct {
    ScheduleAt string `json:"scheduleAt"`
}
c.ShouldBindJSON(&req);

// time.RFC3339, ISO 8601 formatına karşılık gelen sabittir.
scheduledTime, err := time.Parse(time.RFC3339, req.ScheduleAt)
if err != nil {
    // Hata yönetimi
}

// `scheduledTime` artık servise verilmeye hazır.
automationService.Schedule("some-key", scheduledTime, someJob)
```

## 📚 Kod Örnekleri

### Periyodik Cache Temizleme
```go
// Her 6 saatte bir cache temizle
cacheJob := func() {
    log.Println("Cache temizleniyor...")
    cacheService.ClearGroup("permissions")
}

automationService.Add("cache:permissions", "@every 6h", cacheJob)
```

### Blog Post Zamanlama
```go
// Frontend'den gelen zamanlama isteği
type SchedulePostRequest struct {
    PostID     string `json:"postId" binding:"required"`
    ScheduleAt string `json:"scheduleAt" binding:"required"`
}

func (h *BlogHandler) SchedulePost(c *gin.Context) {
    var req SchedulePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // ISO 8601 string'ini time.Time'a çevir
    scheduledTime, err := time.Parse(time.RFC3339, req.ScheduleAt)
    if err != nil {
        c.JSON(400, gin.H{"error": "Geçersiz tarih formatı"})
        return
    }

    // Yayınlama görevini oluştur
    publishJob := func() {
        h.blogRepo.SetStatus(req.PostID, "published")
        log.Printf("Blog post %s yayınlandı", req.PostID)
    }

    // Görevi zamanla
    scheduleKey := fmt.Sprintf("blog:publish:%s", req.PostID)
    err = h.automationService.Schedule(scheduleKey, scheduledTime, publishJob)
    if err != nil {
        c.JSON(500, gin.H{"error": "Zamanlama başarısız"})
        return
    }

    c.JSON(200, gin.H{"message": "Post zamanlandı"})
}
```

### Admin Panel Manuel Tetikleme
```go
// handlers/blog/schedule-post.go

package BlogHandler

import (
    "net/http"
    "time"
    "fmt"

    "github.com/gin-gonic/gin"
    // Diğer importlar: AutomationService, types vb.
)


// ScheduleRequest, frontend'den gelen JSON verisini karşılamak için kullanılır.
// Tarih burada `string` olarak alınır.
type ScheduleRequest struct {
    PostID     string `json:"postId" binding:"required"`
    ScheduleAt string `json:"scheduleAt" binding:"required"`
}


func (h *BlogHandler) SchedulePostHandler(c *gin.Context) {
    var req ScheduleRequest

    // 1. Adım: Frontend'den gelen JSON'ı `req` struct'ına bind et.
    // `req.ScheduleAt` şu an "2025-12-20T12:00:00.000Z" gibi bir string.
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formatı."})
        return
    }

    // 2. Adım: String'i `time.Time` nesnesine dönüştür. (En Kritik Adım)
    // time.RFC3339 sabiti, "2025-12-20T12:00:00.000Z" formatını tanır.
    scheduledTime, err := time.Parse(time.RFC3339, req.ScheduleAt)
    if err != nil {
        // Eğer frontend yanlış bir format yollarsa, burada hata alırız.
        c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tarih formatı. Lütfen ISO 8601 (UTC) formatı kullanın."})
        return
    }

    // `scheduledTime` artık Go'nun anladığı bir time.Time nesnesidir.

    // 3. Adım: `time.Time` nesnesini otomasyon servisine ver.
    publishJob := func() {
        // Bu fonksiyon, zamanı geldiğinde çalışacak olan asıl iştir.
        log.Printf("Blog post %s yayınlanıyor...", req.PostID)
        // h.blogRepository.Publish(req.PostID)
    }

    scheduleKey := fmt.Sprintf("blog:publish:%s", req.PostID)

    err = h.automationService.Schedule(scheduleKey, scheduledTime, publishJob)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Görevin planlanması sırasında bir hata oluştu."})
        return
    }

    // 4. Adım: Başarılı yanıtı döndür.
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": fmt.Sprintf("Post '%s' başarıyla %s tarihine planlandı.", req.PostID, scheduledTime.Format(time.RFC1123)),
    })
}
```

---

**💡 Bu servis, modern web uygulamalarının otomasyon ihtiyaçlarını karşılamak üzere tasarlanmıştır ve production ortamında güvenle kullanılabilir.**
