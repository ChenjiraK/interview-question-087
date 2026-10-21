# back-go

โปรเจกต์ REST API ภาษา Go โดยใช้ [Gin](https://gin-gonic.com/) พร้อม middleware สำหรับบันทึก request และกู้คืนจาก panic

## สิ่งที่ต้องมี

- Go 1.25 ขึ้นไป

## เริ่มใช้งาน

รันคำสั่งจากโฟลเดอร์ `back-go`:

```bash
go mod download
go run .
```

เซิร์ฟเวอร์จะเปิดที่ `http://localhost:8080` และหยุดได้ด้วย `Ctrl+C` โดยรอ request ที่กำลังทำงานสูงสุด 5 วินาที

เปลี่ยนพอร์ตหรือเปิด release mode ผ่าน environment variables:

```bash
PORT=3000 GIN_MODE=release go run .
```

## Endpoints

| Method | Path | Response |
| --- | --- | --- |
| GET | `/health` | `{"status":"ok"}` |
| GET | `/ping` | `{"message":"pong"}` |
| GET | `/api/v1/ping` | `{"message":"pong"}` |

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ping
```

Path ที่ไม่มีอยู่จะตอบ HTTP 404 พร้อม JSON `{"error":"route not found"}`

## โครงสร้าง

```text
back-go/
├── main.go                       # เริ่มเซิร์ฟเวอร์และ graceful shutdown
├── main_test.go                  # ทดสอบ endpoints ผ่าน entry point
├── internal/httpapi/
│   ├── router.go                 # Routes และ middleware
│   └── router_test.go            # ทดสอบ HTTP responses
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
