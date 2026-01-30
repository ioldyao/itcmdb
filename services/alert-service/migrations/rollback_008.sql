-- Rollback Script: Alert Center Enterprise Refactoring
-- Description: Complete rollback to pre-migration state
-- Author: Claude Sonnet 4.5
-- Date: 2026-01-30

BEGIN;

RAISE NOTICE '========================================';
RAISE NOTICE 'Starting Rollback of Alert Center Migration';
RAISE NOTICE '========================================';

-- ============================================================================
-- PHASE 1: Drop new constraints
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping new constraints...';

    -- Drop alert_instances constraints
    ALTER TABLE alert_instances DROP CONSTRAINT IF EXISTS alert_instances_labels_is_object;
    ALTER TABLE alert_instances DROP CONSTRAINT IF EXISTS alert_instances_annotations_is_object;

    -- Drop alert_routing_rules constraints
    ALTER TABLE alert_routing_rules DROP CONSTRAINT IF EXISTS alert_routing_rules_group_wait_positive;
    ALTER TABLE alert_routing_rules DROP CONSTRAINT IF EXISTS alert_routing_rules_group_interval_positive;
    ALTER TABLE alert_routing_rules DROP CONSTRAINT IF EXISTS alert_routing_rules_repeat_interval_positive;

    RAISE NOTICE 'Constraints dropped successfully';
END $$;

-- ============================================================================
-- PHASE 2: Drop new indexes
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping new indexes...';

    -- Drop alert_instances indexes
    DROP INDEX IF EXISTS idx_alert_instances_labels;
    DROP INDEX IF EXISTS idx_alert_instances_fingerprint;
    DROP INDEX IF EXISTS idx_alert_instances_labels_severity;
    DROP INDEX IF EXISTS idx_alert_instances_labels_alertname;
    DROP INDEX IF EXISTS idx_alert_instances_labels_instance;

    -- Drop alert_routing_rules indexes
    DROP INDEX IF EXISTS idx_alert_routing_rules_parent_id;
    DROP INDEX IF EXISTS idx_alert_routing_rules_route_path;
    DROP INDEX IF EXISTS idx_alert_routing_rules_enabled_priority;

    -- Drop notification_logs indexes
    DROP INDEX IF EXISTS idx_notification_logs_alert_instance_id;
    DROP INDEX IF EXISTS idx_notification_logs_receiver_id;
    DROP INDEX IF EXISTS idx_notification_logs_receiver_group_id;
    DROP INDEX IF EXISTS idx_notification_logs_status;
    DROP INDEX IF EXISTS idx_notification_logs_created_at;
    DROP INDEX IF EXISTS idx_notification_logs_next_retry_at;

    RAISE NOTICE 'Indexes dropped successfully';
END $$;

-- ============================================================================
-- PHASE 3: Drop new functions and triggers
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping new functions and triggers...';

    -- Drop trigger
    DROP TRIGGER IF EXISTS trigger_notification_logs_updated_at ON notification_logs;

    -- Drop functions
    DROP FUNCTION IF EXISTS update_notification_logs_updated_at();
    DROP FUNCTION IF EXISTS migrate_tags_to_labels();
    DROP FUNCTION IF EXISTS generate_alert_fingerprint(JSONB);

    RAISE NOTICE 'Functions and triggers dropped successfully';
END $$;

-- ============================================================================
-- PHASE 4: Drop new views
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping new views...';

    DROP VIEW IF EXISTS alert_instances_with_legacy;

    RAISE NOTICE 'Views dropped successfully';
END $$;

-- ============================================================================
-- PHASE 5: Drop new table
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping notification_logs table...';

    DROP TABLE IF EXISTS notification_logs CASCADE;

    RAISE NOTICE 'notification_logs table dropped successfully';
END $$;

-- ============================================================================
-- PHASE 6: Remove new columns from alert_routing_rules
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Removing new columns from alert_routing_rules...';

    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS group_by;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS group_wait;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS group_interval;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS repeat_interval;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS parent_id;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS route_path;
    ALTER TABLE alert_routing_rules DROP COLUMN IF EXISTS continue_matching;

    RAISE NOTICE 'Columns removed from alert_routing_rules successfully';
END $$;

-- ============================================================================
-- PHASE 7: Remove new columns from alert_instances
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Removing new columns from alert_instances...';

    ALTER TABLE alert_instances DROP COLUMN IF EXISTS labels;
    ALTER TABLE alert_instances DROP COLUMN IF EXISTS annotations;
    ALTER TABLE alert_instances DROP COLUMN IF EXISTS fingerprint;

    RAISE NOTICE 'Columns removed from alert_instances successfully';
END $$;

-- ============================================================================
-- PHASE 8: Remove comments from deprecated fields
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Removing deprecation comments...';

    -- Remove comments from alert_instances
    COMMENT ON COLUMN alert_instances.tags IS NULL;
    COMMENT ON COLUMN alert_instances.notification_channels IS NULL;

    -- Remove comments from alert_rules
    COMMENT ON COLUMN alert_rules.notification_channels IS NULL;

    -- Remove comments from tables
    COMMENT ON TABLE alert_rule_receiver_groups IS NULL;
    COMMENT ON TABLE outbound_webhooks IS NULL;

    RAISE NOTICE 'Deprecation comments removed successfully';
END $$;

-- ============================================================================
-- PHASE 9: Delete auto-migrated routing rules
-- ============================================================================

DO $$
DECLARE
    deleted_count INTEGER := 0;
BEGIN
    RAISE NOTICE 'Deleting auto-migrated routing rules...';

    -- Delete routing rules created during migration
    DELETE FROM alert_routing_rules
    WHERE name LIKE 'Auto-migrated:%'
    OR name = 'Default Catch-All Route'
    OR name = 'Critical Alerts'
    OR name = 'Warning Alerts'
    OR name = 'Info Alerts';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RAISE NOTICE 'Deleted % auto-migrated routing rules', deleted_count;
END $$;

-- ============================================================================
-- PHASE 10: Restore default_receiver_group_id to inbound_webhooks (if needed)
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Checking if default_receiver_group_id needs to be restored...';

    -- Only restore if the column doesn't exist
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'inbound_webhooks'
        AND column_name = 'default_receiver_group_id'
    ) THEN
        ALTER TABLE inbound_webhooks
        ADD COLUMN default_receiver_group_id INTEGER REFERENCES alert_receiver_groups(id) ON DELETE SET NULL;

        RAISE NOTICE 'Restored default_receiver_group_id column to inbound_webhooks';
    ELSE
        RAISE NOTICE 'default_receiver_group_id column already exists, skipping';
    END IF;
END $$;

-- ============================================================================
-- PHASE 11: Validation
-- ============================================================================

DO $$
DECLARE
    labels_exists BOOLEAN;
    annotations_exists BOOLEAN;
    fingerprint_exists BOOLEAN;
    notification_logs_exists BOOLEAN;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Rollback Validation';
    RAISE NOTICE '========================================';

    -- Check if new columns still exist
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'alert_instances' AND column_name = 'labels'
    ) INTO labels_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'alert_instances' AND column_name = 'annotations'
    ) INTO annotations_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'alert_instances' AND column_name = 'fingerprint'
    ) INTO fingerprint_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'notification_logs'
    ) INTO notification_logs_exists;

    -- Report validation results
    IF labels_exists THEN
        RAISE WARNING 'labels column still exists in alert_instances';
    ELSE
        RAISE NOTICE '✓ labels column removed from alert_instances';
    END IF;

    IF annotations_exists THEN
        RAISE WARNING 'annotations column still exists in alert_instances';
    ELSE
        RAISE NOTICE '✓ annotations column removed from alert_instances';
    END IF;

    IF fingerprint_exists THEN
        RAISE WARNING 'fingerprint column still exists in alert_instances';
    ELSE
        RAISE NOTICE '✓ fingerprint column removed from alert_instances';
    END IF;

    IF notification_logs_exists THEN
        RAISE WARNING 'notification_logs table still exists';
    ELSE
        RAISE NOTICE '✓ notification_logs table removed';
    END IF;

    RAISE NOTICE '========================================';
END $$;

-- ============================================================================
-- PHASE 12: Summary
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Rollback Completed';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'The database has been restored to pre-migration state.';
    RAISE NOTICE 'Please verify that the application is functioning correctly.';
    RAISE NOTICE '========================================';
END $$;

COMMIT;
