CREATE TABLE IF NOT EXISTS app_users (
    username TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS default_slots (
    class_id TEXT NOT NULL,
    slot_index INTEGER NOT NULL,
    course_code TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    venue TEXT NOT NULL,
    status TEXT NOT NULL,
    PRIMARY KEY (class_id, slot_index)
);

CREATE TABLE IF NOT EXISTS overrides (
    class_id TEXT NOT NULL,
    slot_index INTEGER NOT NULL,
    course_code TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    venue TEXT NOT NULL,
    status TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, slot_index)
);