# Hệ Thống Quản Lý Dịch Vụ Công

Hệ thống quản lý dịch vụ công trực tuyến cho phép công dân đăng ký, theo dõi và quản lý các dịch vụ công một cách hiệu quả.

## 📋 Tổng Quan

Dự án này xây dựng một nền tảng quản lý dịch vụ công hoàn chỉnh với các tính năng:

- **Quản lý người dùng**: Công dân, nhân viên, quản lý, admin
- **Quản lý dịch vụ**: Đăng ký dịch vụ công trực tuyến
- **Theo dõi hồ sơ**: Theo dõi trạng thái đơn đăng ký thời gian thực
- **Quản lý phòng ban**: Phân chia bộ phận xử lý dịch vụ
- **Thông báo**: Thông báo tự động cho công dân
- **Nhật ký hoạt động**: Ghi lại tất cả các hoạt động trong hệ thống

## 🛠 Công Nghệ

- **Backend**: Go (Golang) với Echo Web Framework
- **Database**: PostgreSQL
- **ORM**: GORM
- **Validation**: Validator v10
- **Logging**: Structured logging

## 📦 Cấu Trúc Dự Án

```bash
golang_baoan_thao/
├── cmd/
│   ├── app/              # Ứng dụng chính
│   └── seed/             # Công cụ seed dữ liệu
├── internal/
│   ├── configs/          # Cấu hình ứng dụng
│   ├── models/           # Database models
│   ├── seeder/           # Database seeder
│   └── ...
├── .env_sample           # File mẫu biến môi trường
├── go.mod                # Go module
└── README.md             # Tài liệu này
```

## 🚀 Bắt Đầu Nhanh

### Yêu Cầu

- Go 1.21 hoặc cao hơn
- PostgreSQL 12 hoặc cao hơn
- Git

### Cài Đặt

1. **Clone repository**

   ```bash
   git clone <repo-url>
   cd golang_baoan_thao
   ```

1. **Cài đặt dependencies**

   ```bash
   go mod download
   ```

1. **Cấu hình môi trường**

   Sao chép `.env_sample` thành `.env` và cập nhật cấu hình:

   ```bash
   cp .env_sample .env
   ```

   Nội dung `.env`:

   ```bash
   # Database Configuration
   DATABASE_URL=host=localhost user=postgres password=123456 dbname='Service Management System' port=5432 sslmode=disable

   # Seeding - Set to 'true' to seed database on startup
   SEED_DB=true
   ```

1. **Chạy ứng dụng**

   **Tùy chọn A - Chạy bình thường:**

   ```bash
   go run cmd/app/main.go
   ```

   **Tùy chọn B - Với Hot Reload (Khuyến nghị):**

   Trước tiên, cài đặt Air:

   ```bash
   go install github.com/air-verse/air@latest
   ```

   Sau đó chạy với Air:

   ```bash
   air
   ```

   Ứng dụng sẽ chạy trên `http://localhost:8080`

   **Lưu ý**: Khi sử dụng Air, ứng dụng sẽ tự động reload khi bạn thay đổi code

## 🐳 Chạy với Docker Compose

### Khởi động toàn bộ hệ thống (App + PostgreSQL)

```bash
docker compose up -d
```

### Dừng hệ thống

```bash
docker compose down
```

### Xem logs

```bash
docker compose logs -f app
```

### Xóa hoàn toàn (bao gồm database)

```bash
docker compose down -v
```

---

## 🛠️ Phát Triển với Hot Reload

### Cài đặt Air

```bash
go install github.com/air-verse/air@latest
```

### Chạy ứng dụng với Auto Reload

**Linux/Mac:**

```bash
make dev
```

**Windows (PowerShell):**

```powershell
.\dev.ps1
```

Air sẽ:

- ✅ Tự động biên dịch khi có thay đổi file `.go`
- ✅ Tự động restart ứng dụng
- ✅ Giám sát các thư mục: `cmd/`, `internal/`
- ✅ Ghi lỗi biên dịch vào file `build-errors.log`

### Tính năng của Air

- **Delay**: 1 giây (chờ sau khi phát hiện thay đổi trước khi rebuild)
- **Exclude**: Bỏ qua `testdata/`, `vendor/`, file `*_test.go`
- **Log**: `build-errors.log` - xem lỗi biên dịch nếu có

Cấu hình chi tiết xem tại `.air.toml`

## 🌱 Database Seeding

Chạy **một lần** sau khi database đã khởi động:

```bash
go run cmd/seed/main.go
```

Hoặc xây dựng binary:

```bash
go build -o seed cmd/seed/main.go
./seed
```

### Dữ Liệu Được Tạo

Seeder tự động tạo:

- **6 người dùng**: 1 Admin, 1 Manager, 2 Staff, 2 Citizen
- **3 phòng ban**: PGTT, PDXCG, PGPLX
- **4 dịch vụ công**: CCCD, Lái xe hạng A, Lái xe hạng C, Đăng ký xe
- **Hồ sơ công dân & nhân viên**: Với dữ liệu liên kết thực tế
- **3 đơn đăng ký**: Ở các trạng thái khác nhau
- **Thông báo mẫu**: Cho các kịch bản khác nhau

### Thông Tin Đăng Nhập Test

```text
Email: admin@example.com          | Password: admin123 | Role: Super Admin
Email: manager@example.com        | Password: admin123 | Role: Manager
Email: staff1@example.com         | Password: admin123 | Role: Staff
Email: staff2@example.com         | Password: admin123 | Role: Staff
Email: citizen1@example.com       | Password: admin123 | Role: Citizen
Email: citizen2@example.com       | Password: admin123 | Role: Citizen
```
