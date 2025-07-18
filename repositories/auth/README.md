# Auth Repository (`repositories/auth`)

Bu paket, kullanıcı kimlik doğrulama ve kullanıcı verilerinin yönetimiyle ilgili tüm veritabanı işlemlerinden sorumludur. Kullanıcıların sisteme kaydolması, giriş yapması, sosyal medya hesaplarıyla bağlanması ve profil bilgilerinin yönetilmesi gibi temel operasyonları içerir.

## Temel Çalışma Prensibi ve Tablolar

Bu repository, temel olarak iki ana tabloyu yönetir:

1.  **`users` Tablosu:** Kullanıcının kimlik (ID, email, şifre hash'i), rol (User, Admin vb.) ve statü (Active, Suspended vb.) gibi temel bilgilerini tutar.
2.  **`user_details` Tablosu:** Kullanıcının adı, soyadı, avatar URL'i gibi opsiyonel ve sosyal medya sağlayıcılarından gelen profil bilgilerini barındırır. Bu tablo, `users` tablosuyla `user_id` üzerinden bire bir ilişkiye sahiptir.

## Fonksiyonlar

Bu repository'deki ana fonksiyonlar ve görevleri aşağıda açıklanmıştır.

---

### `CreateUser`

Şifre ile yeni bir kullanıcı oluşturur.

-   **Ne Yapar?:** Bir transaction başlatarak hem `users` tablosuna ana kullanıcı kaydını hem de `user_details` tablosuna bu kullanıcıya ait boş bir profil kaydını ekler. Bu sayede veri bütünlüğü garanti altına alınır.
-   **Ne Alır?:** `context`, `types.UserCreateRequest` (email ve şifre)
-   **Ne Döndürür?:** Oluşturulan `*types.User` (kullanıcı bilgileri) ve `error`.

```go
func (r *Repository) CreateUser(ctx context.Context, data types.UserCreateRequest) (*types.User, error)
```

---

### `FindOrCreateFromProvider`

Sosyal medya sağlayıcısından (Google, Apple vb.) gelen bilgilerle kullanıcıyı bulur veya yeni bir kullanıcı oluşturur.

-   **Ne Yapar?:**
    1.  Önce `provider_id` ile kullanıcıyı arar. Bulursa son giriş zamanını güncelleyip döndürür.
    2.  Bulamazsa, bu kez `email` ile arar. Eğer aynı e-postaya sahip bir kullanıcı varsa, bu kullanıcının profilini yeni sosyal medya bilgileriyle güncelleyerek hesapları birleştirir.
    3.  Eğer e-posta ile de bulunamazsa, hem `users` hem de `user_details` tablolarına tamamen yeni bir kullanıcı kaydı oluşturur.
-   **Ne Alır?:** `context`, `*types.ProviderUserData` (sağlayıcıdan gelen tüm kullanıcı verileri)
-   **Ne Döndürür?:** Bulunan veya yeni oluşturulan `*types.User` ve `error`.

```go
func (r *Repository) FindOrCreateFromProvider(ctx context.Context, data *types.ProviderUserData) (*types.User, error)
```

---

### `SelectByEmail` / `SelectByID`

Bir kullanıcıyı e-posta adresine veya ID'sine göre veritabanından getirir.

-   **Ne Yapar?:** `users` tablosunda `email` veya `id` alanına göre tek bir kullanıcı kaydını sorgular.
-   **Ne Alır?:** `context`, `string` (email) veya `uuid.UUID` (ID)
-   **Ne Döndürür?:** `*types.User` ve `error`.

```go
func (r *Repository) SelectByEmail(ctx context.Context, email string) (*types.User, error)
func (r *Repository) SelectByID(ctx context.Context, id uuid.UUID) (*types.User, error)
```

---

### `UpdateLastLogin`

Bir kullanıcının son giriş yaptığı zamanı günceller.

-   **Ne Yapar?:** `users` tablosundaki `last_login` alanını mevcut zamanla (`NOW()`) günceller.
-   **Ne Alır?:** `context`, `uuid.UUID` (kullanıcı ID'si)
-   **Ne Döndürür?:** `error`.

```go
func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
```

---

### `SelectUserDetailsByID`

Bir kullanıcının profil detaylarını (`user_details` tablosundan) getirir.

-   **Ne Yapar?:** `user_id`'ye göre `user_details` tablosundan ilgili kaydı sorgular.
-   **Ne Alır?:** `context`, `uuid.UUID` (kullanıcı ID'si)
-   **Ne Döndürür?:** `*types.UserDetails` ve `error`. Eğer detay kaydı yoksa hata yerine `nil, nil` döner.

```go
func (r *Repository) SelectUserDetailsByID(ctx context.Context, userID uuid.UUID) (*types.UserDetails, error)
```

## Önemli Notlar

-   **UUID Üretimi:** Bu repository'de oluşturulan tüm yeni kayıtların `ID`'leri, veritabanı yerine Go backend'inde `uuid.NewV7()` ile üretilir ve sorguyla birlikte gönderilir.
-   **Transaction Kullanımı:** Birden fazla tabloyu etkileyen `CreateUser` ve `FindOrCreateFromProvider` gibi kritik fonksiyonlar, veri tutarlılığını sağlamak için `transaction` blokları içinde çalışır.
