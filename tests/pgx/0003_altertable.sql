ALTER TABLE example ADD COLUMN IF NOT EXISTS ctype enumtype NOT NULL;

ALTER TABLE example ADD COLUMN IF NOT EXISTS depends_on enumtype[];
