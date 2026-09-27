DROP INDEX IF EXISTS idx_company_members_user;
DROP INDEX IF EXISTS idx_resume_versions_user;
DROP INDEX IF EXISTS idx_resumes_active;
DROP INDEX IF EXISTS idx_vacancies_company;

DROP TABLE IF EXISTS vacancies;
DROP TABLE IF EXISTS resume_versions;
DROP TABLE IF EXISTS resumes;
DROP TABLE IF EXISTS company_members;
DROP TABLE IF EXISTS companies;

ALTER TABLE users DROP COLUMN IF EXISTS gender;
ALTER TABLE users DROP COLUMN IF EXISTS birth_date;
ALTER TABLE users DROP COLUMN IF EXISTS role;
