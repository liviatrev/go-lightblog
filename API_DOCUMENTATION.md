# Dokumentasi API - go-lightblog

Dokumentasi ini menjelaskan secara lengkap seluruh endpoint API yang tersedia di aplikasi **go-lightblog**, termasuk metode autentikasi, pembagian hak akses (role), parameter request, serta contoh request menggunakan `curl` dan format response JSON.

---

## Daftar Isi
1. [Informasi Umum & Base URL](#1-informasi-umum--base-url)
2. [Autentikasi & Otorisasi](#2-autentikasi--otorisasi)
3. [Endpoint Kategori (Category)](#3-endpoint-kategori-category)
4. [Endpoint Label (Tag)](#4-endpoint-label-tag)
5. [Endpoint Pengguna (User)](#5-endpoint-pengguna-user)
6. [Endpoint Pengaturan (Settings)](#6-endpoint-pengaturan-settings)
7. [Endpoint Artikel (Post)](#7-endpoint-artikel-post)
8. [Endpoint SEO Generator](#8-endpoint-seo-generator)
9. [Endpoint Upload Gambar](#9-endpoint-upload-gambar)
10. [Endpoint Publik (Public Endpoints)](#10-endpoint-publik-public-endpoints)
11. [Cache Control & CDN](#11-cache-control--cdn)
12. [IndexNow Integration](#12-indexnow-integration)

---

## 1. Informasi Umum & Base URL

Secara default, server **go-lightblog** berjalan pada port `5800` menggunakan TCP atau melalui Unix Socket. 

- **Base URL API Admin**: `http://localhost:5800/api/v1/admin`
- **Base URL API Publik**: `http://localhost:5800`
- **Format Data**: Seluruh request body (kecuali upload gambar) dan response menggunakan format **JSON**. Untuk upload artikel yang menyertakan cover gambar, gunakan format **multipart/form-data**.

---

## 2. Autentikasi & Otorisasi

### Autentikasi menggunakan Bearer Token (API Key)
Semua endpoint di bawah `/api/v1/admin/*` bersifat privat dan membutuhkan header autentikasi berupa **API Key** milik user yang terdaftar. API Key ini dapat ditemukan di database atau di halaman profil pengguna.

Kirimkan token tersebut melalui header HTTP:
```http
Authorization: Bearer <API_KEY_ANDA>
```

### Otorisasi Berbasis Role (RBAC)
Aplikasi membagi hak akses ke dalam 2 tingkatan role:
1. **admin**: Memiliki akses penuh ke seluruh endpoint API (termasuk manajemen pengguna dan pengaturan global website).
2. **editor**: Hanya memiliki akses untuk melihat, membuat, mengupdate, dan menghapus **Artikel (Posts)**, **Kategori (Categories)**, dan **Label (Tags)**.

---

## 3. Endpoint Kategori (Category)

Grup Endpoint: `/api/v1/admin/categories`  
Akses: **admin**, **editor**

### A. Mendapatkan Semua Kategori
Mengambil daftar seluruh kategori yang tersedia.

*   **URL**: `/`
*   **Method**: `GET`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X GET http://localhost:5800/api/v1/admin/categories \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Name": "Teknologi",
      "Slug": "teknologi"
    },
    {
      "ID": 2,
      "Name": "Gaya Hidup",
      "Slug": "gaya-hidup"
    }
  ]
}
```

### B. Membuat Kategori Baru
Membuat kategori baru. Jika field `slug` dikosongkan, slug akan digenerate otomatis dari `name`.

*   **URL**: `/`
*   **Method**: `POST`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "name": "Kategori Baru",
      "slug": "kategori-baru"  // Opsional
    }
    ```

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/api/v1/admin/categories \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Kesehatan",
    "slug": "kesehatan-dan-kebugaran"
  }'
```

#### Contoh Response (201 Created)
```json
{
  "success": true,
  "message": "Category created successfully",
  "data": {
    "ID": 3,
    "Name": "Kesehatan",
    "Slug": "kesehatan-dan-kebugaran",
    "posts": null
  }
}
```

> **Catatan**: Setelah kategori dibuat/diupdate, sistem secara otomatis melakukan **Cloudflare cache purge** untuk URL kategori dan homepage (jika Cloudflare diaktifkan), serta **purge sitemap.xml**.

### C. Mengupdate Kategori
Memperbarui data kategori berdasarkan ID (bisa update nama, slug, atau keduanya).

*   **URL**: `/:id`
*   **Method**: `PUT`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "name": "Kesehatan & Kebugaran",
      "slug": "kesehatan-kebugaran"
    }
    ```

#### Contoh Request (curl)
```bash
curl -X PUT http://localhost:5800/api/v1/admin/categories/3 \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Kesehatan & Kebugaran"
  }'
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Category updated successfully",
  "data": {
    "ID": 3,
    "Name": "Kesehatan & Kebugaran",
    "Slug": "kesehatan-dan-kebugaran",
    "posts": null
  }
}
```

### D. Menghapus Kategori
Menghapus kategori berdasarkan ID dari database.

*   **URL**: `/:id`
*   **Method**: `DELETE`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X DELETE http://localhost:5800/api/v1/admin/categories/3 \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Category deleted successfully"
}
```

---

## 4. Endpoint Label (Tag)

Grup Endpoint: `/api/v1/admin/tags`  
Akses: **admin**, **editor**

### A. Mendapatkan Semua Label
Mengambil daftar seluruh label yang tersedia.

*   **URL**: `/`
*   **Method**: `GET`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X GET http://localhost:5800/api/v1/admin/tags \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Name": "Go",
      "Slug": "go"
    },
    {
      "ID": 2,
      "Name": "Tutorial",
      "Slug": "tutorial"
    }
  ]
}
```

### B. Membuat Label Baru
Membuat label baru. Jika field `slug` dikosongkan, slug digenerate otomatis berdasarkan `name`.

*   **URL**: `/`
*   **Method**: `POST`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "name": "Web Development",
      "slug": "web-dev"  // Opsional
    }
    ```

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/api/v1/admin/tags \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "JavaScript"
  }'
```

#### Contoh Response (201 Created)
```json
{
  "success": true,
  "message": "Tag created successfully",
  "data": {
    "ID": 3,
    "Name": "JavaScript",
    "Slug": "javascript",
    "posts": null
  }
}
```

> **Catatan**: Setelah tag dibuat/diupdate, sistem secara otomatis melakukan **Cloudflare cache purge** untuk URL tag dan homepage (jika Cloudflare diaktifkan), serta **purge sitemap.xml**.

### C. Mengupdate Label
Memperbarui data label berdasarkan ID.

*   **URL**: `/:id`
*   **Method**: `PUT`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "name": "JS Modern",
      "slug": "js-modern"
    }
    ```

#### Contoh Request (curl)
```bash
curl -X PUT http://localhost:5800/api/v1/admin/tags/3 \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "JS Modern"
  }'
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Tag updated successfully",
  "data": {
    "ID": 3,
    "Name": "JS Modern",
    "Slug": "javascript",
    "posts": null
  }
}
```

### D. Menghapus Label
Menghapus label berdasarkan ID.

*   **URL**: `/:id`
*   **Method**: `DELETE`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X DELETE http://localhost:5800/api/v1/admin/tags/3 \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Tag deleted successfully"
}
```

---

## 5. Endpoint Pengguna (User)

Grup Endpoint: `/api/v1/admin/users`  
Akses: **admin** saja (Editor tidak diperbolehkan mengakses endpoint ini).

### A. Mendapatkan Semua Pengguna
Mengambil daftar semua pengguna di database. Kolom password dan API Key disembunyikan secara otomatis.

*   **URL**: `/`
*   **Method**: `GET`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X GET http://localhost:5800/api/v1/admin/users \
  -H "Authorization: Bearer token_admin"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "data": [
    {
      "ID": 1,
      "Username": "admin",
      "Name": "Super Administrator",
      "Role": "admin"
    },
    {
      "ID": 2,
      "Username": "livia",
      "Name": "Livia Editor",
      "Role": "editor"
    }
  ]
}
```

### B. Membuat Pengguna Baru
Membuat pengguna baru dengan password terenkripsi (Bcrypt) dan secara otomatis menggenerate API Key baru.

*   **URL**: `/`
*   **Method**: `POST`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "username": "budi",
      "name": "Budi Raharjo",
      "password": "rahasia_budi",
      "role": "editor"  // "admin" atau "editor". Default "editor" jika kosong
    }
    ```

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/api/v1/admin/users \
  -H "Authorization: Bearer token_admin" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "budi",
    "name": "Budi Raharjo",
    "password": "rahasia_budi",
    "role": "editor"
  }'
```

#### Contoh Response (201 Created)
```json
{
  "success": true,
  "message": "User added successfully!",
  "data": {
    "ID": 3,
    "Username": "budi",
    "Name": "Budi Raharjo",
    "Role": "editor"
  }
}
```

### C. Mengupdate Data Pengguna (Partial)
Memperbarui data pengguna secara sebagian. Mendukung pembaruan username, name, role, dan password (akan di-hash otomatis jika diisi).

*   **URL**: `/:id`
*   **Method**: `PUT`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body** (Kirimkan field yang ingin diubah saja):
    ```json
    {
      "name": "Budi Raharjo S.Kom",
      "password": "password_baru_budi"
    }
    ```

#### Contoh Request (curl)
```bash
curl -X PUT http://localhost:5800/api/v1/admin/users/3 \
  -H "Authorization: Bearer token_admin" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi Raharjo S.Kom"
  }'
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "User data updated successfully!",
  "data": {
    "ID": 3,
    "Username": "budi",
    "Name": "Budi Raharjo S.Kom",
    "Role": "editor"
  }
}
```

### D. Menghapus Pengguna
Menghapus data pengguna dari database (soft delete karena model User memiliki `gorm.DeletedAt`).

*   **URL**: `/:id`
*   **Method**: `DELETE`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X DELETE http://localhost:5800/api/v1/admin/users/3 \
  -H "Authorization: Bearer token_admin"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

---

## 6. Endpoint Pengaturan (Settings)

Grup Endpoint: `/api/v1/admin/settings`  
Akses: **admin** saja

### A. Mendapatkan Semua Pengaturan
Mengambil seluruh konfigurasi sistem dalam bentuk map key-value yang ringkas dan bersih.

*   **URL**: `/`
*   **Method**: `GET`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X GET http://localhost:5800/api/v1/admin/settings \
  -H "Authorization: Bearer token_admin"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "data": {
    "site_title": "Blog Livia",
    "site_description": "Platform berbagi cerita dan teknologi",
    "site_keywords": "go, fiber, blogging",
    "site_headline": "Selamat Datang di LightBlog",
    "site_tagline": "Sederhana, Cepat, Modern",
    "upload_mode": "local",
    "imagekit_private_key": "",
    "imagekit_folder": "/lightblog",
    "remark42_url": "http://127.0.0.1:8080",
    "remark42_site_id": "lightblog-utama",
    "enable_gemini": "no",
    "gemini_api_key": "",
    "gemini_model": "gemini-flash-latest",
    "enable_cloudflare": "no",
    "cloudflare_api_key": "",
    "cloudflare_zone_id": "",
    "site_url": "",
    "public_theme": "light",
    "indexnow": "no",
    "indexnow_key": "a1b2c3...",
    "indexnow_submitted": "no",
    "header_script": "",
    "footer_script": ""
  }
}
```

### B. Mengupdate Pengaturan (Partial Update)
Memperbarui satu atau beberapa pengaturan secara dinamis. Sistem menerapkan filter keamanan ketat (allowlist) sehingga hanya key tertentu yang diizinkan untuk diubah.

#### Daftar Key yang Diizinkan (Allowlist)
- `site_title`, `site_description`, `site_keywords`, `site_headline`, `site_tagline`
- `upload_mode`, `imagekit_private_key`, `imagekit_folder`
- `remark42_url`, `remark42_site_id`
- `enable_gemini`, `gemini_api_key`, `gemini_model`
- `login_token`
- `enable_cloudflare`, `cloudflare_api_key`, `cloudflare_zone_id`, `site_url`
- `public_theme`
- `indexnow`, `indexnow_key`
- `header_script`, `footer_script`

*   **URL**: `/`
*   **Method**: `PUT`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: application/json`
*   **Body**:
    ```json
    {
      "site_title": "My New LightBlog",
      "enable_gemini": "yes",
      "gemini_api_key": "AIzaSy...",
      "public_theme": "ocean",
      "enable_cloudflare": "yes",
      "cloudflare_api_key": "abc123...",
      "cloudflare_zone_id": "zone-id-123",
      "site_url": "https://myblog.com",
      "indexnow": "yes"
    }
    ```

#### Contoh Request (curl)
```bash
curl -X PUT http://localhost:5800/api/v1/admin/settings \
  -H "Authorization: Bearer token_admin" \
  -H "Content-Type: application/json" \
  -d '{
    "site_title": "My New LightBlog",
    "site_tagline": "Blogging Engine Tercepat"
  }'
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Settings updated partially successfully.",
  "updated_keys": [
    "site_title",
    "site_tagline"
  ]
}
```

---

## 7. Endpoint Artikel (Post)

Grup Endpoint: `/api/v1/admin/posts`  
Akses: **admin**, **editor**

### A. Mendapatkan Semua Artikel (dengan Pagination)
Mengambil daftar artikel yang diurutkan dari yang terbaru (`created_at desc`). Mendukung pagination via query parameter `limit` dan `offset` untuk efisiensi loading database.

*   **URL**: `/`
*   **Method**: `GET`
*   **Headers**: `Authorization: Bearer <API_KEY>`
*   **Query Parameters**:
    *   `limit` (integer, default: 50, max: 100): Jumlah data per halaman.
    *   `offset` (integer, default: 0): Index awal pengambilan data.

#### Contoh Request (curl)
```bash
curl -X GET "http://localhost:5800/api/v1/admin/posts?limit=10&offset=0" \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "total": 25,
  "limit": 10,
  "offset": 0,
  "data": [
    {
      "ID": 12,
      "Title": "Panduan Belajar Go untuk Pemula",
      "Slug": "panduan-belajar-go-untuk-pemula",
      "Content": "<p>Go adalah bahasa pemrograman yang tangguh...</p>",
      "IsDraft": false,
      "CoverImage": "/public/uploads/cover-1234.jpg",
      "Type": "post",
      "MetaTitle": "Belajar Go Pemula",
      "MetaDescription": "Panduan terlengkap belajar bahasa pemrograman Go",
      "TargetKeyword": "belajar go",
      "AuthorID": 1,
      "Author": {
        "ID": 1,
        "Username": "admin",
        "Name": "Super Admin",
        "Role": "admin"
      },
      "CategoryID": 1,
      "Category": {
        "ID": 1,
        "Name": "Teknologi",
        "Slug": "teknologi"
      },
      "Tags": [
        {
          "ID": 1,
          "Name": "Go",
          "Slug": "go"
        }
      ],
      "CreatedAt": "2026-08-09T22:00:00Z",
      "UpdatedAt": "2026-08-09T22:00:00Z"
    }
  ]
}
```

### B. Membuat Artikel Baru
Membuat artikel baru. Endpoint ini menerima input berupa **multipart/form-data** karena mendukung pengunggahan file gambar cover artikel (`cover`).

Jika fitur **AI SEO (Gemini)** diaktifkan pada sistem (`enable_gemini = yes` dan `gemini_api_key` terisi), sistem akan secara otomatis memicu AI untuk memproses konten artikel dan menggenerate `MetaTitle`, `MetaDescription`, dan `TargetKeyword` yang relevan secara internal.

Jika field `cover` tidak diunggah, sistem akan **otomatis menggenerate cover image** berukuran 1200x630 piksel dengan judul artikel yang ditampilkan di atasnya (menggunakan `default-cover.jpg` sebagai background).

*   **URL**: `/`
*   **Method**: `POST`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: multipart/form-data`
*   **Form Parameters**:
    *   `title` (string, wajib): Judul artikel.
    *   `content` (string, wajib): Konten utama artikel (mendukung tag HTML).
    *   `category_id` (integer, opsional): ID kategori tujuan.
    *   `tags` (string, opsional): Daftar ID label dipisahkan koma, contoh: `1,2,5`.
    *   `isDraft` (string, opsional, "true"/"false"): Simpan sebagai draf atau langsung publish.
    *   `cover` (file, opsional): File gambar cover artikel (PNG, JPG, WebP, GIF, BMP).
    *   `seo_title` (string, opsional): Meta title kustom (digunakan jika AI mati).
    *   `seo_description` (string, opsional): Meta description kustom.
    *   `seo_keywords` (string, opsional): Target keyword kustom.

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/api/v1/admin/posts \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -F "title=Belajar Framework Fiber di Go" \
  -F "content=<p>Fiber adalah framework web terinspirasi dari Expressjs...</p>" \
  -F "category_id=1" \
  -F "tags=1,3" \
  -F "isDraft=false" \
  -F "cover=@/path/ke/gambar/cover.jpg"
```

#### Contoh Response (201 Created)
```json
{
  "success": true,
  "message": "Article published successfully!",
  "data": {
    "ID": 13,
    "Title": "Belajar Framework Fiber di Go",
    "Slug": "belajar-framework-fiber-di-go",
    "Content": "<p>Fiber adalah framework web terinspirasi dari Expressjs...</p>",
    "IsDraft": false,
    "CoverImage": "/public/uploads/cover-hash.jpg",
    "Type": "post",
    "MetaTitle": "Belajar Framework Fiber di Go",
    "MetaDescription": "Tutorial membuat web backend dengan bahasa Go dan framework Fiber",
    "TargetKeyword": "framework fiber go",
    "AuthorID": 2,
    "Author": {
      "ID": 2,
      "Username": "livia",
      "Name": "Livia Editor",
      "Role": "editor"
    },
    "CategoryID": 1,
    "Category": {
      "ID": 1,
      "Name": "Teknologi",
      "Slug": "teknologi"
    },
    "Tags": [
      {
        "ID": 1,
        "Name": "Go",
        "Slug": "go"
      },
      {
        "ID": 3,
        "Name": "JavaScript",
        "Slug": "javascript"
      }
    ],
    "CreatedAt": "2026-08-09T22:30:15Z",
    "UpdatedAt": "2026-08-09T22:30:15Z"
  }
}
```

> **Catatan**: Setelah artikel dibuat, sistem secara otomatis:
> - Melakukan **Cloudflare cache purge** untuk URL post, kategori, semua tag, dan homepage.
> - **Purge sitemap.xml**.
> - **Submit URL ke IndexNow** (jika `indexnow = yes` dan artikel bukan draft).

### C. Mengupdate Artikel (Partial Update)
Memperbarui data artikel secara sebagian. Menggunakan **multipart/form-data** agar tetap dapat melakukan update file cover baru secara opsional. Kirimkan parameter yang ingin diubah saja.

Jika judul diubah, slug akan digenerate ulang secara otomatis dan **SlugRedirect** akan dibuat untuk mengarahkan slug lama ke slug baru (301 redirect).

*   **URL**: `/:id`
*   **Method**: `PUT`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: multipart/form-data`
*   **Form Parameters** (Kirim yang ingin diubah):
    *   `title` (string)
    *   `content` (string)
    *   `category_id` (integer)
    *   `isDraft` (string, "true"/"false")
    *   `tags` (string, contoh: `1,2` atau `""` untuk menghapus seluruh tag)
    *   `seo_title` (string)
    *   `seo_description` (string)
    *   `seo_keywords` (string)
    *   `cover` (file)

#### Contoh Request (curl)
```bash
curl -X PUT http://localhost:5800/api/v1/admin/posts/13 \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -F "title=Belajar Framework Fiber Go Terlengkap" \
  -F "isDraft=true"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Article updated partially successfully!",
  "data": {
    "ID": 13,
    "Title": "Belajar Framework Fiber Go Terlengkap",
    "Slug": "belajar-framework-fiber-di-go",
    "Content": "<p>Fiber adalah framework web terinspirasi dari Expressjs...</p>",
    "IsDraft": true,
    "CoverImage": "/public/uploads/cover-hash.jpg",
    "Type": "post",
    "MetaTitle": "Belajar Framework Fiber di Go",
    "MetaDescription": "Tutorial membuat web backend dengan bahasa Go dan framework Fiber",
    "TargetKeyword": "framework fiber go",
    "AuthorID": 2,
    "Author": {
      "ID": 2,
      "Username": "livia",
      "Name": "Livia Editor",
      "Role": "editor"
    },
    "CategoryID": 1,
    "Category": {
      "ID": 1,
      "Name": "Teknologi",
      "Slug": "teknologi"
    },
    "Tags": [
      {
        "ID": 1,
        "Name": "Go",
        "Slug": "go"
      },
      {
        "ID": 3,
        "Name": "JavaScript",
        "Slug": "javascript"
      }
    ],
    "CreatedAt": "2026-08-09T22:30:15Z",
    "UpdatedAt": "2026-08-09T22:35:40Z"
  }
}
```

### D. Menghapus Artikel (Soft Delete)
Menghapus artikel berdasarkan ID menggunakan mekanisme **Soft Delete**. Baris data tidak akan langsung dihapus dari database SQLite, melainkan diisi kolom `deleted_at`-nya, sehingga artikel pindah ke tempat sampah (Trash).

*   **URL**: `/:id`
*   **Method**: `DELETE`
*   **Headers**: `Authorization: Bearer <API_KEY>`

#### Contoh Request (curl)
```bash
curl -X DELETE http://localhost:5800/api/v1/admin/posts/13 \
  -H "Authorization: Bearer token_admin_atau_editor"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "message": "Article moved to trash successfully"
}
```

> **Catatan**: Setelah artikel dihapus, sistem melakukan **Cloudflare cache purge** dan **submit URL ke IndexNow** agar search engine menghapus URL tersebut dari index.

---

## 8. Endpoint SEO Generator

Grup Endpoint: `/seo/generate`  
Akses: **admin**, **editor** (melalui session login, bukan Bearer Token)

Endpoint ini digunakan untuk menggenerate metadata SEO (MetaTitle, MetaDescription, TargetKeyword) menggunakan **Google Gemini AI** berdasarkan konten artikel yang dikirim.

*   **URL**: `/seo/generate`
*   **Method**: `POST`
*   **Headers**: 
    *   `Content-Type: application/json`
    *   Session cookie (harus login terlebih dahulu)
*   **Body**:
    ```json
    {
      "content": "<p>Konten artikel yang akan dianalisis oleh AI...</p>"
    }
    ```

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/seo/generate \
  -H "Content-Type: application/json" \
  -b "session_cookie=..." \
  -d '{
    "content": "<p>Go adalah bahasa pemrograman yang dikembangkan oleh Google...</p>"
  }'
```

#### Contoh Response (200 OK)
```json
{
  "meta_title": "Panduan Lengkap Belajar Go untuk Pemula",
  "meta_description": "Pelajari bahasa pemrograman Go dari dasar hingga mahir dengan panduan lengkap ini.",
  "target_keyword": "belajar go, golang tutorial"
}
```

#### Status Gagal Terkait:
- `400 Bad Request`: Format JSON tidak valid.
- `500 Internal Server Error`: Gemini API Key tidak dikonfigurasi, atau AI gagal merespons.

---

## 9. Endpoint Upload Gambar

Grup Endpoint: `/api/upload`  
Akses: **admin**, **editor** (melalui session login, bukan Bearer Token)

Endpoint ini digunakan untuk mengunggah gambar dari editor WYSIWYG (SunEditor) ke penyimpanan lokal atau ImageKit CDN.

*   **URL**: `/api/upload`
*   **Method**: `POST`
*   **Headers**: 
    *   `Content-Type: multipart/form-data`
    *   Session cookie (harus login terlebih dahulu)
*   **Form Parameters**:
    *   `image` (file, wajib): File gambar yang akan diunggah (PNG, JPG, WebP, GIF, BMP).

#### Contoh Request (curl)
```bash
curl -X POST http://localhost:5800/api/upload \
  -H "Content-Type: multipart/form-data" \
  -b "session_cookie=..." \
  -F "image=@/path/ke/gambar.jpg"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "url": "/public/uploads/gambar_1712345678.jpg"
}
```

#### Status Gagal Terkait:
- `500 Internal Server Error`: Gagal mengunggah ke CDN atau penyimpanan lokal.

### B. Upload Gambar via API Key (Cover & In-Content)

Grup Endpoint: `/api/v1/admin/upload`  
Akses: **admin**, **editor** (melalui Bearer Token)

Endpoint ini digunakan untuk mengunggah gambar **cover image** maupun **in-content image** melalui REST API menggunakan API Key (Bearer Token). Endpoint menerima file dari field `image` (untuk in-content image) atau field `cover` (untuk cover image), sehingga satu endpoint dapat melayani kedua kebutuhan.

*   **URL**: `/api/v1/admin/upload`
*   **Method**: `POST`
*   **Headers**: 
    *   `Authorization: Bearer <API_KEY>`
    *   `Content-Type: multipart/form-data`
*   **Form Parameters**:
    *   `image` (file, opsional): File gambar in-content yang akan diunggah (PNG, JPG, WebP, GIF, BMP).
    *   `cover` (file, opsional): File gambar cover yang akan diunggah (PNG, JPG, WebP, GIF, BMP).

> **Catatan**: Kirim salah satu field (`image` atau `cover`). Jika keduanya dikirim, field `image` yang diprioritaskan.

#### Contoh Request (curl) - In-Content Image
```bash
curl -X POST http://localhost:5800/api/v1/admin/upload \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -F "image=@/path/ke/gambar.jpg"
```

#### Contoh Request (curl) - Cover Image
```bash
curl -X POST http://localhost:5800/api/v1/admin/upload \
  -H "Authorization: Bearer token_admin_atau_editor" \
  -F "cover=@/path/ke/cover.jpg"
```

#### Contoh Response (200 OK)
```json
{
  "success": true,
  "url": "/public/uploads/gambar_1712345678.jpg"
}
```

#### Status Gagal Terkait:
- `401 Unauthorized`: Header `Authorization` tidak dikirim atau API Key salah/tidak terdaftar.
- `403 Forbidden`: Pengguna dengan role selain `admin`/`editor` mencoba mengakses endpoint ini.
- `500 Internal Server Error`: Gagal mengunggah ke CDN atau penyimpanan lokal.

---

## 10. Endpoint Publik (Public Endpoints)

Endpoint di bawah ini dapat diakses oleh siapa saja tanpa membutuhkan autentikasi Bearer Token.

### A. Proxy Thumbnail Gambar
Menghasilkan atau menyajikan thumbnail gambar berukuran kustom secara instan dan best-effort dengan konversi format yang didukung (WebP/JPG). Endpoint ini menerapkan browser caching otomatis jangka panjang (`Cache-Control: public, max-age=31536000`).

*   **URL**: `/api/thumb`
*   **Method**: `GET`
*   **Query Parameters**:
    *   `src` (string, wajib): Path relatif/absolut dari file gambar asli di server (contoh: `/public/uploads/cover.png`).
    *   `w` (integer, opsional, default: 600, max: 2000): Lebar (width) piksel thumbnail yang diinginkan.
    *   `f` (string, opsional, "jpg" atau "webp", default: "jpg"): Format file output target.

#### Contoh Request (Browser atau curl)
```bash
curl -I "http://localhost:5800/api/thumb?src=/public/uploads/cover-123.jpg&w=300&f=webp"
```

#### Contoh Response Header (200 OK)
```http
HTTP/1.1 200 OK
Content-Type: image/webp
Cache-Control: public, max-age=31536000
Content-Length: 12542
```
*(Response body akan berupa raw binary file gambar WebP berukuran lebar 300px).*

#### Status Gagal Terkait:
- `403 Forbidden`: Jika file asal berada di luar direktori yang diizinkan (Proteksi path traversal).
- `404 Not Found`: Jika file asal gambar (`src`) tidak ditemukan di disk server.
- `400 Bad Request`: Jika parameter format atau dimensi terlalu ekstrim / tidak didukung.
- `500 Internal Server Error`: Gagal memproses thumbnail.

### B. Halaman Publik (SSR)

| Route | Method | Deskripsi |
| :--- | :--- | :--- |
| `/` | `GET` | Homepage dengan daftar artikel terbaru (pagination). |
| `/post/:slug` | `GET` | Halaman detail artikel. |
| `/page/:slug` | `GET` | Halaman statis (menggunakan handler yang sama dengan post). |
| `/search?q=:query` | `GET` | Pencarian artikel berdasarkan judul, meta description, atau konten. |
| `/category/:slug` | `GET` | Arsip artikel berdasarkan kategori. |
| `/tag/:slug` | `GET` | Arsip artikel berdasarkan tag. |
| `/author/:id` | `GET` | Arsip artikel berdasarkan penulis. |
| `/manifest.json` | `GET` | Web App Manifest dinamis dengan warna tema. |
| `/sitemap.xml` | `GET` | Sitemap XML dinamis. |
| `/feed.xml` | `GET` | RSS 2.0 Feed (20 artikel terbaru). |
| `/rss.xml` | `GET` | Alias dari `/feed.xml`. |
| `/:key.txt` | `GET` | File verifikasi IndexNow key. |
| `/robots.txt` | `GET` | Robots.txt dinamis. |

---

## 11. Cache Control & CDN

Aplikasi **go-lightblog** menerapkan middleware `CacheControl` secara global yang
mengatur header `Cache-Control` berdasarkan jenis route. Ini memungkinkan
penggunaan CDN (seperti Cloudflare) secara optimal sekaligus menjaga data sensitif
tidak tersimpan di cache.

### Aturan Cache Berdasarkan Route

| Route Pattern | Cache-Control Header | Keterangan |
| :--- | :--- | :--- |
| `/public/*` (js, css, image) | `public, max-age=31536000, immutable` | Static files di-cache agresif selama 1 tahun dan ditandai immutable |
| `/` (homepage) | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/category/:slug` | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/tag/:slug` | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/search` | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/author/:id` | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/post/:slug` | `public, max-age=3600, s-maxage=604800` | Browser cache 1 jam, CDN cache 7 hari |
| `/page/:slug` | `public, max-age=3600, s-maxage=604800` | Browser cache 1 jam, CDN cache 7 hari |
| `/sitemap.xml`, `/feed.xml`, `/rss.xml` | `public, max-age=3600, s-maxage=86400` | Browser cache 1 jam, CDN cache 1 hari |
| `/manifest.json` | `public, max-age=60, s-maxage=86400` | Browser cache 1 menit, CDN cache 1 hari |
| `/api/*` | `no-store` | REST-API tidak pernah di-cache |
| `/dashboard`, `/posts/*`, `/categories`, `/tags`, `/settings`, `/users`, `/login-*`, `/setup`, `/logout`, `/seo/*` | `no-store` | Dashboard & admin routes tidak pernah di-cache |
| Semua route lain (fallthrough) | `no-cache` | Tidak ada cache yang diterapkan |

### Cloudflare Cache Purge

Jika fitur Cloudflare diaktifkan (`enable_cloudflare = yes` dengan
`cloudflare_api_key`, `cloudflare_zone_id`, dan `site_url` yang valid), sistem
akan otomatis **meng-purge cache berdasarkan URL** (bukan purge everything)
ketika terjadi perubahan data:

- **Create / Edit Post**: purge URL post, kategori, semua tag, dan homepage
  dalam satu request.
- **Create / Edit Category**: purge URL kategori dan homepage.
- **Create / Edit Tag**: purge URL tag dan homepage.
- **Delete Post**: purge URL post, kategori, semua tag, dan homepage.
- **Sitemap Update**: purge URL sitemap.xml.

> **Catatan**: Purge dilakukan secara asynchronous (non-blocking) sehingga
> tidak menghambat respons API. Jika konfigurasi Cloudflare tidak lengkap,
> purge akan dilewati tanpa error.

---

## 12. IndexNow Integration

**go-lightblog** mendukung protokol **IndexNow** untuk memberi tahu search engine
(Bing, Yandex, dll.) tentang URL baru atau yang berubah secara instan.

### Konfigurasi
Aktifkan melalui pengaturan:
- `indexnow` = `"yes"` untuk mengaktifkan.
- `indexnow_key` = key verifikasi (otomatis digenerate saat instalasi, 80-128 karakter alfanumerik).

### File Verifikasi
File verifikasi tersedia di: `/{indexnow_key}.txt`
- Jika key di URL tidak cocok dengan key yang dikonfigurasi, akan mengembalikan `404 Not Found`.

### Perilaku Otomatis
- **Saat IndexNow diaktifkan pertama kali**: Sistem mengirim homepage URL ke IndexNow (hanya sekali, ditandai dengan `indexnow_submitted = yes`).
- **Saat artikel dibuat/diupdate (bukan draft)**: URL artikel dikirim ke IndexNow.
- **Saat artikel dihapus**: URL artikel dikirim ke IndexNow agar search engine menghapusnya dari index.

> **Catatan**: Semua submission dilakukan secara asynchronous (non-blocking) dan membutuhkan `site_url` yang valid.

---

## Ringkasan Kode Status HTTP (HTTP Status Codes)

Berikut adalah daftar kode status HTTP yang dikembalikan oleh API ini:

| Status Code | Makna | Kondisi Terjadinya |
| :--- | :--- | :--- |
| `200 OK` | Sukses | Request GET, PUT, atau DELETE berhasil diproses. |
| `201 Created` | Berhasil Dibuat | Entitas baru (User/Post/Tag/Category) berhasil disimpan. |
| `400 Bad Request` | Permintaan Salah | Format JSON tidak valid, field wajib kosong, parameter gambar tidak valid. |
| `401 Unauthorized` | Tidak Terautentikasi | Header `Authorization` tidak dikirim atau API Key salah/tidak terdaftar. |
| `403 Forbidden` | Akses Dilarang | Pengguna dengan role `editor` mencoba mengakses endpoint manajemen user/settings. |
| `404 Not Found` | Tidak Ditemukan | ID entitas yang dicari tidak ada, atau login token URL tidak sesuai. |
| `409 Conflict` | Konflik | Username, Category Slug, atau Tag Slug sudah ada di database (Duplikat). |
| `500 Internal Server Error` | Masalah Server | Terjadi kegagalan pada database SQLite atau filesystem lokal server. |