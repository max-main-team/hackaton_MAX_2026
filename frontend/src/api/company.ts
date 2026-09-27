import { request } from './client'
import type { Company, Vacancy, VacancyInput } from './types'

export function createCompany(input: {
  name: string
  description: string
  website: string
  logo_url: string
  address: string
  position: string
}): Promise<Company> {
  return request<Company>('/companies', { method: 'POST', body: JSON.stringify(input) })
}

export function myCompanies(): Promise<Company[]> {
  return request<Company[]>('/my/companies')
}

export function companyVacancies(companyId: number): Promise<Vacancy[]> {
  return request<Vacancy[]>(`/companies/${companyId}/vacancies`)
}

export function createVacancy(companyId: number, input: VacancyInput): Promise<Vacancy> {
  return request<Vacancy>(`/companies/${companyId}/vacancies`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateVacancy(
  id: number,
  patch: { response_ttl_hours?: number; is_active?: boolean },
): Promise<Vacancy> {
  return request<Vacancy>(`/vacancies/${id}`, { method: 'PATCH', body: JSON.stringify(patch) })
}
