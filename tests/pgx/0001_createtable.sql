CREATE TABLE IF NOT EXISTS example();

ALTER TABLE example ADD COLUMN IF NOT EXISTS rowid BIGINT UNIQUE NOT NULL;