INSERT INTO courses (id, name, university) VALUES
(1, 'Licenciatura em Computação', 'UnB')
ON CONFLICT DO NOTHING;

UPDATE disciplines SET course_id = 1 WHERE course_id IS NULL;