import { api } from './api.ts'
import type { PersonForm, Person, UserResponse } from '../types/person.ts'

function toPerson(user: UserResponse): Person {
  return {
    id: user.id,
    firstName: user.firstname,
    lastName: user.lastname,
    name: `${user.firstname} ${user.lastname}`,
    birthday: user.birthday ?? '',
    address: user.address ?? '',
  }
}

export async function getUsers(): Promise<Person[]> {
  const { data } = await api.get<UserResponse[]>('/user')
  return data.map(toPerson)
}

export async function getUserById(id: Person['id']): Promise<Person> {
  const { data } = await api.get<UserResponse>(`/user/${encodeURIComponent(id)}`)
  return toPerson(data)
}

export async function createUser(person: PersonForm): Promise<Person> {
  const { data } = await api.post<UserResponse>('/user', {
    firstname: person.firstName.trim(),
    lastname: person.lastName.trim(),
    birthday: person.birthday || null,
    address: person.address?.trim() || null,
  })
  return toPerson(data)
}
