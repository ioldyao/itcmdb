-- Migration: Alert Center Enterprise Refactoring
-- Description: Transform fragmented alert service into clean, enterprise-grade alert center
-- Author: Claude Sonnet 4.5
-- Date: 2026-01-30

BEGIN;

-- ============================================================================
-- PHASE 1: Add new fields to alert_instances
-- ============================================================================

-- Add labels (JSONB) for label-based routing (Alertmanager-style)
ALTER TABLE alert_instances
ADD COLUMN IF NOT EXISTS labels JSONB DEFAULT '{}'::jsonb;

-- Add annotations (JSONB) for additional metadata
ALTER TABLE alert_instances
ADD COLUMN IF NOT EXISTS annotations JSONB DEFAULT '{}'::jsonb;

-- Add fingerprint for deduplication
ALTER TABLE alert_instances
ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(64);

-- Create GIN index for efficient label queries
CREATE INDEX IF NOT EXISTS idx_alert_instances_labels
ON alert_instances USING GIN (labels);

-- Create index for fingerprint lookups
CREATE INDEX IF NOT EXISTS idx_alert_instances_fingerprint
ON alert_instances (fingerprint);

-- Mark deprecated field with comment (keep for backward compatibility)
COMMENT ON COLUMN alert_instances.tags IS 'DEPRECATED: Use labels instead. Will be removed in future version.';
COMMENT ON COLUMN alert_instances.notification_channels IS 'DEPRECATED: Use routing rules instead. Will be removed in future version.';

-- ============================================================================
-- PHASE 2: Enhance alert_routing_rules with grouping features
-- ============================================================================

-- Add grouping configuration (Alertmanager-style)
ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS group_by TEXT[] DEFAULT ARRAY[]::TEXT[];

ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS group_wait INTEGER DEFAULT 30;

ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS group_interval INTEGER DEFAULT 300;

ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS repeat_interval INTEGER DEFAULT 3600;

-- Add hierarchical routing support
ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES alert_routing_rules(id) ON DELETE CASCADE;

ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS route_path TEXT;

-- Add continue flag for multi-rule matching
ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS continue_matching BOOLEAN DEFAULT false;

-- Create index for parent_id lookups
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_parent_id
ON alert_routing_rules (parent_id);

-- Create index for route_path
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_route_path
ON alert_routing_rules (route_path);

-- Add comments for new fields
COMMENT ON COLUMN alert_routing_rules.group_by IS 'Labels to group alerts by (e.g., [''alertname'', ''cluster''])';
COMMENT ON COLUMN alert_routing_rules.group_wait IS 'Seconds to wait before sending first notification for a group';
COMMENT ON COLUMN alert_routing_rules.group_interval IS 'Seconds to wait before sending updates for a group';
COMMENT ON COLUMN alert_routing_rules.repeat_interval IS 'Seconds to wait before repeating notifications';
COMMENT ON COLUMN alert_routing_rules.parent_id IS 'Parent routing rule for hierarchical routing';
COMMENT ON COLUMN alert_routing_rules.route_path IS 'Full path in routing tree (e.g., ''/root/team-a/critical'')';
COMMENT ON COLUMN alert_routing_rules.continue_matching IS 'Continue evaluating rules after match';

-- ============================================================================
-- PHASE 3: Create notification_logs table
-- ============================================================================

CREATE TABLE IF NOT EXISTS notification_logs (
    id SERIAL PRIMARY KEY,
    alert_instance_id INTEGER NOT NULL REFERENCES alert_instances(id) ON DELETE CASCADE,
    receiver_id INTEGER NOT NULL REFERENCES alert_receivers(id) ON DELETE CASCADE,
    receiver_group_id INTEGER NOT NULL REFERENCES alert_receiver_groups(id) ON DELETE CASCADE,
    routing_rule_id INTEGER REFERENCES alert_routing_rules(id) ON DELETE SET NULL,

    -- Notification details
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, sent, failed, retrying
    notification_type VARCHAR(50) NOT NULL, -- email, webhook, slack, etc.

    -- Content
    subject TEXT,
    body TEXT,
    rendered_template TEXT,

    -- Delivery tracking
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP,

    -- Error handling
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP,

    -- Metadata
    request_payload JSONB,
    response_payload JSONB,
    delivery_metadata JSONB DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT valid_status CHECK (status IN ('pending', 'sent', 'failed', 'retrying', 'max_retries_exceeded'))
);

-- Create indexes for notification_logs
CREATE INDEX IF NOT EXISTS idx_notification_logs_alert_instance_id
ON notification_logs (alert_instance_id);

CREATE INDEX IF NOT EXISTS idx_notification_logs_receiver_id
ON notification_logs (receiver_id);

CREATE INDEX IF NOT EXISTS idx_notification_logs_receiver_group_id
ON notification_logs (receiver_group_id);

CREATE INDEX IF NOT EXISTS idx_notification_logs_status
ON notification_logs (status);

CREATE INDEX IF NOT EXISTS idx_notification_logs_created_at
ON notification_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_logs_next_retry_at
ON notification_logs (next_retry_at) WHERE status = 'retrying';

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_notification_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_notification_logs_updated_at
    BEFORE UPDATE ON notification_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_logs_updated_at();

-- ============================================================================
-- PHASE 4: Mark deprecated tables and fields
-- ============================================================================

-- Mark deprecated fields in alert_rules
COMMENT ON COLUMN alert_rules.notification_channels IS 'DEPRECATED: Use routing rules instead. Will be removed in future version.';

-- Mark deprecated table alert_rule_receiver_groups
COMMENT ON TABLE alert_rule_receiver_groups IS 'DEPRECATED: Use alert_routing_rules instead. Will be removed in future version.';

-- Mark deprecated table outbound_webhooks
COMMENT ON TABLE outbound_webhooks IS 'DEPRECATED: Use alert_receivers with type=webhook instead. Will be removed in future version.';

-- Remove default_receiver_group_id from inbound_webhooks if it exists
ALTER TABLE inbound_webhooks
DROP COLUMN IF EXISTS default_receiver_group_id;

-- ============================================================================
-- PHASE 5: Create backward compatibility view
-- ============================================================================

-- Create view for legacy API compatibility
CREATE OR REPLACE VIEW alert_instances_with_legacy AS
SELECT
    ai.*,
    ai.labels AS tags_from_labels,
    COALESCE(ai.tags, '{}'::jsonb) AS legacy_tags
FROM alert_instances ai;

COMMENT ON VIEW alert_instances_with_legacy IS 'Backward compatibility view mapping labels to tags';

-- ============================================================================
-- PHASE 6: Create default routing rule if none exists
-- ============================================================================

-- Insert default catch-all routing rule if no rules exist
INSERT INTO alert_routing_rules (
    name,
    description,
    match_labels,
    match_type,
    receiver_group_id,
    priority,
    enabled,
    continue_matching,
    route_path
)
SELECT
    'Default Catch-All Route',
    'Default routing rule that matches all alerts. Created during migration.',
    '{}'::jsonb,
    'all',
    (SELECT id FROM alert_receiver_groups ORDER BY id LIMIT 1),
    0,
    true,
    false,
    '/default'
WHERE NOT EXISTS (SELECT 1 FROM alert_routing_rules)
AND EXISTS (SELECT 1 FROM alert_receiver_groups);

-- ============================================================================
-- PHASE 7: Add helpful functions
-- ============================================================================

-- Function to extract labels from tags (for migration)
CREATE OR REPLACE FUNCTION migrate_tags_to_labels()
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER := 0;
BEGIN
    UPDATE alert_instances
    SET labels = COALESCE(tags, '{}'::jsonb)
    WHERE labels = '{}'::jsonb
    AND tags IS NOT NULL
    AND tags != '{}'::jsonb;

    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION migrate_tags_to_labels() IS 'Migrate existing tags to labels field';

-- Function to generate fingerprint from labels
CREATE OR REPLACE FUNCTION generate_alert_fingerprint(labels_input JSONB)
RETURNS VARCHAR(64) AS $$
BEGIN
    RETURN encode(digest(labels_input::text, 'sha256'), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION generate_alert_fingerprint(JSONB) IS 'Generate SHA256 fingerprint from alert labels';

-- ============================================================================
-- PHASE 8: Data validation
-- ============================================================================

-- Ensure all alert_instances have valid labels
UPDATE alert_instances
SET labels = '{}'::jsonb
WHERE labels IS NULL;

-- Ensure all alert_instances have valid annotations
UPDATE alert_instances
SET annotations = '{}'::jsonb
WHERE annotations IS NULL;

-- Generate fingerprints for existing alerts without one
UPDATE alert_instances
SET fingerprint = generate_alert_fingerprint(labels)
WHERE fingerprint IS NULL AND labels != '{}'::jsonb;

-- ============================================================================
-- PHASE 9: Add constraints and validations
-- ============================================================================

-- Ensure labels is valid JSON
ALTER TABLE alert_instances
ADD CONSTRAINT alert_instances_labels_is_object
CHECK (jsonb_typeof(labels) = 'object');

-- Ensure annotations is valid JSON
ALTER TABLE alert_instances
ADD CONSTRAINT alert_instances_annotations_is_object
CHECK (jsonb_typeof(annotations) = 'object');

-- Ensure group_wait is positive
ALTER TABLE alert_routing_rules
ADD CONSTRAINT alert_routing_rules_group_wait_positive
CHECK (group_wait > 0);

-- Ensure group_interval is positive
ALTER TABLE alert_routing_rules
ADD CONSTRAINT alert_routing_rules_group_interval_positive
CHECK (group_interval > 0);

-- Ensure repeat_interval is positive
ALTER TABLE alert_routing_rules
ADD CONSTRAINT alert_routing_rules_repeat_interval_positive
CHECK (repeat_interval > 0);

-- ============================================================================
-- PHASE 10: Migration summary
-- ============================================================================

DO $$
DECLARE
    alert_count INTEGER;
    routing_rule_count INTEGER;
    receiver_group_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO alert_count FROM alert_instances;
    SELECT COUNT(*) INTO routing_rule_count FROM alert_routing_rules;
    SELECT COUNT(*) INTO receiver_group_count FROM alert_receiver_groups;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'Alert Center Migration Completed';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Alert Instances: %', alert_count;
    RAISE NOTICE 'Routing Rules: %', routing_rule_count;
    RAISE NOTICE 'Receiver Groups: %', receiver_group_count;
    RAISE NOTICE '========================================';
    RAISE NOTICE 'New Features:';
    RAISE NOTICE '  - Label-based routing';
    RAISE NOTICE '  - Unified notification logs';
    RAISE NOTICE '  - Alert grouping support';
    RAISE NOTICE '  - Hierarchical routing';
    RAISE NOTICE '  - Fingerprint deduplication';
    RAISE NOTICE '========================================';
END $$;

COMMIT;
