import { test as base, expect } from '@playwright/test'
import type { PersonResponse } from '../src/types/person'

export const firstPerson: PersonResponse = {
  id: 1,
  firstname: 'สมชาย',
  lastname: 'ใจดี',
  birthday: '1995-01-15',
  address: 'กรุงเทพมหานคร',
}

export function peopleResponse(people: PersonResponse[], page = 1, limit = 10) {
  return {
    data: people.slice((page - 1) * limit, page * limit),
    pagination: {
      page,
      limit,
      total: people.length,
      total_pages: Math.ceil(people.length / limit),
    },
  }
}

// Each test receives its own in-memory API; no backend or shared database is required.
export const test = base.extend<{ people: PersonResponse[] }>({
  people: async ({ page }, use) => {
    const people = [
      firstPerson,
      ...Array.from({ length: 11 }, (_, index) => ({
        id: index + 2,
        firstname: `ผู้ใช้${index + 2}`,
        lastname: 'ทดสอบ',
        birthday: null,
        address: null,
      })),
    ]

    await page.clock.setFixedTime(new Date('2026-10-22T12:00:00+07:00'))
    await page.route('**/api/**', async (route) => {
      const request = route.request()
      const url = new URL(request.url())

      if (request.method() === 'GET' && url.pathname === '/api/persons') {
        const currentPage = Number(url.searchParams.get('page'))
        const limit = Number(url.searchParams.get('limit'))
        await route.fulfill({ json: peopleResponse(people, currentPage, limit) })
      } else if (request.method() === 'GET' && /^\/api\/person\/\d+$/.test(url.pathname)) {
        const person = people.find((person) => person.id === Number(url.pathname.split('/').pop()))
        await route.fulfill({
          status: person ? 200 : 404,
          json: person ?? { error: 'Person not found' },
        })
      } else if (request.method() === 'POST' && url.pathname === '/api/person') {
        const person: PersonResponse = {
          ...request.postDataJSON(),
          id: Math.max(0, ...people.map((person) => person.id)) + 1,
        }
        people.unshift(person)
        await route.fulfill({ status: 201, json: person })
      } else {
        throw new Error(`Unexpected API request: ${request.method()} ${url.pathname}`)
      }
    })

    await use(people)
  },
})

export { expect }
