import type { Locator, Page } from '@playwright/test'
import { test, expect, firstPerson, peopleResponse } from './fixtures'

const listRoute = '**/api/persons?*'
const createRoute = '**/api/person'
const fullName = `${firstPerson.firstname} ${firstPerson.lastname}`

async function openCreate(page: Page) {
  await page.getByRole('button', { name: 'ADD', exact: true }).click()
  return page.getByRole('dialog', { name: 'เพิ่มข้อมูล' })
}

async function fillPerson(dialog: Locator) {
  await dialog.getByRole('textbox', { name: 'ชื่อ', exact: true }).fill(' มาลี ')
  await dialog.getByRole('textbox', { name: 'นามสกุล', exact: true }).fill(' รักเรียน ')
  await dialog.getByLabel('วันเกิด', { exact: true }).fill('2000-12-25')
  await dialog.getByRole('textbox', { name: 'ที่อยู่', exact: true }).fill(' เชียงใหม่ ')
}

test.beforeEach(async ({ page, people }) => {
  // Requesting the fixture registers the API routes before navigation.
  expect(people).toHaveLength(12)
  await page.goto('/')
  await expect(page.getByRole('rowheader', { name: fullName, exact: true })).toBeVisible()
})

test('renders API records, formatted birthdays, ages, and nullable fields', async ({ page }) => {
  const row = page.getByRole('row').filter({ has: page.getByRole('rowheader', { name: fullName }) })
  await expect(row.getByRole('cell', { name: 'กรุงเทพมหานคร', exact: true })).toBeVisible()
  await expect(row.getByRole('cell', { name: '15 Jan 1995', exact: true })).toBeVisible()
  await expect(row.getByRole('cell', { name: '31', exact: true })).toBeVisible()
  const nullableRow = page.getByRole('row').filter({ hasText: 'ผู้ใช้2 ทดสอบ' })
  await expect(nullableRow.getByRole('cell', { name: '-', exact: true })).toHaveCount(2)
  await expect(page.getByRole('rowheader')).toHaveCount(10)
  await expect(page.getByText('แสดง 1–10 จาก 12 รายการ')).toBeVisible()
  // Screen-reader labels inside the wide table must not expand the mobile viewport.
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= document.documentElement.clientWidth,
    ),
  ).toBe(true)
})

test('paginates using server metadata and disables buttons at page boundaries', async ({
  page,
}) => {
  await expect(page.getByRole('button', { name: 'ก่อนหน้า' })).toBeDisabled()
  const nextRequest = page.waitForRequest('**/api/persons?page=2&limit=10')
  await page.getByRole('button', { name: 'ถัดไป' }).click()
  await nextRequest
  await expect(page.getByText('แสดง 11–12 จาก 12 รายการ')).toBeVisible()
  await expect(page.getByRole('rowheader')).toHaveText(['ผู้ใช้11 ทดสอบ', 'ผู้ใช้12 ทดสอบ'])
  await expect(page.getByRole('button', { name: 'ถัดไป' })).toBeDisabled()
  await page.getByRole('button', { name: 'ก่อนหน้า' }).click()
  await expect(page.getByText('หน้า 1 / 2')).toBeVisible()
  await expect(page.getByRole('rowheader', { name: fullName, exact: true })).toBeVisible()
})

test('shows an empty state and disables pagination when there are no users', async ({
  page,
  people,
}) => {
  people.length = 0
  await page.reload()
  await expect(page.getByText('ยังไม่มีข้อมูลผู้ใช้')).toBeVisible()
  await expect(page.getByRole('rowheader')).toHaveCount(0)
  await expect(page.getByText('แสดง 0–0 จาก 0 รายการ')).toBeVisible()
  await expect(page.getByRole('button', { name: 'ก่อนหน้า' })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'ถัดไป' })).toBeDisabled()
})

test('shows loading state and disables actions until the list arrives', async ({
  page,
  people,
}) => {
  let respond!: () => Promise<void>
  await page.route(listRoute, (route) => {
    respond = () => route.fulfill({ json: peopleResponse(people) })
  })
  await page.reload()
  await expect(page.getByRole('status')).toHaveText('กำลังโหลดข้อมูล...')
  await expect(page.getByRole('region', { name: 'รายชื่อผู้ใช้' })).toHaveAttribute(
    'aria-busy',
    'true',
  )
  await expect(page.getByRole('button', { name: 'ADD', exact: true })).toBeDisabled()
  await expect(page.getByRole('button', { name: 'ถัดไป' })).toBeDisabled()
  await respond()
  await expect(page.getByRole('status')).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'ADD', exact: true })).toBeEnabled()
})

test('displays a list error and recovers when the user retries', async ({ page }) => {
  await page.route(listRoute, (route) =>
    route.fulfill({ status: 503, json: { error: 'รายชื่อไม่พร้อมใช้งาน' } }),
  )
  await page.reload()
  await expect(page.getByRole('alert')).toContainText('รายชื่อไม่พร้อมใช้งาน')
  await expect(page.getByText('ยังไม่มีข้อมูลผู้ใช้')).toHaveCount(0)
  await page.unroute(listRoute)
  await page.getByRole('button', { name: 'ลองอีกครั้ง' }).click()
  await expect(page.getByRole('alert')).toHaveCount(0)
  await expect(page.getByRole('rowheader', { name: fullName, exact: true })).toBeVisible()
})

test('displays the fallback message for network failures', async ({ page }) => {
  await page.route(listRoute, (route) => route.abort('failed'))
  await page.reload()
  await expect(page.getByRole('alert')).toContainText('โหลดรายชื่อไม่สำเร็จ กรุณาลองอีกครั้ง')
  await expect(page.getByRole('button', { name: 'ลองอีกครั้ง' })).toBeEnabled()
})

test('returns to the last available page if a page disappears', async ({ page, people }) => {
  people.splice(10)
  await page.getByRole('button', { name: 'ถัดไป' }).click()
  await expect(page.getByText('หน้า 1 / 1')).toBeVisible()
  await expect(page.getByText('แสดง 1–10 จาก 10 รายการ')).toBeVisible()
  await expect(page.getByRole('rowheader', { name: fullName, exact: true })).toBeVisible()
  await expect(page.getByRole('button', { name: 'ถัดไป' })).toBeDisabled()
})

test('fetches fresh details and opens a read-only dialog that can be closed', async ({ page }) => {
  await page.route('**/api/person/1', (route) =>
    route.fulfill({ json: { ...firstPerson, address: 'ที่อยู่ล่าสุดจากเซิร์ฟเวอร์' } }),
  )
  await page.getByRole('button', { name: `View data for ${fullName}`, exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'ข้อมูลผู้ใช้' })
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('textbox', { name: 'ชื่อ', exact: true })).toHaveValue('สมชาย')
  await expect(dialog.getByRole('textbox', { name: 'นามสกุล', exact: true })).toHaveValue('ใจดี')
  await expect(dialog.getByLabel('วันเกิด', { exact: true })).toHaveValue('1995-01-15')
  await expect(dialog.getByRole('textbox', { name: 'ที่อยู่' })).toHaveValue(
    'ที่อยู่ล่าสุดจากเซิร์ฟเวอร์',
  )
  for (const field of await dialog.locator('input, textarea').all()) {
    await expect(field).toBeDisabled()
  }
  await dialog.getByRole('button', { name: 'ปิด', exact: true }).click()
  await expect(dialog).not.toBeVisible()
})

test('reports detail errors without opening the dialog and allows another attempt', async ({
  page,
}) => {
  await page.route('**/api/person/1', (route) =>
    route.fulfill({ status: 404, json: { error: 'ไม่พบข้อมูลผู้ใช้' } }),
  )
  const view = page.getByRole('button', { name: `View data for ${fullName}`, exact: true })
  await view.click()
  await expect(page.getByRole('alert')).toHaveText('ไม่พบข้อมูลผู้ใช้')
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect(view).toBeEnabled()
  await page.unroute('**/api/person/1')
  await view.click()
  await expect(page.getByRole('dialog', { name: 'ข้อมูลผู้ใช้' })).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})

test('creates a user with trimmed API fields and refreshes the table', async ({ page }) => {
  const dialog = await openCreate(page)
  await fillPerson(dialog)
  await expect(dialog.getByRole('status', { name: 'อายุ', exact: true })).toHaveText(/^25\s*ปี$/)
  const request = page.waitForRequest((request) => request.method() === 'POST')
  await dialog.getByRole('button', { name: 'บันทึก', exact: true }).click()
  expect((await request).postDataJSON()).toEqual({
    firstname: 'มาลี',
    lastname: 'รักเรียน',
    birthday: '2000-12-25',
    address: 'เชียงใหม่',
  })
  await expect(dialog).not.toBeVisible()
  await expect(page.getByRole('rowheader', { name: 'มาลี รักเรียน', exact: true })).toBeVisible()
  await expect(page.getByText('แสดง 1–10 จาก 13 รายการ')).toBeVisible()
})

test('validates required fields and rejects future birthdays before sending a request', async ({
  page,
}) => {
  const submissions: string[] = []
  page.on('request', (request) => {
    if (request.method() === 'POST') submissions.push(request.url())
  })
  const dialog = await openCreate(page)
  const save = dialog.getByRole('button', { name: 'บันทึก', exact: true })
  await save.click()
  await expect(dialog.locator('input:invalid')).toHaveCount(3)
  await expect(dialog.getByRole('textbox', { name: 'ชื่อ', exact: true })).toBeFocused()
  await fillPerson(dialog)
  const birthday = dialog.getByLabel('วันเกิด', { exact: true })
  await expect(birthday).toHaveAttribute('max', '2026-10-22')
  await birthday.fill('2026-10-23')
  await save.click()
  await expect(birthday).toBeFocused()
  await expect(dialog.locator('input:invalid')).toHaveCount(1)
  await expect(dialog).toBeVisible()
  expect(submissions).toEqual([])
})

test('cancels an add form with the button or Escape and resets its contents', async ({ page }) => {
  const submissions: string[] = []
  page.on('request', (request) => {
    if (request.method() === 'POST') submissions.push(request.url())
  })
  const dialog = await openCreate(page)
  await fillPerson(dialog)
  await dialog.getByRole('button', { name: 'ยกเลิก', exact: true }).click()
  await expect(dialog).not.toBeVisible()
  await openCreate(page)
  for (const field of await dialog.locator('input, textarea').all()) {
    await expect(field).toHaveValue('')
  }
  await page.keyboard.press('Escape')
  await expect(dialog).not.toBeVisible()
  expect(submissions).toEqual([])
})

test('keeps failed form values for retry and prevents closing or duplicate saves while saving', async ({
  page,
}) => {
  await page.route(createRoute, (route) =>
    route.fulfill({ status: 503, json: { error: 'บันทึกไม่สำเร็จ กรุณาลองอีกครั้ง' } }),
  )
  const dialog = await openCreate(page)
  await fillPerson(dialog)
  await dialog.getByRole('button', { name: 'บันทึก', exact: true }).click()
  await expect(dialog.getByRole('alert')).toHaveText('บันทึกไม่สำเร็จ กรุณาลองอีกครั้ง')
  await expect(dialog.getByRole('textbox', { name: 'ชื่อ', exact: true })).toHaveValue('มาลี')
  await expect(dialog.getByRole('textbox', { name: 'นามสกุล', exact: true })).toHaveValue(
    'รักเรียน',
  )
  await expect(dialog.getByLabel('วันเกิด', { exact: true })).toHaveValue('2000-12-25')
  await expect(dialog.getByRole('textbox', { name: 'ที่อยู่' })).toHaveValue('เชียงใหม่')

  await page.unroute(createRoute)
  let submissions = 0
  let finishSave!: () => Promise<void>
  await page.route(createRoute, (route) => {
    submissions++
    finishSave = () => route.fallback()
  })
  await dialog.getByRole('button', { name: 'บันทึก', exact: true }).click()
  await expect(dialog.getByRole('button', { name: 'กำลังบันทึก...', exact: true })).toBeDisabled()
  await expect(dialog.getByRole('button', { name: 'ยกเลิก', exact: true })).toBeDisabled()
  for (const field of await dialog.locator('input, textarea').all()) {
    await expect(field).toBeDisabled()
  }
  await page.keyboard.press('Escape')
  await page.keyboard.press('Enter')
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('alert')).toHaveCount(0)
  await expect.poll(() => submissions).toBe(1)
  await finishSave()
  await expect(dialog).not.toBeVisible()
  await expect(page.getByRole('rowheader', { name: 'มาลี รักเรียน', exact: true })).toBeVisible()
  expect(submissions).toBe(1)
})
