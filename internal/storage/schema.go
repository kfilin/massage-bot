package storage

const Schema = `
CREATE TABLE IF NOT EXISTS patients (
    telegram_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    first_visit TIMESTAMP NOT NULL,
    last_visit TIMESTAMP NOT NULL,
    total_visits INTEGER NOT NULL DEFAULT 0,
    health_status TEXT,
    therapist_notes TEXT,
    voice_transcripts TEXT,
    current_service TEXT
);

CREATE TABLE IF NOT EXISTS blacklist (
    telegram_id TEXT PRIMARY KEY,
    username TEXT
);

CREATE TABLE IF NOT EXISTS analytics_events (
    id SERIAL PRIMARY KEY,
    patient_id TEXT,
    event_type TEXT NOT NULL,
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_analytics_patient_id ON analytics_events(patient_id);
CREATE INDEX IF NOT EXISTS idx_analytics_event_type ON analytics_events(event_type);

CREATE TABLE IF NOT EXISTS sessions (
    user_id BIGINT PRIMARY KEY,
    data JSONB NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS appointment_metadata (
    appointment_id TEXT PRIMARY KEY,
    confirmed_at TIMESTAMP,
    reminders_sent JSONB DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
