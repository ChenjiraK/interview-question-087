import assert from 'node:assert/strict'
import { afterEach, test } from 'node:test'
import type { InternalAxiosRequestConfig } from 'axios'
import { api, getApiErrorMessage } from './api.ts'
import { createUser, getUserById, getUsers } from './users.ts'

const originalAdapter = api.defaults.adapter
const user = {
  id: 42,
  firstname: 'สมชาย',
  lastname: 'ใจดี',
  birthday: '1995-01-15',
  address: 'กรุงเทพมหานคร',
}

function respondWith(
  data: unknown,
  inspect: (config: InternalAxiosRequestConfig) => void = () => {},
) {
  api.defaults.adapter = async (config) => {
    inspect(config)
    return { data, status: 200, statusText: 'OK', headers: {}, config }
  }
}

afterEach(() => {
  api.defaults.adapter = originalAdapter
})

test('GET /user maps backend fields to table and form fields', async () => {
  respondWith([user], (config) => {
    assert.equal(config.method, 'get')
    assert.equal(config.url, '/user')
  })
  assert.deepEqual(await getUsers(), [
    {
      id: 42,
      firstName: 'สมชาย',
      lastName: 'ใจดี',
      name: 'สมชาย ใจดี',
      birthday: '1995-01-15',
      address: 'กรุงเทพมหานคร',
    },
  ])
})

test('GET /user handles an empty list', async () => {
  respondWith([])
  assert.deepEqual(await getUsers(), [])
})

test('GET /user/:id fetches details and handles nullable fields', async () => {
  respondWith({ ...user, birthday: null, address: null }, (config) => {
    assert.equal(config.method, 'get')
    assert.equal(config.url, '/user/42')
  })
  const person = await getUserById(42)
  assert.equal(person.firstName, user.firstname)
  assert.equal(person.lastName, user.lastname)
  assert.equal(person.birthday, '')
  assert.equal(person.address, '')
})

test('POST /user sends only supported fields and uses the server ID', async () => {
  respondWith(user, (config) => {
    assert.equal(config.method, 'post')
    assert.equal(config.url, '/user')
    assert.deepEqual(JSON.parse(config.data), {
      firstname: 'สมชาย',
      lastname: 'ใจดี',
      birthday: '1995-01-15',
      address: 'กรุงเทพมหานคร',
    })
  })
  const input = {
    id: 1,
    firstName: ' สมชาย ',
    lastName: ' ใจดี ',
    birthday: user.birthday,
    address: user.address,
    name: 'ignored',
    age: 31,
  }
  const person = await createUser(input)
  assert.equal(person.id, 42)
})

test('POST /user converts empty birthday and address to null for the API', async () => {
  respondWith(user, (config) => {
    assert.deepEqual(JSON.parse(config.data), {
      firstname: 'สมชาย',
      lastname: 'ใจดี',
      birthday: null,
      address: null,
    })
  })
  await createUser({ firstName: 'สมชาย', lastName: 'ใจดี', birthday: '', address: '' })
})

test('request errors reach callers with backend error or a network fallback', async () => {
  const backendError = { response: { status: 404, data: { error: 'user not found' } } }
  api.defaults.adapter = async () => {
    throw backendError
  }
  await assert.rejects(getUserById(42), (error) => error === backendError)
  assert.equal(getApiErrorMessage(backendError, 'fallback'), 'user not found')
  assert.equal(getApiErrorMessage(new Error('Network Error'), 'fallback'), 'fallback')
})
