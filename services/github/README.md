# GitHub Service (`services/github`)

Bu paket, uygulamanın GitHub üzerindeki bir repository ile etkileşim kurmasını sağlayan bir servis katmanıdır. Özellikle CMS (İçerik Yönetim Sistemi) benzeri yapıların, içeriklerini doğrudan bir Git repository'sinde yönetmesi için tasarlanmıştır. `go-github` kütüphanesini kullanarak GitHub API'sine istekler gönderir.

## Temel Çalışma Prensibi: Draft (Taslak) ve Publish (Yayınla) Akışı

Bu servisin temel amacı, "Infrastructure as Code" veya "Content as Code" prensiplerini uygulamaktır. Değişiklikler doğrudan ana branch'e (`main`) gönderilmez. Bunun yerine, aşağıdaki akış izlenir:

1.  **Taslak Branch'i Oluşturma:** Bir içerik (örn: tema CSS'i, dil dosyası) değiştirilmek istendiğinde, `main` branch'inden `i18n-draft` veya `theme-draft` gibi yeni bir "taslak branch'i" oluşturulur.
2.  **Değişiklikleri Commit'leme:** Tüm değişiklikler bu taslak branch'ine commit edilir. Bu sayede ana branch her zaman stabil ve canlıya alınmaya hazır kalır.
3.  **Yayınlama:** Değişiklikler onaylandığında, servis otomatik olarak taslak branch'inden `main` branch'ine bir **Pull Request (PR)** oluşturur.
4.  **Birleştirme ve Temizlik:** Oluşturulan bu PR, yine otomatik olarak birleştirilir (merge edilir) ve sonrasında kullanılan taslak branch'i silinir. Bu, Git akışını temiz ve yönetilebilir tutar.

## Fonksiyonlar

---

### `Branch` Yönetimi

-   **`CreateBranch(baseBranch, newBranch)`:** Bir temel branch'ten (genellikle `main`) yeni bir branch oluşturur.
-   **`DeleteBranch(branch)`:** Belirtilen branch'i siler.
-   **`BranchExists(branch)`:** Bir branch'in var olup olmadığını kontrol eder.

---

### `Dosya` Yönetimi

-   **`GetFileContent(branch, path)`:** Belirli bir branch'teki bir dosyanın içeriğini ve o anki SHA hash'ini getirir. SHA, dosyayı güncellemek için gereklidir.
-   **`CommitFile(branch, path, content, sha, message)`:** Bir dosyayı belirtilen branch'e commit'ler. Eğer `sha` boş ise yeni bir dosya oluşturur; dolu ise mevcut dosyayı günceller.

---

### `Değişiklik` ve `Yayınlama` Yönetimi

-   **`GetBranchChanges(baseBranch, compareBranch)`:** İki branch arasındaki farkları (eklenen, silinen, değiştirilen dosyalar) listeler.
-   **`PublishBranchToMain(branch)`:** Yukarıda açıklanan "Draft ve Publish" akışının son adımını gerçekleştirir: Otomatik olarak bir Pull Request oluşturur, bunu `main` branch'i ile birleştirir ve işlemin raporunu döndürür.

```go
func (r *Service) PublishBranchToMain(branch string) (report string, err error)
```
