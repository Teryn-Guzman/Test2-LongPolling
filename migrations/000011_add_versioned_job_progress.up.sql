BEGIN;

ALTER TABLE image_jobs
    ADD COLUMN version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    ADD COLUMN variants_completed integer NOT NULL DEFAULT 0 CHECK (variants_completed BETWEEN 0 AND 3),
    ADD COLUMN stage text NOT NULL DEFAULT 'starting',
    ADD COLUMN simulation_step integer NOT NULL DEFAULT 0 CHECK (simulation_step >= 0);

UPDATE image_jobs AS jobs
SET variants_completed = counts.completed,
    stage = CASE
        WHEN counts.completed >= 3 OR jobs.status = 'completed' THEN 'display_ready'
        WHEN counts.completed >= 2 THEN 'preview_ready'
        WHEN counts.completed = 1 THEN 'thumbnail_ready'
        ELSE 'starting'
    END
FROM (
    SELECT image_id, COUNT(*)::integer AS completed
    FROM image_variants
    GROUP BY image_id
) AS counts
WHERE jobs.image_id = counts.image_id;

COMMIT;