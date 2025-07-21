# 🚀 GitHub Actions ile Sıfır Kesintili (Zero-Downtime) Deployment

Bu rehber, GoLang backend uygulamalarını Ubuntu VPS'e systemd kullanmadan, sıfır kesinti ile deploy etmek için gereken tüm sunucu yapılandırmasını ve GitHub Actions iş akışını açıklar. Bu yöntem, "Blue-Green Deployment" stratejisinin basitleştirilmiş bir versiyonunu kullanır.

## 📋 İçindekiler

1. [Konsept: Sıfır Kesintili Deployment Nasıl Çalışır?](#konsept)
2. [Gerekli Sunucu Yapılandırması](#sunucu-yapılandırması)
3. [GitHub Actions Workflow (deploy.yml)](#github-actions-workflow)
4. [Manuel Kontrol ve Sorun Giderme](#manuel-kontrol)
5. [İlk Deploy Adımları](#ilk-deploy-adımları)

## 💡 1. Konsept: Sıfır Kesintili Deployment Nasıl Çalışır? {#konsept}

Bu sistem, systemd gibi statik servisler yerine, uygulama süreçlerini doğrudan script ile yöneterek kesintisiz geçiş yapmayı hedefler. Temel mantık, iki farklı port arasında (örneğin 4040 ve 4041) geçiş yapmaktır.

### Olay Akışı:

1. **Mevcut Durum**: Canlı uygulamanız Port A'da (4040) çalışıyor ve Nginx tüm trafiği bu porta yönlendiriyor.

2. **Deploy Başlar**: main branch'ine yeni kod gönderildiğinde GitHub Actions tetiklenir.

3. **Yeni Versiyon Hazırlanır**: Script, sunucuda yeni versiyonu kurar ve boşta olan Port B'de (4041) başlatır.

4. **Sağlık Kontrolü**: Script, Port B'deki yeni uygulamanın çalışıp çalışmadığını curl ile test eder.

5. **Anlık Geçiş (Soft Switch)**: Sağlık kontrolü başarılı olursa, script Nginx'in yapılandırmasını güncelleyerek gelen tüm yeni trafiği anında Port B'ye yönlendirir. Bu işlem `nginx -s reload` komutu sayesinde mevcut bağlantıları kesmeden yapılır.

6. **Eski Versiyon Durdurulur**: Geçiş tamamlandıktan sonra, artık trafik almayan Port A'daki eski uygulama güvenli bir şekilde sonlandırılır.

7. **Temizlik**: Belirlenen sayıdan daha eski release klasörleri sunucudan silinir.

Bir sonraki deploy'da bu süreç tersine işler: Canlı uygulama Port B'de çalışırken, yeni versiyon Port A'da test edilir ve geçiş yapılır.

## 🔧 2. Gerekli Sunucu Yapılandırması {#sunucu-yapılandırması}

GitHub Actions'ın çalışabilmesi için sunucunuzda bir defaya mahsus bu ayarları yapmanız gerekir.

### Nginx Kurulumu ve Dinamik Yönlendirme Ayarları

Bu yapılandırma, Nginx'in hangi porta gideceği bilgisini harici bir dosyadan okumasını sağlar. Deploy script'i sadece bu küçük dosyayı değiştirir.

#### Adım 1: Gerekli Dizini Oluşturun

Nginx'in otomatik taramadığı, bizim upstream konfigürasyonunu saklayacağımız özel bir dizin oluşturun.

```bash
sudo mkdir -p /etc/nginx/includes
```

#### Adım 2: Ana Nginx Konfigürasyonunu Oluşturun

Dosya: `/etc/nginx/sites-available/backend-template`

```nginx
# Yönlendirilecek backend uygulamasını tanımlayan upstream bloku.
upstream backend_app {
    # Port bilgisi, deploy script'i tarafından güncellenecek bu dosyadan okunacak.
    include /etc/nginx/includes/backend_upstream.conf;
}

server {
    listen 80;
    # SSL aktif edildiğinde bu satır otomatik olarak 443'e güncellenir.

    server_name senin-alan-adin.com.tr; # Kendi alan adınızı yazın

    location / {
        proxy_pass http://backend_app;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Cloudflare kullanıyorsanız bu header'ı ekleyin
        proxy_set_header CF-Connecting-IP $http_cf_connecting_ip;

        # Timeout ayarları
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    access_log /var/log/nginx/senin-alan-adin-access.log;
    error_log /var/log/nginx/senin-alan-adin-error.log;
}
```

#### Adım 3: Başlangıç Port Dosyasını Oluşturun

Deploy script'i bu dosyanın içeriğini her seferinde güncelleyecektir.

```bash
# Başlangıçta 4040 portunu dinlemesi için dosyayı oluşturun
echo "server 127.0.0.1:4040;" | sudo tee /etc/nginx/includes/backend_upstream.conf
```

#### Adım 4: Nginx'i Aktifleştirin ve Test Edin

```bash
# Oluşturduğunuz konfigürasyonu aktifleştirin
sudo ln -s /etc/nginx/sites-available/backend-template /etc/nginx/sites-enabled/

# Yazdığınız konfigürasyonda bir hata olup olmadığını kontrol edin
sudo nginx -t

# Her şey yolundaysa Nginx'i yeniden başlatın
sudo systemctl restart nginx
```

### Proje ve Durum Klasörleri

#### 1. Proje Ana Dizini (PROJECT_ROOT)
Sunucuda projenin yaşayacağı ana klasörü oluşturun (örneğin: `/root/2025-backend-template`). Bu yolu deploy.yml dosyasındaki `PROJECT_ROOT` değişkenine yazmalısınız.

#### 2. Merkezi .env Dosyası
Tüm hassas bilgileri (veritabanı şifreleri, API anahtarları vb.) içeren `.env` dosyanızı sadece bu ana dizine (`/root/2025-backend-template/.env`) koyun. Bu dosya Git'e dahil edilmemelidir. Deploy script'i, her yeni versiyonu kurarken bu merkezi dosyayı kopyalayacaktır.

#### 3. releases ve state Klasörleri
Bu klasörleri sizin manuel olarak oluşturmanıza gerek yoktur. Deploy script'i ilk çalıştığında bunları otomatik olarak oluşturacaktır.

- **releases**: Her deploy'un kendi zaman damgalı kopyasının yaşadığı yer.
- **state**: Sistemin "hafızası". `live.port` ve `app.pid` dosyalarını burada tutar.

## ⚙️ 3. GitHub Actions Workflow (deploy.yml) {#github-actions-workflow}

Bu dosya, yukarıda anlatılan tüm süreci otomatize eder.

### Temel Ayarlar ve Anlamları

Workflow dosyasının en üstündeki `env:` bloğu, tüm süreci yönetir:

- **PROJECT_ROOT**: Projenin sunucudaki ana yolu.
- **PORT_A & PORT_B**: Geçiş yapılacak iki port.
- **KEEP_RELEASES**: Sunucuda tutulacak eski versiyon sayısı.
- **INSTALL_COMMAND**: Bağımlılıkları kuran komut (go mod tidy, npm install vb.).
- **BUILD_COMMAND**: Projeyi derleyen komut (go build..., npm run build vb.).

### Adım Adım İşleyişi

Script çalıştığında loglarda göreceğiniz adımlar şunlardır:

1. **Ortamı Hazırla**: Gerekli değişkenleri ayarlar.
2. **Kodu Klonla**: Projenin son halini yeni bir release klasörüne indirir.
3. **.env Kopyala**: Merkezi .env dosyasını bu yeni klasöre kopyalar.
4. **Bağımlılıkları Yükle**: INSTALL_COMMAND komutunu çalıştırır.
5. **Build Et**: BUILD_COMMAND komutuyla uygulamayı derler.
6. **Portları Belirle**: state/live.port dosyasını okuyarak hangi portun boşta olduğunu anlar.
7. **Uygulamayı Başlat ve Test Et**: Yeni uygulamayı boş portta başlatır ve curl ile çalışıp çalışmadığını kontrol eder.
8. **Nginx'i Yönlendir**: includes/backend_upstream.conf dosyasını güncelleyip nginx -s reload ile geçişi yapar.
9. **Eski Süreci Durdur**: state/app.pid dosyasından eski uygulamanın PID'sini okur ve sonlandırır.
10. **Durumu Güncelle**: state klasöründeki dosyaları yeni port ve PID ile günceller.
11. **Temizlik Yap**: Eski release klasörlerini siler.

## 🔍 4. Manuel Kontrol ve Sorun Giderme {#manuel-kontrol}

Bir sorun yaşarsanız veya mevcut durumu kontrol etmek isterseniz bu komutları kullanabilirsiniz.

### Hangi port canlıda?
```bash
cat /root/2025-backend-template/state/live.port
```

### Çalışan uygulamanın PID'si nedir?
```bash
cat /root/2025-backend-template/state/app.pid
```

### Tüm Go süreçlerini acil durdurmak için:
```bash
sudo fuser -k 4040/tcp
sudo fuser -k 4041/tcp
```

### Son deploy'un uygulama loglarını görmek için:
```bash
# Önce son release klasörünü bulun
LATEST_RELEASE=$(ls -1tr /root/2025-backend-template/releases | tail -n 1)

# Sonra log dosyasını okuyun
cat /root/2025-backend-template/releases/$LATEST_RELEASE/app.log
```

## ✅ 5. İlk Deploy Adımları {#ilk-deploy-adımları}

1. Ubuntu sunucunuza Nginx'i kurun (`sudo apt install nginx`).

2. Yukarıdaki **Nginx Kurulumu ve Dinamik Yönlendirme Ayarları** adımlarını eksiksiz uygulayın.

3. Sunucuda proje ana klasörünü (`/root/2025-backend-template`) oluşturun.

4. Merkezi `.env` dosyanızı bu ana klasörün içine yerleştirin.

5. GitHub projenizin **Settings > Secrets and variables > Actions** menüsünden `HOST`, `USERNAME`, `PRIVATE_KEY` gibi secrets'larınızı tanımlayın.

6. `.github/workflows/deploy.yml` dosyasındaki `env` değişkenlerini kendi projenize göre düzenleyin.

7. Kodunuzu `main` branch'ine push'layın ve **Actions** sekmesinden deploy sürecini canlı olarak izleyin!
