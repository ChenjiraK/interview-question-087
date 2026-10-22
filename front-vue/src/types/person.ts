export interface UserResponse {
  id: number
  firstname: string
  lastname: string
  birthday: string | null
  address: string | null
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
