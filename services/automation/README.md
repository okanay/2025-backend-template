# Sistem Otomasyon Servisi (`services/automation`)

Bu servis, uygulamanın tüm otomasyon ihtiyaçlarını karşılayan merkezi bir çözümdür. Rutin görevleri otomatikleştirerek sistem yönetimini kolaylaştırır ve operasyonel verimliliği artırır.

## Temel Kabiliyetler

### Cache Otomasyonu
- Popüler içeriklerin proaktif cache'lenmesi
- Kullanıcı izinlerinin önbelleğe alınması
- Sistem ayarlarının cache yenilenmesi
- Olay güdümlü cache geçersiz kılma (yeni post eklendiğinde vs.)
- Zamanlanmış cache yenileme işlemleri

### Sistem Komut Yürütme
- Makefile komutlarının güvenli çalıştırılması
- Database migration işlemleri (`make migrate-up`, `make migrate-down`)
- Database backup işlemleri (`make migrate-db-pull`)
- Docker container yönetimi
- Git işlemleri (pull, deploy)
- Dosya sistemı operasyonları

### Zamanlanmış Görevler
- Günlük/haftalık/aylık periyodik işlemler
- Veritabanı yedekleme (her gece, haftalık)
- Log dosyası temizleme
- Eski dosyaların arşivlenmesi
- Sistem sağlık kontrolleri
- Cache yenileme döngüleri

### Manuel Tetikleme
- Admin paneli üzerinden anlık komut çalıştırma
- Emergency işlemler için hızlı erişim
- Test ve geliştirme süreçleri için kontrollü tetikleme
- Maintenance işlemleri

### Çakışma Yönetimi
- Aynı job'ın aynı anda iki kez çalışmasını engelleme
- Manuel tetikleme ile otomatik scheduler koordinasyonu
- Job öncelik sistemi (kritik işlemler önce)
- Resource conflict çözümü

### Güvenlik ve Kontrol
- Komut whitelist sistemi
- Admin yetkilendirme kontrolü
- Command injection koruması
- Job timeout yönetimi
- Error handling ve recovery
- Job execution tracking

## Kullanım Senaryoları

**Veritabanı Yönetimi**
- Otomatik günlük backup'lar
- Migration işlemlerinin güvenli yürütülmesi
- Database maintenance görevleri

**Sistem Bakımı**
- Log dosyalarının periyodik temizlenmesi
- Eski dosyaların arşivlenmesi
- Disk alanı optimizasyonu
- Geçici dosya temizleme

**Cache Optimizasyonu**
- Popüler içeriklerin önceden yüklenmesi
- Kullanıcı session verilerinin yönetimi
- API response cache'lerinin yenilenmesi

**Deployment ve DevOps**
- Otomatik code deployment
- Container restart işlemleri
- Environment güncelleme
- Health check işlemleri

**Monitoring ve Alerting**
- Sistem kaynaklarının izlenmesi
- Performance metriklerinin toplanması
- Kritik durumlarda otomatik müdahale

Bu servis, system admin görevlerini minimize ederek, geliştiricilerin core business logic'e odaklanmasını sağlar.
