ALTER TABLE tasks ADD COLUMN position integer NOT NULL DEFAULT 0;

UPDATE tasks
SET position = ordered.rn
FROM (
  SELECT id, row_number() OVER (ORDER BY created_at ASC, id ASC) AS rn
  FROM tasks
) AS ordered
WHERE tasks.id = ordered.id;
