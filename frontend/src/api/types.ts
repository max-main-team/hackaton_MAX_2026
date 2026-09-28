export interface User {
  id: number
  username: string
  first_name: string
  last_name: string
  photo_url: string
  language_code: string
  role: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
  referral_code: string
}

export interface MeResponse {
  user: User
  personal_data_accepted_at: string | null
  referral_code: string
}

export interface ResumeLink {
  type: string
  url: string
}

export interface ResumeInput {
  title: string
  skills: string
  experience_months: number
  about: string
  education: string
  links: ResumeLink[]
  city: string
  work_format: string
  employment_type: string
  salary_min: number | null
  salary_max: number | null
}

export interface ParseResumeResult {
  resume: Resume
  ai_comment: string
}

export interface ParseResumeResult {
  resume: Resume
  ai_comment: string
}

export interface Resume extends ResumeInput {
  id: number
  user_id: number
  source: string
  source_text: string
  is_active: boolean
  last_confirmed_at: string
  created_at: string
  updated_at: string
}

export interface CompanyInput {
  name: string
  description: string
  website: string
  logo_url: string
  address: string
  position: string
}

export interface Company extends CompanyInput {
  id: number
  verified: boolean
  bot_user_id: number | null
  bot_username: string
  created_at: string
  updated_at: string
}

export interface VacancyInput {
  title: string
  description: string
  required_skills: string
  min_experience_months: number
  city: string
  work_format: string
  employment_type: string
  salary_min: number | null
  salary_max: number | null
  response_ttl_hours: number
}

export interface Vacancy extends VacancyInput {
  id: number
  company_id: number
  lat: number | null
  lng: number | null
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface MapVacancy {
  id: number
  title: string
  company_name: string
  verified: boolean
  city: string
  lat: number
  lng: number
  salary_min: number | null
  salary_max: number | null
}

export interface Breakdown {
  skills: number
  city: number
  schedule: number
  experience: number
}

export interface CandidateItem {
  score: number
  final_score: number
  ai_score: number | null
  ai_comment: string
  breakdown: Breakdown
  user: User
  resume: Resume
}

export interface CandidatesResponse {
  total: number
  items: CandidateItem[]
}

export interface RecruiterActionResponse {
  id: number
  action: string
  created_at: string
}

export interface InvitationCompany {
  id: number
  name: string
  verified: boolean
}

export interface InvitationVacancy {
  id: number
  title: string
  city: string
  work_format: string
  salary_min: number | null
  salary_max: number | null
}

export interface Invitation {
  id: number
  status: 'pending' | 'overdue' | 'responded'
  deadline_at: string
  hours_left: number
  company: InvitationCompany
  vacancy: InvitationVacancy
  response: string | null
}

export interface MatchResult {
  id: number
  response: string
  company: InvitationCompany
  vacancy: InvitationVacancy
  recruiter_contact: string
}

export interface ReferralsResponse {
  count: number
  items: { id: number; first_name: string; joined_at: string }[]
}

export const WORK_FORMATS = ['onsite', 'hybrid', 'remote']
export const EMPLOYMENT_TYPES = ['full_time', 'part_time', 'contract', 'internship']
export const WORK_FORMAT_LABELS: Record<string, string> = {
  onsite: 'Офис',
  hybrid: 'Гибрид',
  remote: 'Удалённо',
  full_time: 'Полный день',
  part_time: 'Частичная',
  contract: 'Контракт',
  internship: 'Стажировка',
}
