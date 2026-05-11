INSERT INTO app_users (username, password_hash)
VALUES ('admin', '83480902f34f7a8a5f67f2ef473ba4930214620dfb66c7f6de3adc06502fe46a')
ON CONFLICT (username) DO NOTHING;

INSERT INTO default_slots (class_id, slot_index, course_code, start_time, end_time, venue, status)
VALUES
    ('107125', 1, 'MAIR21', '09:20', '10:10', 'LH210', 'scheduled'),
    ('107125', 2, 'EEIR15', '10:30', '11:20', 'LH210', 'scheduled'),
    ('107125', 3, 'CSIR12', '13:30', '15:30', 'Violet Lab', 'scheduled')
ON CONFLICT (class_id, slot_index) DO NOTHING;