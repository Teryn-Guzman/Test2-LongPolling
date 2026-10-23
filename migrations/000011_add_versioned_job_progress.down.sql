BEGIN;

ALTER TABLE image_jobs
    DROP COLUMN simulation_step,
    DROP COLUMN stage,
    DROP COLUMN variants_completed,
    DROP COLUMN version;

COMMIT;