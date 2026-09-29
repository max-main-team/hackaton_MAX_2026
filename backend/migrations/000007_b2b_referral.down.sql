ALTER TABLE users DROP COLUMN IF EXISTS pending_ref_company_id;
ALTER TABLE companies DROP COLUMN IF EXISTS invite_used;
ALTER TABLE companies DROP COLUMN IF EXISTS invite_quota;
ALTER TABLE companies DROP COLUMN IF EXISTS promo_until;
ALTER TABLE companies DROP COLUMN IF EXISTS referrer_at;
ALTER TABLE companies DROP COLUMN IF EXISTS referrer_company_id;
