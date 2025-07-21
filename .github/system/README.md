# 🚀 Ubuntu VPS - Nginx + Systemd Backend Deployment

GoLang backend uygulamalarını Ubuntu VPS'e deploy etmek için **sadece 2 katman**: Nginx (reverse proxy) + Systemd (servis yöneticisi).

## 📋 Temel Formül

1. **Nginx**: Domain'den backend'e yönlendirme (port 80/443 → 4040)
2. **Systemd**: Backend uygulamasını servis olarak çalıştırma

---

## 🔧 1. Nginx Konfigürasyonu

> ⚠️ **Çoklu proje**: Aynı sunucuda birden fazla proje varsa her proje için farklı port (4040, 5050, 6060...) kullanın.

**Dosya:** `/etc/nginx/sites-available/PROJE_ADI`

```nginx
server {
    listen 80;
    server_name DOMAIN_ADINIZ.com;  # 🔧 Değiştirin

    location / {
        proxy_pass http://localhost:4040;  # 🔧 Port değiştirin
        proxy_http_version 1.1;

        # Temel header'lar
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket desteği
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_cache_bypass $http_upgrade;

        # Timeout ayarları
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Log dosyaları
    access_log /var/log/nginx/PROJE_ADI-access.log;  # 🔧 Değiştirin
    error_log /var/log/nginx/PROJE_ADI-error.log;    # 🔧 Değiştirin
}
```

### Nginx Aktivasyonu
```bash
# Konfigürasyon aktifleştir
sudo ln -s /etc/nginx/sites-available/PROJE_ADI /etc/nginx/sites-enabled/

# Test et
sudo nginx -t

# Restart
sudo systemctl restart nginx
```

---

## ⚙️ 2. Systemd Servis

**Dosya:** `/etc/systemd/system/PROJE_ADI.service`

```ini
[Unit]
Description=PROJE_ADI Backend Service  # 🔧 Değiştirin
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/PROJE_ADI        # 🔧 Değiştirin
ExecStart=/root/PROJE_ADI/main          # 🔧 Değiştirin
Restart=on-failure
RestartSec=5
EnvironmentFile=/root/PROJE_ADI/.env    # 🔧 Değiştirin

# Log yönetimi
StandardOutput=journal
StandardError=journal

# Graceful shutdown
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
```

### Servis Aktivasyonu
```bash
# Servis yükle
sudo systemctl daemon-reload

# Başlat ve aktifleştir
sudo systemctl start PROJE_ADI
sudo systemctl enable PROJE_ADI

# Durum kontrol
sudo systemctl status PROJE_ADI
```

---

## 🔒 3. SSL (Let's Encrypt)

```bash
# Certbot kur
sudo apt install certbot python3-certbot-nginx -y

# SSL al (otomatik nginx güncellemesi)
sudo certbot --nginx -d DOMAIN_ADINIZ.com  # 🔧 Değiştirin

# Otomatik yenileme test
sudo certbot renew --dry-run
```

---

## 📊 4. Log İzleme

```bash
# Nginx logları
sudo tail -f /var/log/nginx/PROJE_ADI-access.log
sudo tail -f /var/log/nginx/PROJE_ADI-error.log

# Servis logları
sudo journalctl -u PROJE_ADI -f

# Son 100 satır
sudo journalctl -u PROJE_ADI -n 100
```

---

## 🚀 5. Deployment Adımları

```bash
# 1. Nginx kur
sudo apt update && sudo apt install nginx -y

# 2. Nginx konfigürasyon oluştur
sudo nano /etc/nginx/sites-available/PROJE_ADI

# 3. Aktifleştir
sudo ln -s /etc/nginx/sites-available/PROJE_ADI /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl restart nginx

# 4. Backend binary ve .env hazırla
# /root/PROJE_ADI/ klasörüne yerleştir

# 5. Systemd servis oluştur
sudo nano /etc/systemd/system/PROJE_ADI.service

# 6. Servisi başlat
sudo systemctl daemon-reload
sudo systemctl start PROJE_ADI
sudo systemctl enable PROJE_ADI

# 7. SSL kur
sudo certbot --nginx -d DOMAIN_ADINIZ.com

# 8. Kontrol et
sudo systemctl status PROJE_ADI
curl https://DOMAIN_ADINIZ.com
```

---

## 🚨 6. Sorun Giderme

### Yaygın Hatalar

**502 Bad Gateway:**
```bash
# Backend çalışıyor mu?
sudo systemctl status PROJE_ADI

# Port dinleniyor mu?
sudo netstat -tlnp | grep :4040
```

**Port çakışması:**
```bash
# Hangi process kullanıyor?
sudo lsof -i :4040

# Zorla durdur
sudo fuser -k 4040/tcp
```

**SSL sorunu:**
```bash
# DNS doğru mu?
nslookup DOMAIN_ADINIZ.com

# Sertifika durumu
sudo certbot certificates
```

---

## 🎯 Çoklu Proje Örneği

Aynı sunucuda 3 proje:

| Proje | Port | Domain | Servis |
|-------|------|--------|--------|
| API | 4040 | api.domain.com | api.service |
| Admin | 5050 | admin.domain.com | admin.service |
| User | 6060 | user.domain.com | user.service |

Her proje için aynı adımları tekrarlayın, sadece port ve isim değiştirin.

---

## 💡 Hızlı Komutlar

```bash
# Nginx test ve restart
sudo nginx -t && sudo systemctl restart nginx

# Servis restart
sudo systemctl restart PROJE_ADI

# Logları canlı izle
sudo journalctl -u PROJE_ADI -f

# Port kontrol
sudo netstat -tlnp | grep :4040

# SSL yenileme
sudo certbot renew --dry-run
```

**Not:** Rate limiting, monitoring, backup gibi gelişmiş özellikler için uygulamanızın gereksinimlerine göre ek konfigürasyonlar yapabilirsiniz. Bu rehber sadece temel deployment için yeterlidir.
