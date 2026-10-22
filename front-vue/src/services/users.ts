import { api } from './api.ts'
import type { PersonForm, Person, PersonResponse, PaginatedResponse } from '../types/person.ts'

function toPerson(user: PersonResponse): Person {
  return {
    id: user.id,
    firstName: user.firstname,
    lastName: user.lastname,
    name: `${user.firstname} ${user.lastname}`,
    birthday: user.birthday ?? '',
    address: user.address ?? '',
  }
}

export async function getUsers(page = 1, limit = 10): Promise<PaginatedResponse<Person>> {
  const { data } = await api.get<PaginatedResponse<PersonResponse>>('/persons', {
    params: { page, limit },
  })
  return { data: data.data.map(toPerson), pagination: data.pagination }
}

export async function getUserById(id: Person['id']): Promise<Person> {
  const { data } = await api.get<PersonResponse>(`/person/${encodeURIComponent(id)}`)
  return toPerson(data)
}

export async function createUser(person: PersonForm): Promise<Person> {
  const { data } = await api.post<PersonResponse>('/person', {
    firstname: person.firstName.trim(),
    lastname: person.lastName.trim(),
    birthday: person.birthday || null,
    address: person.address?.trim() || null,
  })
  return toPerson(data)
}
