# Validation Service (`services/validation`)

Bu servis, uygulamanıza gelen tüm verilerin doğruluğunu ve tutarlılığını garanti altına alan merkezi bir "kalite kontrol" katmanıdır.

## 🎯 Temel Felsefe

**Bir verinin kuralları, verinin kendisiyle birlikte, `types` paketinde yaşamalıdır.**

Bu servis, bu kuralları okur, uygular ve bir hata varsa frontend'e anlaşılır, yapısal bir yanıt döner.

## 📋 İçindekiler

- [Temel Kullanım](#temel-kullanım)
- [Validate Tag'leri](#validate-tagleri)
- [Frontend Hata Yapısı](#frontend-hata-yapısı)
- [Özel Validasyon Kuralları](#özel-validasyon-kuralları)

## 🚀 Temel Kullanım

Go struct'larınıza ekleyeceğiniz `validate` etiketleri (tags) ile kuralları belirlersiniz:

```go
// types/profile.go
package types

type UpdateProfileRequest struct {
    // Bu alan zorunludur, geçerli bir e-posta formatında olmalıdır.
    Email string `json:"email" validate:"required,email"`

    // Bu alan zorunlu DEĞİLDİR (çünkü pointer *), ama varsa en az 3,
    // en fazla 20 karakter olmalıdır.
    Username *string `json:"username" validate:"omitempty,min=3,max=20"`

    // Bu alan zorunlu DEĞİLDİR, ama varsa 0'dan büyük veya eşit olmalıdır.
    Age *int `json:"age" validate:"omitempty,gte=0"`

    // Bu alan zorunlu DEĞİLDİR, ama varsa bu üç değerden biri olmalıdır.
    Gender *string `json:"gender" validate:"omitempty,oneof=male female other"`
}
```

## 🏷️ Validate Tag'leri

### En Sık Kullanılan Etiketler

| Etiket | Açıklama | Örnek |
|--------|----------|-------|
| `required` | Alanın boş veya sıfır olmamasını sağlar | `validate:"required"` |
| `email` | Geçerli e-posta formatı zorunluluğu | `validate:"required,email"` |
| `min=X` | String için min karakter, sayı için min değer | `validate:"min=3"` |
| `max=X` | String için max karakter, sayı için max değer | `validate:"max=20"` |
| `gte=X` | Sayısal alan için "büyük veya eşit" kontrolü | `validate:"gte=0"` |
| `lte=X` | Sayısal alan için "küçük veya eşit" kontrolü | `validate:"lte=100"` |
| `oneof=a b c` | Belirtilen seçeneklerden biri olma zorunluluğu | `validate:"oneof=male female other"` |
| `omitempty` | **ÖNEMLİ:** Opsiyonel alanlar için diğer kuralları atlar | `validate:"omitempty,min=3"` |

### omitempty Etiketi Detayı

`omitempty` etiketi, opsiyonel alanlar yaratmak için kritik öneme sahiptir. Eğer alan istek içinde gönderilmemişse veya `nil` ise, bu alandaki diğer validasyon kuralları tetiklenmez. Bu, **"opsiyonel ama kurallı"** alanlar yaratmak için kullanılır.

## ⚠️ Frontend Hata Yapısı (En Önemli Kısım)

Validasyon başarısız olduğunda, servis frontend'in kolayca işleyebileceği, yapısal bir JSON yanıtı döner. Bu yanıt, tek bir hata mesajı değil, **geçersiz olan her alan için ayrı bir nesne** içeren bir `errors` dizisi barındırır.

### Örnek Senaryo

Kullanıcı kayıt formunu `email` alanını boş bırakarak ve `password` alanına sadece "123" yazarak gönderdi.

### Backend'den Dönecek HTTP 400 Yanıtı:

```json
{
    "success": false,
    "error": "validation_error",
    "errors": [
        {
            "field": "email",
            "tag": "required",
            "message": "email alanı zorunludur."
        },
        {
            "field": "password",
            "tag": "min",
            "message": "password alanı en az 8 karakter olmalıdır."
        }
    ]
}
```

### Frontend Avantajları

Bu yapı sayesinde frontend geliştiricisi:

- `errors` dizisini döngüye alabilir
- `field` adına göre ("email", "password") doğru form elemanını bulabilir
- `message` içeriğini o elemanın altına hata mesajı olarak yazdırabilir

## ⚙️ Özel Validasyon Kuralları

Basit kurallar yetmediğinde (örneğin "kullanıcı adı veritabanında daha önce alınmış mı?"), kendi özel kurallarınızı ekleyebilirsiniz.

### Senaryo Örneği

Yeni kullanıcı kaydı için `username` alanı belirli bir formata uymalı (`username_format`).

### 1. Kuralı Type Struct'ında Tanımlayın

```go
// types/user.go
package types

type UserRegisterRequest struct {
    // Yeni, özel kuralımızı ekledik: `username_format`
    Username string `json:"username" validate:"required,min=3,max=20,username_format"`
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### 2. Kural Fonksiyonunu Yazın

```go
// services/validation/rules.go (veya validator.go'nun alt kısmı)
package ValidationService

import (
    "regexp"
    "github.com/go-playground/validator/v10"
)

// Sadece harf, rakam ve alt çizgiye izin veren basit bir regex kuralı
var usernameFormatRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func isValidUsernameFormat(fl validator.FieldLevel) bool {
    return usernameFormatRegex.MatchString(fl.Field().String())
}
```

### 3. Yeni Kuralı Servise Kaydedin

```go
// services/validation/validator.go -> NewService() fonksiyonunu güncelliyoruz

func NewService() *Service {
    validate := validator.New()

    // "username_format" tag'i, artık isValidUsernameFormat fonksiyonunu çalıştıracak.
    validate.RegisterValidation("username_format", isValidUsernameFormat)

    return &Service{validate: validate}
}
```

### 4. Özel Hata Mesajını Ekleyin

```go
// services/validation/validator.go -> generateErrorMessage() fonksiyonunu güncelliyoruz

func (s *Service) generateErrorMessage(e validator.FieldError) string {
    field := e.Field()
    tag := e.Tag()

    switch tag {
    // ... (diğer case'ler)
    case "username_format":
        return fmt.Sprintf("%s alanı sadece harf, rakam ve alt çizgi (_) içerebilir.", field)
    // ...
    default:
        return "Geçersiz alan." // Varsayılan mesaj
    }
}
```

## 📝 Önemli Notlar

- **Merkezi Yaklaşım:** Tüm validasyon kuralları `types` paketinde struct tag'leri olarak tanımlanır
- **Yapısal Hata Yanıtları:** Frontend için kolayca işlenebilir JSON formatında hata mesajları
- **Genişletilebilirlik:** Yeni özel kurallar kolayca eklenebilir
- **Kullanıcı Dostu:** Anlaşılır hata mesajları otomatik olarak üretilir

## 🤝 Katkıda Bulunma

Yeni validasyon kuralları eklerken:

1. Kuralı `services/validation/rules.go` dosyasına ekleyin
2. `NewService()` fonksiyonunda kaydedin
3. `generateErrorMessage()` fonksiyonuna hata mesajını ekleyin
4. Kullanım örneğini dokümante edin
