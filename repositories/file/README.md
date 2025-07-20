# File Repository (`repositories/file`)

Bu paket, uygulama içindeki dosya metadatalarının yönetilmesinden sorumludur. Dosyaların kendisi Cloudflare R2 gibi bir object storage servisinde saklanırken, bu dosyalara ait bilgiler (URL, dosya tipi, kategori vb.) veritabanında bu repository tarafından yönetilir.

## Temel Çalışma Prensibi: İki Aşamalı Yükleme

Dosya yükleme süreci, güvenlik ve esneklik sağlamak amacıyla iki ana adımdan oluşur:

1.  **Ön-İmza (Presigned URL) Aşaması:**
    * Frontend, bir dosya yüklemek istediğinde doğrudan backend'e dosya göndermez. Bunun yerine, backend'den bu dosyayı R2'ye yükleyebileceği, süresi kısıtlı ve güvenli bir URL talep eder.
    * Bu aşamada, `files_signatures` tablosuna geçici bir kayıt atılır. Bu kayıt, hangi dosyanın yükleneceğini ve bu yükleme işleminin ne zaman sona ereceğini takip eder.
    * Bu işlemi `CreateUploadSignature` fonksiyonu yönetir.

2.  **Yükleme Onaylama (Confirmation) Aşaması:**
    * Frontend, aldığı ön-imzalı URL'i kullanarak dosyayı doğrudan R2'ye yükler.
    * Yükleme başarılı olduğunda, frontend backend'e geri dönerek "Ben bu dosyayı yükledim, artık kalıcı olarak kaydedebilirsin" mesajı gönderir.
    * Bu aşamada, `files` tablosuna dosyanın kalıcı kaydı oluşturulur (`CreateFileRecord` fonksiyonu ile) ve `files_signatures` tablosundaki ilgili geçici kayıt "tamamlandı" olarak işaretlenir (`MarkUploadAsCompleted` fonksiyonu ile).

Bu iki aşamalı yapı, büyük dosyaların backend sunucusunu yormasını engeller ve yükleme işlemlerini daha güvenli hale getirir.

## Fonksiyonlar

Bu repository'deki ana fonksiyonlar ve görevleri aşağıda açıklanmıştır. (Not: İsimlendirmeler, okunabilirliği artırmak için önerilen yeni halleriyledir.)

---

### `CreateUploadSignature`

Bir dosya yükleme işlemi başlamadan önce, bu işleme özel geçici bir "imza" kaydı oluşturur.

-   **Ne Yapar?:** Frontend'in dosyayı R2'ye yükleyebilmesi için gereken ön-imzalı URL ve diğer bilgileri `files_signatures` tablosuna kaydeder.
-   **Ne Alır?:** `context`, `types.UploadSignatureInput` (Presigned URL, dosya adı, tipi vb. bilgiler)
-   **Ne Döndürür?:** Oluşturulan imza kaydının `uuid.UUID`'si ve `error`.

```go
func (r *Repository) CreateUploadSignature(ctx context.Context, input types.UploadSignatureInput) (uuid.UUID, error)
```

---

### `CreateFileRecord`

Yüklemesi tamamlanmış ve onaylanmış bir dosyanın bilgilerini kalıcı olarak veritabanına kaydeder.

-   **Ne Yapar?:** Dosyanın nihai URL'i, adı, kategorisi gibi bilgilerle `files` tablosuna yeni bir kayıt ekler.
-   **Ne Alır?:** `context`, `types.SaveFileInput` (Nihai URL, dosya adı, boyutu vb. bilgiler)
-   **Ne Döndürür?:** Oluşturulan kalıcı dosya kaydının `uuid.UUID`'si ve `error`.

```go
func (r *Repository) CreateFileRecord(ctx context.Context, input types.SaveFileInput) (uuid.UUID, error)
```

---

### `GetUploadSignatureByID`

Belirli bir yükleme imzasının detaylarını ID'sine göre getirir.

-   **Ne Yapar?:** `files_signatures` tablosundan tek bir kaydı ID ile sorgular. Genellikle yükleme onayı sırasında kullanılır.
-   **Ne Alır?:** `context`, `uuid.UUID` (imza ID'si)
-   **Ne Döndürür?:** `*types.UploadSignature` (imza bilgileri) ve `error`.

```go
func (r *Repository) GetUploadSignatureByID(ctx context.Context, signatureID uuid.UUID) (*types.UploadSignature, error)
```

---

### `MarkUploadAsCompleted`

Bir yükleme imzasını "tamamlandı" olarak işaretler.

-   **Ne Yapar?:** Yükleme ve onaylama işlemi bittiğinde, `files_signatures` tablosundaki `completed` alanını `true` yapar.
-   **Ne Alır?:** `context`, `uuid.UUID` (imza ID'si)
-   **Ne Döndürür?:** `error`.

```go
func (r *Repository) MarkUploadAsCompleted(ctx context.Context, signatureID uuid.UUID) error
```

---

### `GetFileByID`

Kalıcı olarak kaydedilmiş bir dosyanın bilgilerini ID'sine göre getirir.

-   **Ne Yapar?:** `files` tablosundan tek bir kaydı ID ile sorgular.
-   **Ne Alır?:** `context`, `uuid.UUID` (dosya ID'si)
-   **Ne Döndürür?:** `*types.File` (dosya bilgileri) ve `error`.

```go
func (r *Repository) GetFileByID(ctx context.Context, fileID uuid.UUID) (*types.File, error)
```

---

### `GetFilesByCategory`

Belirli bir kategoriye ait tüm dosyaları listeler.

-   **Ne Yapar?:** `files` tablosundaki kayıtları `file_category` alanına göre filtreleyerek listeler.
-   **Ne Alır?:** `context`, `string` (kategori adı)
-   **Ne Döndürür?:** `[]types.File` (dosya listesi) ve `error`.

```go
func (r *Repository) GetFilesByCategory(ctx context.Context, category string) ([]types.File, error)
```

---

### `DeleteFileByID`

Bir dosyayı ID'sine göre "soft delete" yöntemiyle siler.

-   **Ne Yapar?:** Dosyayı veritabanından fiziksel olarak silmek yerine, `files` tablosundaki `status` alanını `'deleted'` olarak günceller. Bu sayede veri kaybı önlenir ve gerektiğinde geri alınabilir.
-   **Ne Alır?:** `context`, `uuid.UUID` (dosya ID'si)
-   **Ne Döndürür?:** `error`.

```go
func (r *Repository) DeleteFileByID(ctx context.Context, fileID uuid.UUID) error
```

## Önemli Notlar

-   **UUID Üretimi:** Bu repository'deki tüm `Primary Key` (`id`) değerleri, veritabanına `DEFAULT` olarak bırakılmamıştır. Bunun yerine, Go backend'inde `uuid.NewV7()` fonksiyonu ile oluşturulur ve `INSERT` sorgularıyla doğrudan veritabanına yazılır. Bu, veritabanı motorundan bağımsızlık sağlar.
-   **Silme Yöntemi:** Dosya silme işlemleri "soft delete" olarak yapılır. Kayıtlar veritabanından kaldırılmaz, sadece durumları güncellenir.
