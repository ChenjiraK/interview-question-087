# interview-question-087

แอปจัดการข้อมูลบุคคล แสดงรายการแบบแบ่งหน้า เพิ่มข้อมูล และดูรายละเอียด โดยใช้ Go (Gin + GORM) เป็น Backend, PostgreSQL เป็นฐานข้อมูล และ Vue 3 + TypeScript + Vite เป็น Frontend

## สิ่งที่ต้องติดตั้ง

- Go 1.25 ขึ้นไป
- Node.js 22.18 ขึ้นไป และ npm
- Docker พร้อม Docker Compose v2 และเปิด Docker ให้พร้อมใช้งาน
- `make` สำหรับคำสั่ง Backend

## โครงสร้างโปรเจกต์

```text
interview-question-087/
├── back-go/       # REST API, PostgreSQL, migrations และ Go tests
└── front-vue/     # Vue application, service tests และ Playwright tests
```

คำสั่งด้านล่างเริ่มจากโฟลเดอร์หลักของโปรเจกต์ ยกเว้นที่ระบุเป็นอย่างอื่น โดยเปิด Backend และ Frontend คนละ Terminal

## Run Backend

เปิด Terminal แรก แล้วรัน:

```bash
cd back-go
go mod download
make migrate-up
make run
```

`make migrate-up` จะสร้างไฟล์ `.env` จาก `.env.example` ถ้ายังไม่มี เปิด PostgreSQL ผ่าน Docker รอฐานข้อมูลพร้อม และรัน migrations ส่วน `make run` จะเริ่ม API ที่ `http://localhost:8080`

หากต้องการเปลี่ยนค่าฐานข้อมูลก่อนเริ่มครั้งแรก ให้คัดลอกไฟล์แล้วแก้ค่าใน `back-go/.env` ก่อนรัน `make migrate-up`:

```bash
# รันจาก back-go เฉพาะเมื่อยังไม่มี .env
cp .env.example .env
```

ค่าเริ่มต้นสำหรับพัฒนาในเครื่อง:

| ตัวแปร | ค่าเริ่มต้น |
| --- | --- |
| `POSTGRES_HOST` | `127.0.0.1` |
| `POSTGRES_PORT` | `5432` |
| `POSTGRES_DB` | `interview-question-087` |
| `POSTGRES_USER` | `user` |
| `POSTGRES_PASSWORD` | `password` |
| `POSTGRES_SSLMODE` | `disable` |
| `PORT` | `8080` (ถ้าไม่กำหนด) |

Backend อ่าน `.env` อัตโนมัติ โดย environment ของ process มีลำดับความสำคัญสูงกว่า และสามารถใช้ `DATABASE_URL` แทนค่าการเชื่อมต่อ `POSTGRES_*` ทั้งหมดได้

ตรวจสอบ API จากอีก Terminal:

```bash
curl http://localhost:8080/health
# ผลลัพธ์: {"status":"ok"}

curl 'http://localhost:8080/persons?page=1&limit=10'
```

หากพอร์ตฐานข้อมูล `5432` ถูกใช้อยู่ ให้เปลี่ยน `POSTGRES_PORT` ใน `.env` เช่น `5433` แล้วเริ่มฐานข้อมูลใหม่ หากต้องการเปลี่ยนพอร์ต API ให้รัน `PORT=8081 make run` และปรับค่า proxy ของ Frontend ให้ตรงกัน

## Run Frontend

เปิด Terminal ที่สองจากโฟลเดอร์หลัก:

```bash
cd front-vue
npm ci
# คัดลอกเฉพาะครั้งแรก ถ้ายังไม่มี .env.local
cp .env.example .env.local
npm run dev
```

เปิด `http://localhost:5173` หรือ URL ที่ Vite แสดงใน Terminal หากพอร์ตเริ่มต้นไม่ว่าง ระหว่างใช้งานให้เปิด Backend และ PostgreSQL ไว้ด้วย

ค่าเริ่มต้นใน `front-vue/.env.local`:

```dotenv
VITE_API_BASE_URL=/api
VITE_API_PROXY_TARGET=http://localhost:8080
```

Vite ส่ง request ที่ขึ้นต้นด้วย `/api` ไปยัง Backend และตัด `/api` ออก เช่น `/api/persons` จะถูกส่งไปที่ `http://localhost:8080/persons` หาก Backend ใช้พอร์ตอื่น ให้แก้ `VITE_API_PROXY_TARGET` แล้วรีสตาร์ต Vite

## Run Tests

### Backend: Unit tests

จากโฟลเดอร์หลัก:

```bash
cd back-go
go mod download
make test
```

คำสั่งนี้เรียก `go test ./...` ทดสอบ handler, service, repository, model และ HTTP routes โดยไม่ต้องเปิด Backend หรือ PostgreSQL หากไม่ได้กำหนด `TEST_DATABASE_URL` จะข้าม integration test ที่ใช้ฐานข้อมูลจริง

ตัวเลือกเพิ่มเติม รันจาก `back-go`:

```bash
go test ./... -v                     # แสดงรายละเอียดแต่ละ test
go test ./... -cover                 # แสดง coverage
go test ./internal/service -v        # ทดสอบเฉพาะ service
go test -race ./...                  # ตรวจ data race (ต้องมี C compiler ที่ Go รองรับ)
make vet                            # ตรวจโค้ดด้วย go vet
```

### Backend: Integration test กับ PostgreSQL จริง

ต้องใช้ฐานข้อมูลทดสอบแยกที่รัน migrations แล้วและตาราง `person` ว่าง เนื่องจาก test จะเพิ่มข้อมูล ตรวจสอบผล และลบข้อมูลที่สร้างเมื่อจบ โดยไม่ต้องเปิด API server แยก

ตัวอย่างด้านล่างรันจาก `back-go` ใช้ Docker Compose project และ volume แยกจากฐานข้อมูลพัฒนา โดยพอร์ต `5433` ต้องว่าง:

```bash
(
  set -e
  export COMPOSE_PROJECT_NAME=interview-087-test
  export POSTGRES_DB=interview-question-087-test
  export POSTGRES_USER=user
  export POSTGRES_PASSWORD=password
  export POSTGRES_PORT=5433

  make migrate-up
  TEST_DATABASE_URL='postgres://user:password@127.0.0.1:5433/interview-question-087-test?sslmode=disable' \
    go test ./internal/httpapi -run '^TestPersonsPostgres$' -v -count=1
)
```

ตัวแปรที่ export ในวงเล็บมีผลเฉพาะชุดคำสั่งนี้ หากใช้ฐานข้อมูลทดสอบที่เตรียมไว้เอง ให้กำหนด `TEST_DATABASE_URL` ให้ตรงกับฐานข้อมูลนั้น แล้วรันคำสั่ง `go test` ข้างต้น

หยุด containers ของฐานข้อมูลทดสอบหลังใช้งาน โดยรันจาก `back-go`:

```bash
COMPOSE_PROJECT_NAME=interview-087-test \
POSTGRES_DB=interview-question-087-test \
POSTGRES_USER=user POSTGRES_PASSWORD=password POSTGRES_PORT=5433 \
  make db-down
```

### Frontend: Service unit tests

จากโฟลเดอร์หลัก:

```bash
cd front-vue
npm ci
npm test
```

ทดสอบ service ที่เรียก API การแปลงข้อมูล pagination และข้อผิดพลาด โดยจำลอง request จึงไม่ต้องเปิด Backend หรือฐานข้อมูล

### Frontend: End-to-end tests ด้วย Playwright

จาก `front-vue` หลังติดตั้ง dependencies แล้ว:

```bash
npx playwright install chromium
npm run test:e2e
```

Playwright เปิดและปิด Vite server ที่ `http://127.0.0.1:4173` ให้เอง ต้องให้พอร์ต `4173` ว่าง โดยไม่ต้องรัน `npm run dev` เพิ่ม ชุดทดสอบจำลอง API และไม่ต้องเปิด Backend หรือ PostgreSQL จึงตรวจการทำงานของ UI แต่ไม่ได้ตรวจการเชื่อมต่อฐานข้อมูลจริง

ทดสอบทั้ง Desktop Chromium และ Mobile Chromium ครอบคลุมรายการ การแบ่งหน้า การเพิ่มข้อมูล validation รายละเอียด และการแสดงข้อผิดพลาด

```bash
npm run test:e2e -- --project=chromium  # เฉพาะ Desktop
npm run test:e2e:ui                    # เปิด interactive test runner
npm run test:e2e:report                # เปิด HTML report ของการรันล่าสุด
```

รายงานอยู่ที่ `front-vue/playwright-report/` ส่วน screenshot และ trace ของ test ที่ล้มเหลวอยู่ที่ `front-vue/test-results/`

สำหรับ CI ใช้ `npm ci`, `npx playwright install --with-deps chromium` และ `npm run test:e2e` ตามลำดับ

## Build และตรวจ TypeScript

รันจาก `back-go`:

```bash
make build
./bin/server
```

ก่อนเปิด executable ต้องเตรียม `.env`, เปิด PostgreSQL และรัน migrations เช่นเดียวกับการรัน Backend ในโหมดพัฒนา

รันจาก `front-vue`:

```bash
npm run type-check
npm run build
```

`npm run build` ตรวจ TypeScript ก่อนสร้างไฟล์ production ใน `front-vue/dist/` โดย Vite proxy ใช้เฉพาะ development หากนำ build ไป deploy ต้องตั้ง reverse proxy ให้ส่ง `/api/*` ไป Backend โดยตัด `/api` ออก รายละเอียดเพิ่มเติมอยู่ใน [Frontend README](front-vue/README.md)

## หยุดการทำงาน

กด `Ctrl+C` ใน Terminal ที่รัน Backend และ Frontend จากนั้นหยุดฐานข้อมูลพัฒนา:

```bash
# รันจากโฟลเดอร์หลัก
cd back-go
make db-down
```

`make db-down` เก็บข้อมูลใน Docker volume ไว้ เมื่อต้องการเริ่มใหม่ให้รัน `make migrate-up` และ `make run` อีกครั้ง

## เอกสารเพิ่มเติม

- [Backend README](back-go/README.md)
- [API Specification](back-go/API_SPEC.md)
- [Frontend README](front-vue/README.md)
