DO 'BEGIN CREATE TYPE enumtype AS ENUM(); EXCEPTION WHEN duplicate_object THEN null; END';

ALTER TYPE enumtype ADD VALUE IF NOT EXISTS 'type1';

ALTER TYPE enumtype ADD VALUE IF NOT EXISTS 'type2';
