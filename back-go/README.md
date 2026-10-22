# back-go

โปรเจกต์ REST API ภาษา Go โดยใช้ [Gin](https://gin-gonic.com/) สำหรับ HTTP และ [GORM](https://gorm.io/) สำหรับ PostgreSQL พร้อม middleware สำหรับบันทึก request และกู้คืนจาก panic

เอกสาร request/response, validation, status codes และตัวอย่าง curl: [API Specification](API_SPEC.md)

## สิ่งที่ต้องมี

- Go 1.25 ขึ้นไป
- Docker พร้อม Docker Compose v2 ขึ้นไป สำหรับ PostgreSQL และ migrations

## เริ่มใช้งาน

รันคำสั่งจากโฟลเดอร์ `back-go`:

```bash
go mod download
make migrate-up
go run .
```

เซิร์ฟเวอร์จะเปิดที่ `http://localhost:8080` และหยุดได้ด้วย `Ctrl+C` โดยรอ request ที่กำลังทำงานสูงสุด 5 วินาที

แอปอ่าน `.env` อัตโนมัติ โดยค่าจาก environment ของ process มีลำดับความสำคัญสูงกว่า ใช้ `POSTGRES_HOST` (ค่าเริ่มต้น `127.0.0.1`), `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` และ `POSTGRES_SSLMODE` (ค่าเริ่มต้น `disable` สำหรับพัฒนาในเครื่อง) หรือกำหนด `DATABASE_URL` เพื่อใช้ connection URL แทนทั้งหมด แอปตรวจการเชื่อมต่อ PostgreSQL ก่อนเริ่มรับ request

เปลี่ยนพอร์ตหรือเปิด release mode ผ่าน environment variables:

```bash
PORT=3000 GIN_MODE=release go run .
```

## PostgreSQL และ migrations

ใช้ PostgreSQL 18 ผ่าน Docker Compose และ [golang-migrate](https://github.com/golang-migrate/migrate) ผ่าน container จึงไม่ต้องติดตั้ง migration CLI บนเครื่อง

```bash
make migrate-up
```

คำสั่งนี้จะสร้าง `.env` จาก `.env.example` ถ้ายังไม่มี เปิดฐานข้อมูล รอให้พร้อม แล้วรัน migration ที่ยังไม่เคยรัน การเรียกซ้ำจะไม่สร้างตารางซ้ำ

ค่าเริ่มต้นสำหรับพัฒนาในเครื่อง:

| ค่า | ค่าเริ่มต้น |
| --- | --- |
| Host | `localhost` |
| Port | `5432` |
| Database | `interview-question-087` |
| Username | `user` |
| Password | `password` |

หากต้องการแก้ค่าก่อนเริ่มครั้งแรก ให้รัน `cp .env.example .env` แล้วแก้ไฟล์ `.env` ซึ่งจะไม่ถูกเก็บใน Git หากพอร์ต 5432 ถูกใช้อยู่ ให้เปลี่ยน `POSTGRES_PORT` เช่น `5433`

ข้อมูลถูกเก็บใน Docker volume `postgres_data` จึงยังอยู่หลัง `make db-down` และเปิดใหม่ ตัวแปร `POSTGRES_DB`, `POSTGRES_USER` และ `POSTGRES_PASSWORD` ใช้สร้างฐานข้อมูลตอน volume ว่างครั้งแรก การแก้ `.env` ภายหลังจะไม่เปลี่ยนชื่อฐานข้อมูลหรือบัญชีใน volume เดิม

Migration `000001_create_user` สร้างตารางเดิม `public."user"` จากนั้น `000002_rename_user_to_person` เปลี่ยนชื่อเป็น `public.person` โดยเก็บข้อมูล คอลัมน์ และ ID เดิมไว้:

สำหรับฐานข้อมูลที่มีอยู่แล้ว ให้รัน `make migrate-up` เพื่ออัปเกรดเป็น v2 แล้วรีสตาร์ต backend เพื่อโหลด routes และ model ชุดใหม่

| Column | Type | เงื่อนไข |
| --- | --- | --- |
| `id` | `BIGINT` | Primary key สร้างเลขอัตโนมัติ |
| `firstname` | `TEXT` | ห้ามเป็น NULL |
| `lastname` | `TEXT` | ห้ามเป็น NULL |
| `birthday` | `DATE` | เป็น NULL ได้ |
| `address` | `TEXT` | เป็น NULL ได้ |
| `created_at` | `TIMESTAMPTZ` | ห้ามเป็น NULL และกำหนดเวลาสร้างให้อัตโนมัติ |
| `updated_at` | `TIMESTAMPTZ` | ห้ามเป็น NULL และกำหนดค่าเริ่มต้นเป็นเวลาปัจจุบัน |

ทั้งสองคอลัมน์ใช้ค่าเริ่มต้น `CURRENT_TIMESTAMP` เมื่อเพิ่มแถวใหม่ ส่วน migration v1 ปัจจุบันไม่ได้สร้าง trigger สำหรับเปลี่ยน `updated_at` จึงต้องกำหนดค่าใหม่ในคำสั่ง UPDATE เช่น `SET updated_at = CURRENT_TIMESTAMP`

เปิด SQL shell และทดลองเพิ่มข้อมูลในตาราง `person` ได้ดังนี้:

```bash
make db-shell
```

```sql
INSERT INTO public.person (firstname, lastname, birthday, address)
VALUES ('Somchai', 'Jaidee', '1995-01-15', 'Bangkok')
RETURNING *;

SELECT * FROM public.person;
```

คำสั่งอื่น:

```bash
make db-up            # เปิด PostgreSQL และรอให้พร้อม
make migrate-version  # ดูเวอร์ชัน migration ปัจจุบัน
make migrate-down     # ย้อนกลับ 1 migration: v2 เปลี่ยนชื่อ person กลับเป็น user โดยเก็บข้อมูล
make db-down          # หยุด containers โดยเก็บ volume ไว้
```

เพิ่ม migration ถัดไปเป็นคู่ไฟล์ เช่น `migrations/000003_add_email.up.sql` และ `migrations/000003_add_email.down.sql` แล้วรัน `make migrate-up` โดยไม่แก้ migration ที่รันไปแล้ว

Go API ใช้ GORM กับตาราง `public.person` สำหรับสร้างบุคคลและอ่านข้อมูล โดยรัน SQL migrations ให้เรียบร้อยก่อนเปิดเซิร์ฟเวอร์ โครงสร้างฐานข้อมูลยังจัดการผ่านไฟล์ migration เดิม แอปไม่เรียก `AutoMigrate`

## Endpoints

| Method | Path | Response |
| --- | --- | --- |
| GET | `/health` | `{"status":"ok"}` |
| GET | `/ping` | `{"message":"pong"}` |
| GET | `/api/v1/ping` | `{"message":"pong"}` |
| POST | `/person` | `201` พร้อมข้อมูลบุคคลที่บันทึกแล้ว |
| GET | `/persons` | `200` พร้อม `data` และ `pagination` เรียงตาม `id` |
| GET | `/person/{id}` | `200` พร้อมข้อมูลบุคคล หรือ `404` เมื่อไม่พบ |

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ping
```

Path ที่ไม่มีอยู่จะตอบ HTTP 404 พร้อม JSON `{"error":"route not found"}`

### สร้างบุคคล

```bash
curl -i -X POST http://localhost:8080/person \
  -H 'Content-Type: application/json' \
  -d '{"firstname":"Somchai","lastname":"Jaidee","birthday":"1995-01-15","address":"Bangkok"}'
```

`firstname` และ `lastname` ต้องเป็นข้อความที่ไม่ว่าง ระบบตัดช่องว่างหัวท้ายให้อัตโนมัติ ส่วน `birthday` ต้องเป็นวันที่จริงในรูปแบบ `YYYY-MM-DD` และ `address` เป็นข้อความ ทั้งสองช่องนี้เว้นไว้หรือส่ง `null` ได้

ฐานข้อมูลสร้าง `id`, `created_at` และ `updated_at` ให้ ไม่ต้องส่งใน request หากส่ง field ที่ไม่รองรับหรือ JSON ไม่ถูกต้องจะตอบ `400` และจำกัด request body ที่ 1 MiB (เกินจะตอบ `413`)

ตัวอย่าง response (`201 Created` พร้อม header `Location: /person/1`):

```json
{
  "id": 1,
  "firstname": "Somchai",
  "lastname": "Jaidee",
  "birthday": "1995-01-15",
  "address": "Bangkok",
  "created_at": "2026-10-21T16:00:00Z",
  "updated_at": "2026-10-21T16:00:00Z"
}
```

### อ่านข้อมูลบุคคล

```bash
curl 'http://localhost:8080/persons?page=1&limit=10'
curl http://localhost:8080/person/1
```

`GET /persons` รับ `page` (ค่าเริ่มต้น 1) และ `limit` (ค่าเริ่มต้น 10, สูงสุด 100) โดยส่ง `{ "data": [...], "pagination": { "page": 1, "limit": 10, "total": 25, "total_pages": 3 } }` เมื่อไม่มีข้อมูลหรือเกินหน้าสุดท้าย `data` จะเป็น `[]` และเมื่อไม่มีข้อมูลทั้งหมด `total_pages` เป็น 0 หาก query ไม่ใช่จำนวนเต็มบวก, `limit` เกิน 100 หรือ offset เกินช่วง int จะตอบ `400` ส่วน `GET /person/{id}` ต้องระบุ ID เป็นจำนวนเต็มบวก หากรูปแบบผิดจะตอบ `400` หากไม่พบจะตอบ `404` พร้อม `{"error":"person not found"}` ส่วนข้อผิดพลาดจากฐานข้อมูลจะตอบ `500` โดยไม่ส่งรายละเอียดภายในกลับไป

## โครงสร้าง

เส้นทางการทำงานคือ `Gin router → handler → service → repository → PostgreSQL` โดยประกอบ dependencies ใน `main.go`:

- **Handler** รับ JSON/URL parameter กำหนด timeout และแปลงผลลัพธ์หรือข้อผิดพลาดเป็น HTTP response
- **Service** ตรวจข้อมูลที่จำเป็น ตัดช่องว่างชื่อ ตรวจวันเกิด/ID และแปลง model เป็น response ที่วันเกิดเป็น `YYYY-MM-DD`
- **Repository** ใช้ GORM สร้างและอ่านข้อมูล พร้อมแปลงกรณีไม่พบแถวเป็น `ErrPersonNotFound`
- **Model** ผูกข้อมูลกับตาราง `public.person` และให้ฐานข้อมูลสร้าง ID และ timestamps ตาม SQL migration

Service พึ่งพา repository ผ่าน interface จึงทดสอบกฎการทำงานได้โดยไม่ต้องเปิดฐานข้อมูล

```text
back-go/
├── main.go                       # เริ่มเซิร์ฟเวอร์และ graceful shutdown
├── API_SPEC.md                   # API specification
├── compose.yaml                  # PostgreSQL และ migration runner
├── .env.example                  # ตัวอย่างค่าฐานข้อมูลสำหรับพัฒนา
├── migrations/
│   ├── 000001_create_user.up.sql
│   ├── 000001_create_user.down.sql
│   ├── 000002_rename_user_to_person.up.sql
│   └── 000002_rename_user_to_person.down.sql
├── internal/
│   ├── database/postgres.go      # เปิด GORM และตั้งค่า connection pool
│   ├── handler/person.go           # HTTP request/response
│   ├── service/person.go           # Validation และการทำงานของบุคคล
│   ├── repository/person.go        # สร้าง/อ่านข้อมูลด้วย GORM
│   ├── model/person.go             # GORM model สำหรับตาราง person
│   └── httpapi/
│       ├── router.go             # Routes และ middleware
│       ├── router_test.go
│       ├── persons_test.go
│       └── persons_integration_test.go
├── go.mod
├── go.sum
└── Makefile
```

## คำสั่งสำหรับพัฒนา

```bash
make run     # รันเซิร์ฟเวอร์
make test    # ทดสอบ
make vet     # ตรวจโค้ด
make fmt     # จัดรูปแบบโค้ด
make build   # สร้าง executable ที่ bin/server
```

รัน executable หลัง build:

```bash
./bin/server
```

### Unit tests

มี unit test แยกในแต่ละ layer โดยไม่ต้องเปิด PostgreSQL:

| Package | สิ่งที่ทดสอบ |
| --- | --- |
| `internal/handler` | ใช้ mock service ตรวจ JSON, query/ID, status และ response, Location header, timeout และการแปลง error |
| `internal/service` | ใช้ mock repository ตรวจ validation, การตัดช่องว่างชื่อ, วันเกิด, pagination/overflow, การแปลง response และส่งต่อ error/context |
| `internal/repository` | ใช้ `go-sqlmock` กับ GORM PostgreSQL ตรวจ SQL/arguments, transaction/rollback, NULL, ค่าที่ฐานข้อมูลสร้าง, pagination และ error |
| `internal/model` | ตรวจ GORM schema: ชื่อตาราง/คอลัมน์, primary key, ชนิดวันที่ และ timestamps ที่สร้างโดยฐานข้อมูล |

```bash
go test ./...                         # รันเทสต์ทั้งหมด
go test ./internal/service -v         # รันเฉพาะ service พร้อมแสดงแต่ละกรณี
go test ./internal/handler ./internal/service ./internal/repository ./internal/model -cover
go test -race ./...                   # ตรวจ data race
```

ทดสอบกับ PostgreSQL จริงได้โดยกำหนด `TEST_DATABASE_URL` ไปยังฐานข้อมูลทดสอบแยกที่รัน migration แล้วและตาราง person ว่าง จากนั้นรัน `go test ./internal/httpapi -run TestPersonsPostgres -v` หากไม่กำหนดตัวแปรนี้ integration test จะถูกข้าม โดย unit tests ปกติไม่ต้องเชื่อมฐานข้อมูล
