ALTER TABLE companies ADD COLUMN referrer_company_id BIGINT REFERENCES companies (id);
ALTER TABLE companies ADD COLUMN referrer_at TIMESTAMPTZ;
ALTER TABLE companies ADD COLUMN promo_until TIMESTAMPTZ;
ALTER TABLE companies ADD COLUMN invite_quota INTEGER NOT NULL DEFAULT 3;
ALTER TABLE companies ADD COLUMN invite_used INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN pending_ref_company_id BIGINT REFERENCES companies (id);
