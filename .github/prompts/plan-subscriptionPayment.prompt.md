# Plan: Subscription & Payment Module

Implementasi modul langganan berbasis seat dengan integrasi Xendit.

## Business Rules Confirmed

- **Grace Period**: 7 hari (past_due sebelum expired)
- **Downgrade**: Efektif di periode berikutnya
- **Seat Limit**: Block employee creation saat max_seats tercapai
- **No proration**: Mid-cycle upgrade = full price untuk periode baru
- **Checkout validation**: `seat_count` harus >= jumlah active employees

## Steps

### 1. Create database migration
Create `internal/infrastructure/database/postgresql/migrations/000007_subscription_system.up.sql` with:
- Enum types: `subscription_status`, `invoice_status`, `billing_cycle_enum`
- Tables: `features`, `subscription_plans`, `plan_features`, `subscriptions`, `invoices`
- Indexes for performance
- Seed data for features and plans

### 2. Create domain layer
Create `internal/domain/subscription/` with:
- `entity.go` - Feature, Plan, PlanFeature, Subscription, Invoice structs
- `dto.go` - CheckoutRequest, UpgradeRequest, WebhookPayload, responses
- `service.go` - Service interface
- `repository.go` - Repository interfaces (Feature, Plan, Subscription, Invoice)
- `errors.go` - ErrSubscriptionNotFound, ErrInsufficientSeats, ErrFeatureNotAllowed, etc.

### 3. Add Xendit config
Add `XenditConfig` struct to `internal/config/config.go`:
- APIKey, WebhookSecret, BaseURL, Environment (sandbox/production)

### 4. Create Xendit package
Create `internal/pkg/xendit/` with:
- `client.go` - HTTP client wrapper
- `invoice.go` - CreateInvoice, GetInvoice functions
- `webhook.go` - Signature verification

### 5. Create repository implementation
Create `internal/repository/postgresql/subscription.go` with:
- FeatureRepository: GetByCode, List
- PlanRepository: GetByID, GetByName, ListActive, GetFeaturesByPlanID
- SubscriptionRepository: GetByCompanyID, Create, Update, UpdateStatus, ListExpiring
- InvoiceRepository: Create, GetByXenditID, UpdateStatus, ListPending
- Helper: CountActiveEmployees(companyID)

### 6. Create subscription service
Create `internal/service/subscription/service.go` with:
- `GetPlans()` - List available plans with features
- `GetMySubscription()` - Get company subscription with features
- `Checkout()` - Validate seats >= active employees, create Xendit invoice
- `HandleWebhook()` - Process Xendit callback, update subscription
- `UpgradePlan()` - Immediate upgrade, full price new period
- `DowngradePlan()` - Set pending_plan_id, apply at next period
- `CancelSubscription()` - Set status cancelled, access until period_end
- `UpdateExpiredSubscriptions()` - For cron: active→past_due→expired

### 7. Modify auth service for trial
Update `internal/service/auth/service.go`:
- In `Register()`: Auto-create subscription with plan_id="Free Trial", 14 days, status=trial
- In JWT generation: Add claims `features[]` and `subscription_expires_at`

### 8. Create subscription middleware
Create `internal/handler/http/middleware/subscription.go`:
- `RequireActiveSubscription` - Check status != expired, current_period_end > now
- `RequireFeature(code string)` - Check feature in JWT claims + verify in DB

### 9. Create HTTP handler
Create `internal/handler/http/subscription.go` with endpoints:
- `GET /plans` - List plans (public)
- `GET /subscription/my` - Get current subscription (auth)
- `POST /subscription/checkout` - Create checkout invoice (owner)
- `POST /subscription/upgrade` - Upgrade plan (owner)
- `POST /subscription/downgrade` - Downgrade plan (owner)
- `POST /subscription/cancel` - Cancel subscription (owner)
- `POST /webhook/xendit` - Webhook receiver (public, verified signature)

### 10. Update router and main.go
- Add subscription routes to `internal/handler/http/router.go`
- Wire dependencies in `cmd/api/main.go`

### 11. Create cron package
Create `internal/cron/` with:
- `scheduler.go` - Cron scheduler setup
- `subscription.go` - Job: update active→past_due (7 days grace), past_due→expired
- `invoice.go` - Job: cleanup stale pending invoices (>24h)

### 12. Add seat limit check
Modify `internal/service/employee/service.go`:
- In `CreateEmployee()`: Check current count < subscription.max_seats
- Return `ErrMaxSeatsReached` if limit exceeded

### 13. Build verification
Run `go build ./...` to verify compilation

## Key Validation Logic

### Checkout Flow
```go
func (s *Service) Checkout(ctx context.Context, req CheckoutRequest) (*Invoice, error) {
    // 1. Get active employee count
    activeCount, err := s.employeeRepo.CountActive(ctx, companyID)
    if err != nil {
        return nil, err
    }
    
    // 2. Validate seat count
    if req.SeatCount < activeCount {
        return nil, fmt.Errorf("%w: minimum %d seats required", 
            ErrInsufficientSeats, activeCount)
    }
    
    // 3. Get plan and calculate amount
    plan, err := s.planRepo.GetByID(ctx, req.PlanID)
    amount := plan.PricePerSeat * decimal.NewFromInt(int64(req.SeatCount))
    
    // 4. Create Xendit invoice
    xenditInvoice, err := s.xendit.CreateInvoice(...)
    
    // 5. Save invoice with snapshot
    invoice := Invoice{
        Amount: amount,
        PlanSnapshotName: plan.Name,
        PricePerSeatSnapshot: plan.PricePerSeat,
        SeatCountSnapshot: req.SeatCount,
        ...
    }
    return s.invoiceRepo.Create(ctx, invoice)
}
```

### Cron: Status Update
```go
func (s *Service) UpdateExpiredSubscriptions(ctx context.Context) error {
    now := time.Now()
    gracePeriod := 7 * 24 * time.Hour
    
    // 1. Trial/Active → Past Due (period ended)
    s.repo.UpdateStatus(ctx, 
        "current_period_end < $1 AND status IN ('trial', 'active')",
        now, "past_due")
    
    // 2. Past Due → Expired (grace period ended)
    s.repo.UpdateStatus(ctx,
        "current_period_end < $1 AND status = 'past_due'",
        now.Add(-gracePeriod), "expired")
    
    return nil
}
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/plans | Public | List all active plans |
| GET | /api/v1/subscription/my | Auth | Get current subscription |
| POST | /api/v1/subscription/checkout | Owner | Create payment invoice |
| POST | /api/v1/subscription/upgrade | Owner | Upgrade to higher plan |
| POST | /api/v1/subscription/downgrade | Owner | Downgrade (next period) |
| POST | /api/v1/subscription/cancel | Owner | Cancel subscription |
| POST | /api/v1/webhook/xendit | Public* | Xendit callback (*signature verified) |

## Database Schema Summary

```
features (id, code, name, description)
    ↓
plan_features (plan_id, feature_id) ← pivot
    ↓
subscription_plans (id, name, price_per_seat, tier_level, max_seats)
    ↓
subscriptions (id, company_id, plan_id, status, max_seats, period_start, period_end, pending_plan_id)
    ↓
invoices (id, company_id, subscription_id, xendit_*, amount, snapshots, status, paid_at)
```


Send email when subscription expires, payment received, etc.


Berikut adalah Rangkuman Master (Master Blueprint) dari seluruh diskusi kita mengenai arsitektur Backend HRIS untuk modul Subscription & Payment.

Rancangan ini disusun dengan standar Enterprise-Grade, mengutamakan keamanan data, integritas keuangan, dan skalabilitas.

I. Ringkasan Arsitektur Sistem

Kita telah menyepakati 5 Pilar Utama dalam sistem ini:

1. Database: Pendekatan Relasional (Option 2)

Kita tidak menyimpan fitur di dalam satu kolom JSONB. Kita memecahnya menjadi tabel terpisah (plan_features) yang ternormalisasi.

Alasan: Lebih rapi, mencegah typo data, memungkinkan pembuatan API Public Pricing Page yang dinamis, dan memudahkan analitik penggunaan fitur.

2. Strategi Trial: "Free Trial" adalah Plan

"Free Trial" bukan sekadar status, melainkan sebuah Paket (Plan) resmi di database dengan harga Rp 0.

Mekanisme: Saat Register, user otomatis dibuatkan subscription dengan plan_id milik "Free Trial" dan durasi 14 hari. Fitur dibatasi lewat tabel relasi.

3. Keamanan: Middleware "Double Protection"

Middleware bertindak sebagai satpam yang tidak hanya percaya pada status database, tapi melakukan cek ulang secara real-time via JWT.

Cek 1 (Fitur): Memastikan user memiliki hak akses ke fitur spesifik (misal: payroll).

Cek 2 (Waktu): Memastikan token belum melewati batas waktu langganan (current_period_end), melindungi sistem dari keterlambatan update status oleh Cron Job.

4. Keuangan: Snapshot Data & Integrasi Xendit

Tabel Invoice dirancang sebagai Audit Trail. Kita menggunakan teknik Snapshot: menyimpan harga dan nama paket saat transaksi terjadi.

Tujuan: Jika tahun depan harga paket naik, data histori transaksi lama tidak ikut berubah (tetap akurat sesuai saat pembelian).

5. Otomasi: Cron Jobs

Karena database tidak bisa berubah sendiri, kita menggunakan Cron Job (Penjadwal) untuk:

Mengubah status active -> past_due -> expired.

Membersihkan invoice Xendit yang menggantung (pending basi).

II. Skema Database Final (PostgreSQL)

Silakan gunakan skema ini. Skema ini sudah mencakup tabel Master, Transaksi, dan Integrasi Xendit.

A. Setup & Enums

SQL


-- Mengaktifkan UUIDCREATE EXTENSION IF NOT EXISTS "uuid-ossp";-- 1. Status Langganan-- 'trial': User baru (14 hari).-- 'active': User berbayar.-- 'past_due': Telat bayar (Grace Period).-- 'cancelled': Stop langganan.-- 'expired': Akses mati total.CREATE TYPE subscription_status AS ENUM ('trial', 'active', 'past_due', 'cancelled', 'expired');-- 2. Status Invoice (Mapping Xendit)CREATE TYPE invoice_status AS ENUM ('pending', 'paid', 'expired', 'failed');-- 3. Siklus TagihanCREATE TYPE billing_cycle_enum AS ENUM ('monthly', 'yearly');

B. Tabel Master Data (Fitur & Paket)

SQL


-- Tabel A: Kamus Fitur (Daftar Menu)-- Ini referensi utama untuk kodingan Middleware GoCREATE TABLE features (

    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    code VARCHAR(50) NOT NULL UNIQUE,  -- Key Coding: 'payroll', 'gps_attendance'

    name VARCHAR(100) NOT NULL,        -- Label UI: 'Absensi GPS'

    description TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW()

);-- Tabel B: Paket Langganan-- 'Free Trial' masuk sini dengan price = 0CREATE TABLE subscription_plans (

    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    name VARCHAR(50) NOT NULL UNIQUE,       -- 'Free Trial', 'Standard', 'Premium'

    price_per_seat DECIMAL(15,2) NOT NULL,  -- Harga per karyawan

    tier_level SMALLINT NOT NULL DEFAULT 0, -- 0=Trial, 1=Basic, 2=Pro

    is_active BOOLEAN DEFAULT true,

    created_at TIMESTAMPTZ DEFAULT NOW()

);-- Tabel C: Pivot / Rules (Inti dari Option 2)-- Menentukan paket mana dapat fitur apaCREATE TABLE plan_features (

    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,

    feature_id UUID NOT NULL REFERENCES features(id) ON DELETE CASCADE,

    is_active BOOLEAN DEFAULT true,

    PRIMARY KEY (plan_id, feature_id)

);

C. Tabel Transaksi (Subscription & Invoice)

SQL


-- Tabel D: Status Langganan Perusahaan (Jantung Sistem)-- Dicek setiap Login untuk generate JWTCREATE TABLE subscriptions (

    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE, 

    plan_id UUID NOT NULL REFERENCES subscription_plans(id),

    

    status subscription_status NOT NULL DEFAULT 'trial',

    

    -- Limit Karyawan (Seat-Based)

    max_seats INT NOT NULL DEFAULT 1, 

    

    -- "The Grim Reaper Fields"

    -- Middleware Logic: Block jika NOW() > current_period_end

    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    current_period_end TIMESTAMPTZ NOT NULL, 

    trial_ends_at TIMESTAMPTZ, -- Penanda visual kapan trial habis

    

    billing_cycle billing_cycle_enum NOT NULL DEFAULT 'monthly',

    created_at TIMESTAMPTZ DEFAULT NOW(),

    updated_at TIMESTAMPTZ DEFAULT NOW(),

    

    CONSTRAINT uk_company_subscription UNIQUE (company_id)

);CREATE INDEX idx_subscriptions_company_id ON subscriptions(company_id);-- Tabel E: Log Invoice & Integrasi Xendit-- Menyimpan Snapshot KeuanganCREATE TABLE invoices (

    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    company_id UUID NOT NULL REFERENCES companies(id),

    subscription_id UUID NOT NULL REFERENCES subscriptions(id),

    

    -- Kolom Integrasi Xendit

    xendit_invoice_id VARCHAR(255), -- ID untuk Webhook matching

    xendit_invoice_url TEXT,        -- Link bayar user

    xendit_expiry_date TIMESTAMPTZ, -- Kapan link mati

    

    -- SNAPSHOT DATA (Data Immutable)

    -- Wajib simpan nilai "Saat Transaksi" agar tahan inflasi harga paket

    amount DECIMAL(15,2) NOT NULL,            -- Total Rupiah (Harga * Qty)

    plan_snapshot_name VARCHAR(50) NOT NULL,  -- Nama paket saat beli (misal: "Premium")

    price_per_seat_snapshot DECIMAL(15,2) NOT NULL, -- Harga saat beli (misal: 12000)

    seat_count_snapshot INT NOT NULL,         -- Jumlah kursi saat beli

    

    status invoice_status NOT NULL DEFAULT 'pending',

    

    -- Data Pelunasan (Diisi Webhook)

    issue_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    paid_at TIMESTAMPTZ,

    payment_method VARCHAR(50),  

    payment_channel VARCHAR(50), 

    

    created_at TIMESTAMPTZ DEFAULT NOW(),

    updated_at TIMESTAMPTZ DEFAULT NOW()

);CREATE INDEX idx_invoices_xendit_id ON invoices(xendit_invoice_id);

III. Data Seeding (Wajib Dijalankan)

Script ini mengatur logika bisnis: Siapa dapat fitur apa.

SQL


-- 1. Input Fitur (Kamus)INSERT INTO features (code, name) VALUES 

('attendance', 'Attendance System'),

('payroll', 'Payroll System');

('invitation', 'Invite Employee');

('leave', 'Leave System');



-- 2. Input Paket (Termasuk Trial)INSERT INTO subscription_plans (name, price_per_seat, tier_level) VALUES 

('Free Trial', 0, 0),

('Standard', 12000, 1),

('Premium', 15000, 2),

('Ultra', 20000, 3);-- 3. Aturan Main (Mapping)-- A. User Trial: Cuma dapat Absensi WebINSERT INTO plan_features (plan_id, feature_id)SELECT p.id, f.id FROM subscription_plans p, features fWHERE p.name = 'Free Trial' AND f.code IN ('attendance');-- B. User Premium: Dapat SemuaINSERT INTO plan_features (plan_id, feature_id)SELECT p.id, f.id FROM subscription_plans p, features fWHERE p.name = 'Premium' AND f.code IN ('attendance', 'attendance', 'payroll');

IV. Alur Kerja Sistem (Flowchart Verbal)

Register:

Backend membuat Company & User.

Backend mencari Plan "Free Trial".

Backend membuat Subscription: status='trial', plan_id=Trial, end=Hari ini+14.

Login (Generate Token):

Backend query JOIN: subscriptions -> plan_features -> features.

Hasil fitur (['attendance']) dan waktu expired dimasukkan ke JWT.

Akses API (Middleware):

User akses /payroll.

Middleware cek JWT: Apakah ada payroll? -> Tidak ada. -> 403 Forbidden.

Upgrade (Checkout):

User beli Premium (10 Seat).

Backend buat Invoice (Snapshot harga 12.000).

Backend panggil Xendit.

Pembayaran (Webhook):

Xendit lapor "PAID".

Backend update Invoice jadi paid.

Backend update Subscription: Ganti plan_id ke Premium, perpanjang current_period_end.

Cron Job (Tengah Malam):

Cek user yang current_period_end < NOW().

Update status jadi past_due atau expired.

Ini adalah panduan lengkap untuk membangun backend HRIS kamu. Tinggal implementasi ke kode Go!