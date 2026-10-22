# API Specification — Person API

เอกสารนี้อธิบาย API ที่มีอยู่ในโปรเจกต์ `back-go` ตามพฤติกรรมของ handler, service และ repository ปัจจุบัน

## 1. ข้อมูลทั่วไป

| รายการ | รายละเอียด |
| --- | --- |
| Base URL สำหรับพัฒนา | `http://localhost:8080` |
| รูปแบบ request body | JSON |
| รูปแบบ response | JSON; `Content-Type: application/json; charset=utf-8` |
| Authentication | เรียกใช้งานได้โดยไม่ต้องส่ง token หรือ API key |
| Person API path | `/person`, `/persons` และ `/person/{id}` |

หากกำหนด environment variable `PORT` ให้ใช้พอร์ตนั้นแทน `8080` โดยพอร์ตนี้เป็นพอร์ต HTTP API ส่วนพอร์ต PostgreSQL เป็นค่าการเชื่อมต่อฐานข้อมูลแยกต่างหาก

เริ่มใช้งานจากโฟลเดอร์โปรเจกต์:

```bash
make migrate-up
make run
```

## 2. สรุป endpoints

| Method | Path | การทำงาน | Success status |
| --- | --- | --- | --- |
| POST | `/person` | สร้างบุคคล | `201 Created` |
| GET | `/persons` | อ่านบุคคลแบบแบ่งหน้า เรียงตาม ID จากน้อยไปมาก | `200 OK` |
| GET | `/person/{id}` | อ่านบุคคลตาม ID | `200 OK` |
| GET | `/health` | ตรวจว่าแอปรับ request ได้ | `200 OK` |
| GET | `/ping` | ตรวจการตอบกลับของแอป | `200 OK` |
| GET | `/api/v1/ping` | ตรวจการตอบกลับผ่าน route group | `200 OK` |

## 3. โครงสร้างข้อมูล Person

| Field | JSON type | Nullable | รายละเอียด |
| --- | --- | --- | --- |
| `id` | integer (int64) | ไม่ได้ | Primary key ที่ฐานข้อมูลสร้างให้ เป็นจำนวนเต็มบวก |
| `firstname` | string | ไม่ได้ | ชื่อที่ตัดช่องว่างหัวท้ายแล้ว |
| `lastname` | string | ไม่ได้ | นามสกุลที่ตัดช่องว่างหัวท้ายแล้ว |
| `birthday` | string | ได้ | วันเกิดรูปแบบ `YYYY-MM-DD` |
| `address` | string | ได้ | ที่อยู่ |
| `created_at` | string (date-time) | ไม่ได้ | เวลาสร้างข้อมูล รูปแบบ RFC 3339 พร้อม timezone และอาจมีเศษวินาที |
| `updated_at` | string (date-time) | ไม่ได้ | เวลาในคอลัมน์ updated_at รูปแบบเดียวกับ created_at |

`id`, `created_at` และ `updated_at` เป็นข้อมูลที่อ่านจากฐานข้อมูล ไม่รับเป็น field ใน request สร้างบุคคล ทั้งสอง timestamps มีค่าเริ่มต้นจากเวลาปัจจุบันของฐานข้อมูลเมื่อสร้างแถว

ตัวอย่าง Person:

```json
{
  "id": 1,
  "firstname": "สมชาย",
  "lastname": "ใจดี",
  "birthday": "1995-01-15",
  "address": "กรุงเทพมหานคร",
  "created_at": "2026-10-21T16:00:00Z",
  "updated_at": "2026-10-21T16:00:00Z"
}
```

ค่า ID และ timestamps ในเอกสารเป็นตัวอย่าง ค่าจริงมาจากฐานข้อมูล และ ID อาจไม่ต่อเนื่อง

## 4. POST /person — สร้างบุคคล

### Request

```http
POST /person
Content-Type: application/json
```

| Field | Type | Required | Validation |
| --- | --- | --- | --- |
| `firstname` | string | ใช่ | ต้องมีข้อความเหลือหลังตัดช่องว่างหัวท้าย |
| `lastname` | string | ใช่ | ต้องมีข้อความเหลือหลังตัดช่องว่างหัวท้าย |
| `birthday` | string หรือ null | ไม่ | หากเป็น string ต้องเป็นวันที่จริงในรูปแบบ `YYYY-MM-DD` เช่น `2000-02-29` |
| `address` | string หรือ null | ไม่ | รับข้อความ รวมถึง string ว่าง |

กติกาเพิ่มเติม:

- หากไม่ส่ง `birthday` หรือ `address` จะเก็บเป็น NULL และตอบกลับเป็น JSON `null`
- `birthday` ที่เป็น string ว่างหรือมีส่วนเวลา เช่น `1995-01-15T00:00:00Z` จะไม่ผ่าน validation
- ไม่รับ field นอกเหนือจาก 4 fields ข้างต้น เช่น `id`, `created_at`, `updated_at` หรือ `email`
- Request ต้องเป็น JSON object เดียว หากมี JSON อีกก้อนหรือข้อมูลที่ parse ไม่ได้ต่อท้ายจะตอบ `400`
- จำกัด request body ที่ 1 MiB หรือ 1,048,576 bytes
- การเรียกแต่ละครั้งจะสร้างบุคคลใหม่ โดยไม่ตรวจชื่อ–นามสกุลซ้ำ

ตัวอย่าง request แบบครบ:

```json
{
  "firstname": "สมชาย",
  "lastname": "ใจดี",
  "birthday": "1995-01-15",
  "address": "กรุงเทพมหานคร"
}
```

ตัวอย่าง request ขั้นต่ำ:

```json
{
  "firstname": "สมชาย",
  "lastname": "ใจดี"
}
```

### Success response

Status: `201 Created`

Response headers:

```http
Content-Type: application/json; charset=utf-8
Location: /person/1
```

Response body เป็น Person object ตามหัวข้อ 3 โดย `Location` ใช้ ID ของแถวที่เพิ่งสร้าง เช่น response ของ request ขั้นต่ำ:

```json
{
  "id": 1,
  "firstname": "สมชาย",
  "lastname": "ใจดี",
  "birthday": null,
  "address": null,
  "created_at": "2026-10-21T16:00:00Z",
  "updated_at": "2026-10-21T16:00:00Z"
}
```

### Error responses

| Status | `error` | กรณี |
| --- | --- | --- |
| `400 Bad Request` | `invalid JSON body or unknown fields` | JSON ผิดรูปแบบ, field ไม่รองรับ, ชนิดข้อมูลไม่ตรง หรือมี JSON หลายก้อน |
| `400 Bad Request` | `firstname and lastname are required` | ไม่ส่งชื่อ/นามสกุล, ส่ง null หรือเป็นข้อความว่างหลังตัดช่องว่าง |
| `400 Bad Request` | `birthday must be a valid date in YYYY-MM-DD format` | วันเกิดผิดรูปแบบหรือเป็นวันที่ไม่มีจริง |
| `413 Request Entity Too Large` | `request body must not exceed 1 MiB` | อ่าน request body แล้วพบว่าเกินขนาดที่กำหนด |
| `500 Internal Server Error` | `internal server error` | บันทึกข้อมูลไม่สำเร็จหรือฐานข้อมูลมีข้อผิดพลาด |

ตัวอย่าง curl:

```bash
curl -i -X POST http://localhost:8080/person \
  -H 'Content-Type: application/json' \
  -d '{"firstname":"สมชาย","lastname":"ใจดี","birthday":"1995-01-15","address":"กรุงเทพมหานคร"}'
```

## 5. GET /persons — อ่านบุคคลแบบแบ่งหน้า

### Request

```http
GET /persons?page=1&limit=10
```

ไม่มี request body ระบบคืนบุคคลในหน้าที่ระบุ เรียงตาม `id` จากน้อยไปมาก โดยแบ่งหน้าที่ฐานข้อมูลด้วย LIMIT/OFFSET

| Query parameter | Type | Default | Validation |
| --- | --- | --- | --- |
| `page` | integer | `1` | จำนวนเต็มบวก และ `(page - 1) * limit` ต้องไม่เกินช่วง int ของเซิร์ฟเวอร์ |
| `limit` | integer | `10` | จำนวนเต็มระหว่าง 1 ถึง 100 |

ถ้าไม่ส่ง query parameter จะใช้ค่าเริ่มต้น แต่ส่งค่าว่าง เช่น `?page=` จะตอบ `400` ไม่มี filter หรือการเปลี่ยนลำดับ

### Success response

Status: `200 OK`

Response เป็น object ที่มี `data` เป็น array ของ Person object และ `pagination` เป็นข้อมูลแบ่งหน้า (เปลี่ยนจาก response เดิมที่เป็น array โดยตรง):

```json
{
  "data": [
    {
      "id": 1,
      "firstname": "สมชาย",
      "lastname": "ใจดี",
      "birthday": "1995-01-15",
      "address": "กรุงเทพมหานคร",
      "created_at": "2026-10-21T16:00:00Z",
      "updated_at": "2026-10-21T16:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

`total` คือจำนวนบุคคลทั้งหมด และ `total_pages` คือจำนวนหน้าทั้งหมดที่คำนวณจาก `total` และ `limit` โดยปัดขึ้น

กรณีไม่มีข้อมูลยังคงตอบ `200 OK`:

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 0,
    "total_pages": 0
  }
}
```

หากขอหน้าที่เกินหน้าสุดท้าย จะตอบ `200` พร้อม `data: []` โดยยังคืน `page` ที่ขอ และ `total`/`total_pages` ตามข้อมูลทั้งหมด

### Error responses

| Status | `error` | กรณี |
| --- | --- | --- |
| `400 Bad Request` | `page must be a positive integer` | page ไม่ใช่จำนวนเต็มบวกหรือเกินช่วง int |
| `400 Bad Request` | `page is too large` | offset ที่คำนวณได้เกินช่วง int |
| `400 Bad Request` | `limit must be an integer between 1 and 100` | limit ไม่ใช่จำนวนเต็มระหว่าง 1 ถึง 100 |
| `500 Internal Server Error` | `internal server error` | นับหรืออ่านข้อมูลจากฐานข้อมูลไม่สำเร็จ |

ตัวอย่าง curl:

```bash
curl -i 'http://localhost:8080/persons?page=2&limit=10'
```

## 6. GET /person/{id} — อ่านบุคคลตาม ID

### Request

```http
GET /person/1
```

| Path parameter | Type | Required | Validation |
| --- | --- | --- | --- |
| `id` | integer (int64) | ใช่ | ต้องเป็นจำนวนเต็มบวก ไม่เกิน `9223372036854775807` |

ตัวอย่าง ID ที่ไม่ผ่าน validation ได้แก่ `0`, `-1`, `abc`, `1.5` และค่าที่เกินช่วง int64

### Success response

Status: `200 OK`

Response เป็น Person object ตามหัวข้อ 3:

```json
{
  "id": 1,
  "firstname": "สมชาย",
  "lastname": "ใจดี",
  "birthday": "1995-01-15",
  "address": "กรุงเทพมหานคร",
  "created_at": "2026-10-21T16:00:00Z",
  "updated_at": "2026-10-21T16:00:00Z"
}
```

### Error responses

| Status | `error` | กรณี |
| --- | --- | --- |
| `400 Bad Request` | `id must be a positive integer` | ID ไม่ใช่จำนวนเต็มบวกหรือเกินช่วง int64 |
| `404 Not Found` | `person not found` | ID ถูกต้องแต่ไม่มีบุคคลนั้นในฐานข้อมูล |
| `500 Internal Server Error` | `internal server error` | อ่านข้อมูลจากฐานข้อมูลไม่สำเร็จ |

ตัวอย่าง curl:

```bash
curl -i http://localhost:8080/person/1
```

## 7. System endpoints

### GET /health

Status: `200 OK`

```json
{
  "status": "ok"
}
```

Endpoint นี้ตอบจากแอปโดยไม่มีการ query ฐานข้อมูล จึงใช้ตรวจว่าแอปรับ request ได้ ส่วนการตรวจการเชื่อมต่อฐานข้อมูลเกิดขึ้นตอนเริ่มเซิร์ฟเวอร์

### GET /ping และ GET /api/v1/ping

ทั้งสอง endpoints ตอบ `200 OK` พร้อม body เดียวกัน:

```json
{
  "message": "pong"
}
```

ตัวอย่าง curl:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ping
curl -i http://localhost:8080/api/v1/ping
```

## 8. รูปแบบ error response

ข้อผิดพลาดที่ handler จัดการจะตอบ JSON object ที่มี field `error` เป็นข้อความ เช่น:

```json
{
  "error": "person not found"
}
```

ตัวอย่างการเรียก route ที่ไม่มีอยู่ เช่น `GET /not-found` จะตอบ `404 Not Found`:

```json
{
  "error": "route not found"
}
```

การเรียก service และ repository ของ Person API ใช้ context timeout 5 วินาที หากการอ่าน/เขียนฐานข้อมูลล้มเหลวรวมถึงกรณี timeout จะตอบ `500` พร้อม `internal server error`

## 9. โค้ดอ้างอิง

- [Routes](internal/httpapi/router.go)
- [HTTP handler และ error responses](internal/handler/person.go)
- [Service, validation และ response schema](internal/service/person.go)
- [GORM repository และลำดับการอ่านข้อมูล](internal/repository/person.go)
- [Model](internal/model/person.go)
- [Migration v1](migrations/000001_create_user.up.sql)
- [Migration v2 — เปลี่ยนชื่อตารางเป็น person](migrations/000002_rename_user_to_person.up.sql)
