# Proje Güvenlik Mimarisi (SECURE.md)

Bu döküman, sunucu altyapısını bot saldırılarına ve diğer yetkisiz erişim denemelerine karşı korumak için uygulanan iki katmanlı güvenlik mimarisini detaylandırmaktadır. Amaç, minimum kaynak kullanımı ile maksimum güvenlik sağlamaktır.

## Savunma Felsefesi: Katmanlı Güvenlik (Kale Analojisi)

Savunma stratejimiz, bir kaleyi koruma mantığına dayanır:

**Dış Duvar (Hetzner Firewall):** Kaleyi (sunucuyu) dış dünyaya karşı görünmez kılan, sadece güvenilir elçilerin (Cloudflare) bildiği gizli kapılardan girişine izin veren aşılmaz surlar.

**İç Nöbetçi (Nginx):** Kaleye girmeyi başaran elçilerin getirdiği misafirleri (gerçek kullanıcı trafiği) denetleyen, şüpheli davranış sergileyenleri anında yakalayıp zindana (kara listeye) atan elit bir muhafız.

Bu iki katman, birbirinin alternatifi değil, birbirini tamamlayan bir bütündür.

## Katman 1: Stratejik Savunma - Hetzner Firewall Yapılandırması

Bu katmanın tek ve net bir amacı vardır: Sunucunun gerçek IP adresine doğrudan yapılan tüm saldırıları, trafik daha sunucuya ulaşmadan engellemek. Sunucuya sadece ve sadece Cloudflare üzerinden gelen trafiğe izin verilir.

### Adım 1: Güvenilir IP Listelerinin Alınması

Hetzner Firewall kurallarını oluşturmak için Cloudflare'in resmi olarak yayınladığı IP aralıkları kullanılır:

- **IPv4 Adresleri:** https://www.cloudflare.com/ips-v4
- **IPv6 Adresleri:** https://www.cloudflare.com/ips-v6

### Adım 2: Hetzner Firewall Kurallarının Yapılandırılması

Hetzner Cloud konsolu üzerinden, sunucuya uygulanmış olan Firewall'un "Rules" sekmesinde aşağıdaki kurallar tanımlanır:

#### Kural 1: SSH Erişimi (INBOUND)
- **Description:** Allow SSH from My IP
- **IPs:** Sadece kendi statik IP adresiniz. (Örn: `YOUR_STATIC_IP/32`)
- **Protocol:** TCP
- **Port:** 22

#### Kural 2: Cloudflare HTTP/S Erişimi (INBOUND)
- **Description:** Allow Cloudflare IPv4
- **IPs:** Yukarıdaki ips-v4 linkinden alınan tüm IPv4 aralıkları.
- **Protocol:** TCP
- **Port:** 80, 443 (Her iki port da virgülle eklenir)

#### Kural 3: Cloudflare IPv6 Erişimi (INBOUND)
- **Description:** Allow Cloudflare IPv6
- **IPs:** Yukarıdaki ips-v6 linkinden alınan tüm IPv6 aralıkları.
- **Protocol:** TCP
- **Port:** 80, 443

#### Kural 4 (İsteğe Bağlı): ICMP Erişimi (INBOUND)
- **Description:** Allow Ping
- **IPs:** Any IPv4, Any IPv6
- **Protocol:** ICMP

Bu yapılandırma tamamlandığında, sunucuya Cloudflare ve sizin IP'niz dışında hiç kimse erişemez.

## Katman 2: Taktiksel Savunma - Nginx Dinamik IP Engelleme

Bu katman, Cloudflare üzerinden gelen trafiği analiz eden Go uygulamamızın, tespit ettiği kötü niyetli IP'leri dinamik olarak engellemesini sağlar.

### Süreç Nasıl İşliyor?

1. Go uygulaması, bir IP'nin saldırgan olduğuna (Agresif Bot, Sürekli Yük vb.) karar verir.
2. Go uygulaması, bu IP'yi `/etc/nginx/blocklist.conf` dosyasına ekler.
3. Go uygulaması, `sudo nginx -s reload` komutunu çalıştırarak Nginx'in yeni kara listeyi hafızasına almasını sağlar.
4. Nginx, bir sonraki istekte bu IP'yi tanır ve isteği Go uygulamasına hiç ulaştırmadan 403 Forbidden hatasıyla engeller.

### Adım 1: Dinamik Kara Liste Dosyasını Oluşturma

Bu dosya, Go uygulaması tarafından yönetilecek olan "yasaklılar defteridir".

```bash
# Dosyayı oluşturun
sudo nano /etc/nginx/blocklist.conf
```

Dosyanın formatı basit bir "IP adresi - değer" eşleşmesidir. Değer her zaman 1 olacaktır.

```nginx
# IP Adresi      Değer
88.254.9.107     1;
123.123.123.123  1;
```

### Adım 2: Ana Nginx Yapılandırmasına map Bloğu Ekleme

Nginx'e, kara listeyi nasıl okuyacağını ve hangi değişkene atayacağını öğretiyoruz. Bu işlem `/etc/nginx/nginx.conf` dosyasında, http bloğu içinde yapılır.

```bash
# Ana Nginx yapılandırma dosyasını açın
sudo nano /etc/nginx/nginx.conf
```

`http { ... }` bloğunun içine, `include /etc/nginx/sites-enabled/*;` satırından önce aşağıdaki bloğu ekleyin:

```nginx
http {
    # ... mevcut diğer ayarlarınız ...

    # --- IP BLOKLAMA HARİTASI ---
    # Cloudflare'den gelen gerçek kullanıcı IP'sini ($http_cf_connecting_ip) alır.
    # Bu IP, blocklist.conf dosyasındaki bir IP ile eşleşirse, $is_blocked değişkenine "1" değerini atar.
    # Eşleşmezse, varsayılan olarak "0" değerini atar. Bu işlem son derece performanslıdır.
    map $http_cf_connecting_ip $is_blocked {
        default 0;
        include /etc/nginx/blocklist.conf;
    }
    # ---------------------------

    include /etc/nginx/sites-enabled/*;
    # ...
}
```

### Adım 3: Site Yapılandırmasında Engellemeyi Aktif Etme

Son olarak, sitemizin yapılandırma dosyasında, Nginx'e `$is_blocked` değişkeninin değeri 1 ise ne yapacağını söylüyoruz.

```bash
# Sitenizin yapılandırma dosyasını açın
sudo nano /etc/nginx/sites-available/backend-template
```

`server { ... }` bloğunun en üstüne, location bloklarından önce aşağıdaki kontrolü ekleyin:

```nginx
server {
    listen 80;
    server_name template.hoi.com.tr;

    # --- HARİTA SONUCUNU KONTROL ET ---
    # $is_blocked değişkeni "1" ise, isteği Go uygulamasına hiç göndermeden
    # 403 Forbidden hatasıyla anında reddet.
    if ($is_blocked) {
        return 403;
    }
    # ---------------------------------

    location / {
        # ... proxy_pass ve diğer proxy ayarlarınız ...
    }
}
```

### Adım 4: Yapılandırmayı Test Etme ve Uygulama

Yapılan tüm değişikliklerin doğru olduğundan emin olmak ve aktif hale getirmek için:

```bash
# Nginx yapılandırma dosyalarınızda bir yazım hatası olup olmadığını kontrol edin
sudo nginx -t

# Test başarılıysa, Nginx'i kesintisiz olarak yeniden yükleyin
sudo nginx -s reload
```

Bu adımların sonunda, `/etc/nginx/blocklist.conf` dosyasına eklenen herhangi bir IP'den gelen istek, Nginx tarafından anında engellenecektir. Go uygulaması artık sadece bu dosyayı yönetmekle sorumludur.
