export interface PersonResponse {
  id: number
  firstname: string
  lastname: string
  birthday: string | null
  address: string | null
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: Pagination
}

export interface PersonForm {
  firstName: string
  lastName: string
  birthday: string
  address: string
}

export interface Person extends PersonForm {
  id: number
  name: string
}

export interface PersonSubmission extends PersonForm {
  name: string
  age: number | null
}
