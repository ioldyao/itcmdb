-- Cleanup Script: Remove deprecated tables and columns
-- Description: Clean up deprecated schema elements after 4 weeks of stable operation
-- Author: Claude Sonnet 4.5
-- Date: 2026-01-30
-- IMPORTANT: Only run this after verifying 4+ weeks of stable operation with the new system

BEGIN;

RAISE NOTICE '========================================';
RAISE NOTICE 'Starting Cleanup of Deprecated Schema Elements';
RAISE NOTICE '========================================';
RAISE NOTICE 'WARNING: This will permanently remove deprecated tables and columns';
RAISE NOTICE 'Ensure you have a backup before proceeding!';
RAISE NOTICE '========================================';

-- ============================================================================
-- PHASE 1: Verify new system is operational
-- ============================================================================

DO $$
DECLARE
    notification_log_count INTEGER;
    routing_rule_count INTEGER;
    alert_with_labels_count INTEGER;
BEGIN
    RAISE NOTICE 'Verifying new system is operational...';

    -- Check notification_logs table exists and has data
    SELECT COUNT(*) INTO notification_log_count FROM notification_logs;
    IF notification_log_count = 0 THEN
        RAISE WARNING 'notification_logs table is empty. Are you sure the new system is working?';
    ELSE
        RAISE NOTICE 'Found % notification logs', notification_log_count;
    END IF;

    -- Check routing rules exist
    SELECT COUNT(*) INTO routing_rule_count FROM alert_routing_rules;
    IF routing_rule_count = 0 THEN
        RAISE WARNING 'No routing rules found. System may not be routing alerts properly.';
    ELSE
        RAISE NOTICE 'Found % routing rules', routing_rule_count;
    END IF;

    -- Check alerts have labels
    SELECT COUNT(*) INTO alert_with_labels_count
    FROM alert_instances
    WHERE labels IS NOT NULL AND labels != '{}'::jsonb;

    RAISE NOTICE 'Found % alerts with labels', alert_with_labels_count;
END $$;

-- ============================================================================
-- PHASE 2: Drop deprecated columns from alert_rules
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Removing deprecated notification_channels column from alert_rules...';

    ALTER TABLE alert_rules DROP COLUMN IF EXISTS notification_channels;

    RAISE NOTICE 'Removed notification_channels from alert_rules';
END $$;

-- ============================================================================
-- PHASE 3: Drop deprecated columns from alert_instances
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Removing deprecated columns from alert_instances...';

    -- Drop notification_channels (replaced by routing rules + notification_logs)
    ALTER TABLE alert_instances DROP COLUMN IF EXISTS notification_channels;

    -- Drop tags (replaced by labels)
    ALTER TABLE alert_instances DROP COLUMN IF EXISTS tags;

    RAISE NOTICE 'Removed deprecated columns from alert_instances';
END $$;

-- ============================================================================
-- PHASE 4: Drop deprecated junction table
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping alert_rule_receiver_groups junction table...';

    DROP TABLE IF EXISTS alert_rule_receiver_groups CASCADE;

    RAISE NOTICE 'Dropped alert_rule_receiver_groups table';
END $$;

-- ============================================================================
-- PHASE 5: Drop deprecated outbound webhooks tables
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping deprecated outbound webhook tables...';

    -- Drop logs first (has foreign key to outbound_webhooks)
    DROP TABLE IF EXISTS outbound_webhook_logs CASCADE;

    -- Drop main table
    DROP TABLE IF EXISTS outbound_webhooks CASCADE;

    RAISE NOTICE 'Dropped outbound webhook tables';
END $$;

-- ============================================================================
-- PHASE 6: Drop backward compatibility view
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping backward compatibility view...';

    DROP VIEW IF EXISTS alert_instances_with_legacy CASCADE;

    RAISE NOTICE 'Dropped alert_instances_with_legacy view';
END $$;

-- ============================================================================
-- PHASE 7: Drop migration helper functions
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Dropping migration helper functions...';

    DROP FUNCTION IF EXISTS migrate_tags_to_labels();
    DROP FUNCTION IF EXISTS generate_alert_fingerprint(JSONB);

    RAISE NOTICE 'Dropped migration helper functions';
END $$;

-- ============================================================================
-- PHASE 8: Clean up auto-migrated routing rules (optional)
-- ============================================================================

DO $$
DECLARE
    cleanup_migrated_rules BOOLEAN := FALSE; -- Set to TRUE to clean up auto-migrated rules
    deleted_count INTEGER := 0;
BEGIN
    IF cleanup_migrated_rules THEN
        RAISE NOTICE 'Cleaning up auto-migrated routing rules...';

        DELETE FROM alert_routing_rules
        WHERE name LIKE 'Auto-migrated:%'
        OR route_path LIKE '/migrated/%';

        GET DIAGNOSTICS deleted_count = ROW_COUNT;
        RAISE NOTICE 'Deleted % auto-migrated routing rules', deleted_count;
    ELSE
        RAISE NOTICE 'Skipping cleanup of auto-migrated routing rules (set cleanup_migrated_rules to TRUE to enable)';
    END IF;
END $$;

-- ============================================================================
-- PHASE 9: Vacuum and analyze
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Running VACUUM ANALYZE to reclaim space and update statistics...';
END $$;

VACUUM ANALYZE alert_rules;
VACUUM ANALYZE alert_instances;
VACUUM ANALYZE alert_routing_rules;
VACUUM ANALYZE notification_logs;

-- ============================================================================
-- PHASE 10: Final validation
-- ============================================================================

DO $$
DECLARE
    tags_exists BOOLEAN;
    notification_channels_exists BOOLEAN;
    outbound_webhooks_exists BOOLEAN;
    alert_rule_receiver_groups_exists BOOLEAN;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Cleanup Validation';
    RAISE NOTICE '========================================';

    -- Check if deprecated columns still exist
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'alert_instances' AND column_name = 'tags'
    ) INTO tags_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'alert_instances' AND column_name = 'notification_channels'
    ) INTO notification_channels_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'outbound_webhooks'
    ) INTO outbound_webhooks_exists;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'alert_rule_receiver_groups'
    ) INTO alert_rule_receiver_groups_exists;

    -- Report validation results
    IF tags_exists THEN
        RAISE WARNING 'tags column still exists in alert_instances';
    ELSE
        RAISE NOTICE '✓ tags column removed from alert_instances';
    END IF;

    IF notification_channels_exists THEN
        RAISE WARNING 'notification_channels column still exists in alert_instances';
    ELSE
        RAISE NOTICE '✓ notification_channels column removed from alert_instances';
    END IF;

    IF outbound_webhooks_exists THEN
        RAISE WARNING 'outbound_webhooks table still exists';
    ELSE
        RAISE NOTICE '✓ outbound_webhooks table removed';
    END IF;

    IF alert_rule_receiver_groups_exists THEN
        RAISE WARNING 'alert_rule_receiver_groups table still exists';
    ELSE
        RAISE NOTICE '✓ alert_rule_receiver_groups table removed';
    END IF;

    RAISE NOTICE '========================================';
END $$;

-- ============================================================================
-- PHASE 11: Summary
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Cleanup Completed Successfully';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Removed:';
    RAISE NOTICE '  - alert_rules.notification_channels column';
    RAISE NOTICE '  - alert_instances.notification_channels column';
    RAISE NOTICE '  - alert_instances.tags column';
    RAISE NOTICE '  - alert_rule_receiver_groups table';
    RAISE NOTICE '  - outbound_webhooks table';
    RAISE NOTICE '  - outbound_webhook_logs table';
    RAISE NOTICE '  - alert_instances_with_legacy view';
    RAISE NOTICE '  - Migration helper functions';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'The alert center refactoring is now complete!';
    RAISE NOTICE '========================================';
END $$;

COMMIT;
