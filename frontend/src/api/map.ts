import { request } from './client'
import type { MapVacancy } from './types'

export function mapVacancies(): Promise<MapVacancy[]> {
  return request<MapVacancy[]>('/vacancies/map')
}
