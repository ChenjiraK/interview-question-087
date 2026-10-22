import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env?.VITE_API_BASE_URL || '/api',
  timeout: 10000,
})

export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (typeof error !== 'object' || error === null || !('response' in error)) return fallback
  const response = error.response
  if (typeof response !== 'object' || response === null || !('data' in response)) return fallback
  const data = response.data
  if (typeof data !== 'object' || data === null || !('error' in data)) return fallback
  const message = data.error
  return typeof message === 'string' && message ? message : fallback
}
