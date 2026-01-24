-- Migration script to add missing columns to CMDB tables
-- This fixes the "CI type not found" error by aligning database schema with Go models

-- Add missing columns to ci_types table
ALTER TABLE ci_types
ADD COLUMN IF NOT EXISTS display_name VARCHAR(100);

-- Update existing records to have display_name
UPDATE ci_types SET display_name = name WHERE display_name IS NULL;

-- Make display_name NOT NULL after populating
ALTER TABLE ci_types
ALTER COLUMN display_name SET NOT NULL;

-- Add missing columns to ci_attributes table
ALTER TABLE ci_attributes
ADD COLUMN IF NOT EXISTS display_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS is_unique BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS default_value VARCHAR(255),
ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Update existing records
UPDATE ci_attributes SET display_name = name WHERE display_name IS NULL;

-- Make display_name NOT NULL after populating
ALTER TABLE ci_attributes
ALTER COLUMN display_name SET NOT NULL;

-- Add index for soft deletes
CREATE INDEX IF NOT EXISTS idx_ci_attributes_deleted_at ON ci_attributes(deleted_at);

-- Add missing columns to ci_instances table
ALTER TABLE ci_instances
ADD COLUMN IF NOT EXISTS tags JSONB,
ADD COLUMN IF NOT EXISTS created_by INTEGER,
ADD COLUMN IF NOT EXISTS updated_by INTEGER;

-- Note: deleted_at already exists in ci_instances

-- Add indexes for ci_instances
CREATE INDEX IF NOT EXISTS idx_ci_instances_created_by ON ci_instances(created_by);
CREATE INDEX IF NOT EXISTS idx_ci_instances_updated_by ON ci_instances(updated_by);

-- Add missing columns to ci_relations table
ALTER TABLE ci_relations
ADD COLUMN IF NOT EXISTS description VARCHAR(255),
ADD COLUMN IF NOT EXISTS created_by INTEGER,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Add indexes for ci_relations
CREATE INDEX IF NOT EXISTS idx_ci_relations_created_by ON ci_relations(created_by);
CREATE INDEX IF NOT EXISTS idx_ci_relations_deleted_at ON ci_relations(deleted_at);

-- Add missing columns to ci_history table
ALTER TABLE ci_history
ADD COLUMN IF NOT EXISTS action VARCHAR(20);

-- Update existing records with default action
UPDATE ci_history SET action = 'update' WHERE action IS NULL;

-- Make action NOT NULL after populating
ALTER TABLE ci_history
ALTER COLUMN action SET NOT NULL;

-- Add trigger for ci_attributes updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ci_attributes_updated_at
    BEFORE UPDATE ON ci_attributes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_relations_updated_at
    BEFORE UPDATE ON ci_relations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Verify the changes
SELECT 'Migration completed successfully' AS status;
