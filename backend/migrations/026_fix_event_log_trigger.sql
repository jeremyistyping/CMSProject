-- Fix event_uuid constraint issue in journal_event_log trigger
-- This migration explicitly sets event_uuid in the trigger function

BEGIN;

-- Drop and recreate the trigger function with explicit event_uuid generation
CREATE OR REPLACE FUNCTION log_journal_event()
RETURNS TRIGGER AS $$
DECLARE
    event_type_val VARCHAR(50);
    event_data_val JSONB;
    new_event_uuid UUID;
BEGIN
    -- Generate new UUID for this event
    new_event_uuid := uuid_generate_v4();
    
    -- Determine event type
    IF TG_OP = 'INSERT' THEN
        event_type_val := 'CREATED';
        event_data_val := jsonb_build_object(
            'operation', 'INSERT',
            'journal_id', NEW.id,
            'entry_number', NEW.entry_number,
            'source_type', NEW.source_type,
            'total_debit', NEW.total_debit,
            'total_credit', NEW.total_credit,
            'status', NEW.status
        );
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status != NEW.status AND NEW.status = 'POSTED' THEN
            event_type_val := 'POSTED';
        ELSIF OLD.status != NEW.status AND NEW.status = 'REVERSED' THEN
            event_type_val := 'REVERSED';
        ELSE
            event_type_val := 'UPDATED';
        END IF;
        
        event_data_val := jsonb_build_object(
            'operation', 'UPDATE',
            'journal_id', NEW.id,
            'entry_number', NEW.entry_number,
            'changes', jsonb_build_object(
                'status', jsonb_build_object('from', OLD.status, 'to', NEW.status),
                'total_debit', jsonb_build_object('from', OLD.total_debit, 'to', NEW.total_debit),
                'total_credit', jsonb_build_object('from', OLD.total_credit, 'to', NEW.total_credit)
            )
        );
    ELSIF TG_OP = 'DELETE' THEN
        event_type_val := 'DELETED';
        event_data_val := jsonb_build_object(
            'operation', 'DELETE',
            'journal_id', OLD.id,
            'entry_number', OLD.entry_number,
            'deleted_at', OLD.deleted_at
        );
    END IF;
    
    -- Insert event log with explicit event_uuid
    INSERT INTO journal_event_log (
        event_uuid,
        journal_id,
        event_type,
        event_data,
        user_id,
        correlation_id
    ) VALUES (
        new_event_uuid,
        COALESCE(NEW.id, OLD.id),
        event_type_val,
        event_data_val,
        COALESCE(NEW.created_by, OLD.created_by),
        COALESCE(NEW.transaction_uuid, OLD.transaction_uuid)
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

COMMIT;
