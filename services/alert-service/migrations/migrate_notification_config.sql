-- Data Migration: Migrate notification configuration to new routing system
-- Description: Migrate tags to labels, create routing rules from alert_rule_receiver_groups
-- Author: Claude Sonnet 4.5
-- Date: 2026-01-30

BEGIN;

-- ============================================================================
-- PHASE 1: Migrate tags to labels
-- ============================================================================

DO $$
DECLARE
    migrated_count INTEGER := 0;
BEGIN
    RAISE NOTICE 'Starting migration of tags to labels...';

    -- Migrate tags to labels for all alert instances
    UPDATE alert_instances
    SET labels = COALESCE(tags, '{}'::jsonb)
    WHERE (labels IS NULL OR labels = '{}'::jsonb)
    AND tags IS NOT NULL
    AND tags != '{}'::jsonb;

    GET DIAGNOSTICS migrated_count = ROW_COUNT;
    RAISE NOTICE 'Migrated % alert instances from tags to labels', migrated_count;
END $$;

-- ============================================================================
-- PHASE 2: Generate fingerprints for existing alerts
-- ============================================================================

DO $$
DECLARE
    fingerprint_count INTEGER := 0;
BEGIN
    RAISE NOTICE 'Generating fingerprints for existing alerts...';

    -- Generate fingerprints based on labels
    UPDATE alert_instances
    SET fingerprint = encode(digest(labels::text, 'sha256'), 'hex')
    WHERE fingerprint IS NULL
    AND labels IS NOT NULL
    AND labels != '{}'::jsonb;

    GET DIAGNOSTICS fingerprint_count = ROW_COUNT;
    RAISE NOTICE 'Generated % fingerprints', fingerprint_count;
END $$;

-- ============================================================================
-- PHASE 3: Migrate alert_rule_receiver_groups to routing rules
-- ============================================================================

DO $$
DECLARE
    rule_count INTEGER := 0;
    existing_rules INTEGER := 0;
BEGIN
    RAISE NOTICE 'Migrating alert_rule_receiver_groups to routing rules...';

    -- Check if there are existing routing rules
    SELECT COUNT(*) INTO existing_rules FROM alert_routing_rules;

    -- Only migrate if there are no existing routing rules (to avoid duplicates)
    IF existing_rules = 0 THEN
        -- Create routing rules from alert_rule_receiver_groups
        INSERT INTO alert_routing_rules (
            name,
            description,
            match_labels,
            match_type,
            receiver_group_id,
            priority,
            enabled,
            continue_matching,
            route_path,
            created_at,
            updated_at
        )
        SELECT
            'Auto-migrated: ' || ar.name,
            'Automatically migrated from alert rule: ' || ar.name || ' (ID: ' || ar.id || ')',
            jsonb_build_object('alert_rule_id', ar.id::text),
            'exact',
            arrg.receiver_group_id,
            100 - ROW_NUMBER() OVER (ORDER BY ar.id), -- Higher priority for earlier rules
            ar.enabled,
            false,
            '/migrated/rule-' || ar.id,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        FROM alert_rules ar
        INNER JOIN alert_rule_receiver_groups arrg ON ar.id = arrg.alert_rule_id
        WHERE arrg.receiver_group_id IS NOT NULL
        ON CONFLICT DO NOTHING;

        GET DIAGNOSTICS rule_count = ROW_COUNT;
        RAISE NOTICE 'Created % routing rules from alert_rule_receiver_groups', rule_count;
    ELSE
        RAISE NOTICE 'Skipping migration: % existing routing rules found', existing_rules;
    END IF;
END $$;

-- ============================================================================
-- PHASE 4: Create default catch-all routing rule
-- ============================================================================

DO $$
DECLARE
    default_receiver_group_id INTEGER;
    rule_exists BOOLEAN;
BEGIN
    RAISE NOTICE 'Creating default catch-all routing rule...';

    -- Check if a catch-all rule already exists
    SELECT EXISTS(
        SELECT 1 FROM alert_routing_rules
        WHERE match_labels = '{}'::jsonb
        AND match_type = 'all'
    ) INTO rule_exists;

    IF NOT rule_exists THEN
        -- Get the first available receiver group
        SELECT id INTO default_receiver_group_id
        FROM alert_receiver_groups
        ORDER BY id
        LIMIT 1;

        IF default_receiver_group_id IS NOT NULL THEN
            INSERT INTO alert_routing_rules (
                name,
                description,
                match_labels,
                match_type,
                receiver_group_id,
                priority,
                enabled,
                continue_matching,
                route_path,
                group_by,
                group_wait,
                group_interval,
                repeat_interval
            ) VALUES (
                'Default Catch-All Route',
                'Default routing rule that matches all alerts without specific routes',
                '{}'::jsonb,
                'all',
                default_receiver_group_id,
                0, -- Lowest priority
                true,
                false,
                '/default',
                ARRAY['alertname']::TEXT[],
                30,
                300,
                3600
            )
            ON CONFLICT DO NOTHING;

            RAISE NOTICE 'Created default catch-all routing rule with receiver group ID: %', default_receiver_group_id;
        ELSE
            RAISE WARNING 'No receiver groups found. Please create at least one receiver group.';
        END IF;
    ELSE
        RAISE NOTICE 'Default catch-all routing rule already exists';
    END IF;
END $$;

-- ============================================================================
-- PHASE 5: Migrate notification_channels to labels
-- ============================================================================

DO $$
DECLARE
    updated_count INTEGER := 0;
BEGIN
    RAISE NOTICE 'Migrating notification_channels to labels...';

    -- Add notification channel info to labels if it exists
    UPDATE alert_instances ai
    SET labels = labels || jsonb_build_object(
        'notification_channels',
        ai.notification_channels::text
    )
    WHERE ai.notification_channels IS NOT NULL
    AND ai.notification_channels != '{}'::jsonb
    AND NOT (labels ? 'notification_channels');

    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RAISE NOTICE 'Added notification_channels to labels for % alerts', updated_count;
END $$;

-- ============================================================================
-- PHASE 6: Create routing rules for severity-based routing
-- ============================================================================

DO $$
DECLARE
    critical_group_id INTEGER;
    warning_group_id INTEGER;
    info_group_id INTEGER;
BEGIN
    RAISE NOTICE 'Creating severity-based routing rules...';

    -- Get receiver groups (if they exist)
    SELECT id INTO critical_group_id FROM alert_receiver_groups WHERE name ILIKE '%critical%' LIMIT 1;
    SELECT id INTO warning_group_id FROM alert_receiver_groups WHERE name ILIKE '%warning%' LIMIT 1;
    SELECT id INTO info_group_id FROM alert_receiver_groups WHERE name ILIKE '%info%' LIMIT 1;

    -- Create critical severity routing rule
    IF critical_group_id IS NOT NULL THEN
        INSERT INTO alert_routing_rules (
            name,
            description,
            match_labels,
            match_type,
            receiver_group_id,
            priority,
            enabled,
            continue_matching,
            route_path,
            group_by,
            group_wait,
            group_interval,
            repeat_interval
        ) VALUES (
            'Critical Alerts',
            'Route critical severity alerts to critical receiver group',
            jsonb_build_object('severity', 'critical'),
            'exact',
            critical_group_id,
            90,
            true,
            false,
            '/severity/critical',
            ARRAY['alertname', 'instance']::TEXT[],
            10, -- Faster notification for critical
            60,
            1800
        )
        ON CONFLICT DO NOTHING;

        RAISE NOTICE 'Created critical severity routing rule';
    END IF;

    -- Create warning severity routing rule
    IF warning_group_id IS NOT NULL THEN
        INSERT INTO alert_routing_rules (
            name,
            description,
            match_labels,
            match_type,
            receiver_group_id,
            priority,
            enabled,
            continue_matching,
            route_path,
            group_by,
            group_wait,
            group_interval,
            repeat_interval
        ) VALUES (
            'Warning Alerts',
            'Route warning severity alerts to warning receiver group',
            jsonb_build_object('severity', 'warning'),
            'exact',
            warning_group_id,
            80,
            true,
            false,
            '/severity/warning',
            ARRAY['alertname']::TEXT[],
            30,
            300,
            3600
        )
        ON CONFLICT DO NOTHING;

        RAISE NOTICE 'Created warning severity routing rule';
    END IF;

    -- Create info severity routing rule
    IF info_group_id IS NOT NULL THEN
        INSERT INTO alert_routing_rules (
            name,
            description,
            match_labels,
            match_type,
            receiver_group_id,
            priority,
            enabled,
            continue_matching,
            route_path,
            group_by,
            group_wait,
            group_interval,
            repeat_interval
        ) VALUES (
            'Info Alerts',
            'Route info severity alerts to info receiver group',
            jsonb_build_object('severity', 'info'),
            'exact',
            info_group_id,
            70,
            true,
            false,
            '/severity/info',
            ARRAY['alertname']::TEXT[],
            60,
            600,
            7200
        )
        ON CONFLICT DO NOTHING;

        RAISE NOTICE 'Created info severity routing rule';
    END IF;
END $$;

-- ============================================================================
-- PHASE 7: Validation queries
-- ============================================================================

DO $$
DECLARE
    total_alerts INTEGER;
    alerts_with_labels INTEGER;
    alerts_with_fingerprints INTEGER;
    total_routing_rules INTEGER;
    total_receiver_groups INTEGER;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Migration Validation';
    RAISE NOTICE '========================================';

    -- Count alerts
    SELECT COUNT(*) INTO total_alerts FROM alert_instances;
    SELECT COUNT(*) INTO alerts_with_labels FROM alert_instances WHERE labels != '{}'::jsonb;
    SELECT COUNT(*) INTO alerts_with_fingerprints FROM alert_instances WHERE fingerprint IS NOT NULL;

    RAISE NOTICE 'Total Alerts: %', total_alerts;
    RAISE NOTICE 'Alerts with Labels: % (%.1f%%)',
        alerts_with_labels,
        CASE WHEN total_alerts > 0 THEN (alerts_with_labels::float / total_alerts * 100) ELSE 0 END;
    RAISE NOTICE 'Alerts with Fingerprints: % (%.1f%%)',
        alerts_with_fingerprints,
        CASE WHEN total_alerts > 0 THEN (alerts_with_fingerprints::float / total_alerts * 100) ELSE 0 END;

    -- Count routing rules
    SELECT COUNT(*) INTO total_routing_rules FROM alert_routing_rules;
    RAISE NOTICE 'Total Routing Rules: %', total_routing_rules;

    -- Count receiver groups
    SELECT COUNT(*) INTO total_receiver_groups FROM alert_receiver_groups;
    RAISE NOTICE 'Total Receiver Groups: %', total_receiver_groups;

    RAISE NOTICE '========================================';

    -- Warnings
    IF total_routing_rules = 0 THEN
        RAISE WARNING 'No routing rules found! Alerts will not be routed.';
    END IF;

    IF total_receiver_groups = 0 THEN
        RAISE WARNING 'No receiver groups found! Please create receiver groups.';
    END IF;

    IF alerts_with_labels = 0 AND total_alerts > 0 THEN
        RAISE WARNING 'No alerts have labels! Check migration logic.';
    END IF;
END $$;

-- ============================================================================
-- PHASE 8: Create indexes for performance
-- ============================================================================

-- Index for label-based queries (already created in main migration, but ensure it exists)
CREATE INDEX IF NOT EXISTS idx_alert_instances_labels_severity
ON alert_instances ((labels->>'severity'));

CREATE INDEX IF NOT EXISTS idx_alert_instances_labels_alertname
ON alert_instances ((labels->>'alertname'));

CREATE INDEX IF NOT EXISTS idx_alert_instances_labels_instance
ON alert_instances ((labels->>'instance'));

-- Index for routing rule lookups
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_enabled_priority
ON alert_routing_rules (enabled, priority DESC);

RAISE NOTICE 'Created performance indexes';

-- ============================================================================
-- PHASE 9: Summary
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Data Migration Completed Successfully';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '1. Verify routing rules in the UI';
    RAISE NOTICE '2. Test alert ingestion with new labels';
    RAISE NOTICE '3. Monitor notification delivery';
    RAISE NOTICE '4. Update alert rules to use labels';
    RAISE NOTICE '========================================';
END $$;

COMMIT;
