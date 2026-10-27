package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrRecordNotFound = errors.New("record not found")

// Image stores the uploaded original file and its metadata.
type Image struct {
	ID, OriginalFilename, StoredFilename, MediaType string
	Size                                            int64
}

// ImageJob tracks the lifecycle of an accepted upload from queued to completed/failed.
type ImageJob struct {
	ID                string     `json:"id"`
	PublicID          string     `json:"public_id"`
	ImageID           string     `json:"image_id"`
	Status            string     `json:"status"`
	Version           int64      `json:"version"`
	Stage             string     `json:"stage"`
	VariantsCompleted int        `json:"variants_completed"`
	SimulationStep    int        `json:"simulation_step"`
	QueuePosition     int        `json:"queue_position"`
	Error             *string    `json:"error,omitempty"`
	QueuedAt          time.Time  `json:"queued_at"`
	StartedAt         *time.Time `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at"`
	FailedAt          *time.Time `json:"failed_at,omitempty"`
	Variants          []Variant  `json:"variants,omitempty"`
}

// JobCommit describes an observable job snapshot committed by a worker update.
type JobCommit struct {
	PublicID          string
	Version           int64
	Status            string
	Stage             string
	VariantsCompleted int
	SimulationStep    int
	CompletedAt       *time.Time
	FailedAt          *time.Time
}

// Variant stores the generated output metadata returned to the browser.
type Variant struct {
	Name     string `json:"name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	URL      string `json:"url"`
	Filename string `json:"-"`
	Size     int64  `json:"-"`
}

type ImageModel struct{ DB *sql.DB }

// Insert persists the original upload and creates the initial queued job.
func (m ImageModel) Insert(ctx context.Context, image *Image, job *ImageJob) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `INSERT INTO images (original_filename, stored_filename, media_type, size_bytes) VALUES ($1,$2,$3,$4) RETURNING id`, image.OriginalFilename, image.StoredFilename, image.MediaType, image.Size).Scan(&image.ID); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `INSERT INTO image_jobs (image_id) VALUES ($1) RETURNING id, public_id, status, queued_at, version, stage, variants_completed, simulation_step`, image.ID).Scan(&job.ID, &job.PublicID, &job.Status, &job.QueuedAt, &job.Version, &job.Stage, &job.VariantsCompleted, &job.SimulationStep); err != nil {
		return err
	}
	job.ImageID = image.ID
	return tx.Commit()
}

// ClaimNext atomically claims the next queued job for processing.
func (m ImageModel) ClaimNext(ctx context.Context) (*ImageJob, []JobCommit, error) {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	job := &ImageJob{}
	if err := tx.QueryRowContext(ctx, `SELECT id, public_id, image_id FROM image_jobs WHERE status='queued' ORDER BY queued_at, id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&job.ID, &job.PublicID, &job.ImageID); err != nil {
		return nil, nil, err
	}
	if err := tx.QueryRowContext(ctx, `UPDATE image_jobs SET status='processing', stage='starting', started_at=now(), version=version+1 WHERE id=$1 RETURNING started_at, version`, job.ID).Scan(&job.StartedAt, &job.Version); err != nil {
		return nil, nil, err
	}
	job.Status = "processing"
	job.Stage = "starting"
	queueChanges := make([]JobCommit, 0)
	rows, err := tx.QueryContext(ctx, `UPDATE image_jobs SET version=version+1 WHERE status='queued' RETURNING public_id,version,status,stage,variants_completed,simulation_step,completed_at,failed_at`)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var commit JobCommit
		if err := rows.Scan(&commit.PublicID, &commit.Version, &commit.Status, &commit.Stage, &commit.VariantsCompleted, &commit.SimulationStep, &commit.CompletedAt, &commit.FailedAt); err != nil {
			rows.Close()
			return nil, nil, err
		}
		queueChanges = append(queueChanges, commit)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return job, queueChanges, nil
}

// GetJob returns the current job state together with any generated variant metadata.
func (m ImageModel) GetJob(ctx context.Context, id string) (*ImageJob, error) {
	job, _, err := m.GetJobWithReads(ctx, id)
	return job, err
}

// GetJobWithReads returns the complete current snapshot and the number of SQL
// statements attempted, for explicit observation-query accounting. Variant
// metadata is only included once the job is completed (all-or-complete results);
// progress before that is reported through variants_completed and stage, so
// queued, processing and failed jobs cost one read instead of two.
func (m ImageModel) GetJobWithReads(ctx context.Context, id string) (*ImageJob, int, error) {
	tx, err := m.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	job := &ImageJob{Variants: make([]Variant, 0)}
	err = tx.QueryRowContext(ctx, `SELECT id,public_id,image_id,status,error_message,queued_at,started_at,completed_at,failed_at,version,stage,variants_completed,simulation_step,
		CASE WHEN status='queued' THEN (
			SELECT COUNT(*) FROM image_jobs AS earlier
			WHERE earlier.status='queued'
			AND (earlier.queued_at, earlier.id) <= (image_jobs.queued_at, image_jobs.id)
		) ELSE 0 END
		FROM image_jobs WHERE public_id=$1`, id).Scan(&job.ID, &job.PublicID, &job.ImageID, &job.Status, &job.Error, &job.QueuedAt, &job.StartedAt, &job.CompletedAt, &job.FailedAt, &job.Version, &job.Stage, &job.VariantsCompleted, &job.SimulationStep, &job.QueuePosition)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 1, ErrRecordNotFound
	}
	if err != nil {
		return nil, 1, err
	}
	if job.Status != "completed" {
		if err := tx.Commit(); err != nil {
			return nil, 1, err
		}
		return job, 1, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT name,width,height,stored_filename,size_bytes FROM image_variants WHERE image_id=$1 ORDER BY CASE name WHEN 'thumbnail' THEN 1 WHEN 'preview' THEN 2 ELSE 3 END`, job.ImageID)
	if err != nil {
		return nil, 2, err
	}
	defer rows.Close()
	for rows.Next() {
		var v Variant
		if err := rows.Scan(&v.Name, &v.Width, &v.Height, &v.Filename, &v.Size); err != nil {
			return nil, 2, err
		}
		v.URL = "/v1/images/" + job.ImageID + "/variants/" + v.Name
		job.Variants = append(job.Variants, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 2, err
	}
	if err := rows.Close(); err != nil {
		return nil, 2, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 2, err
	}
	return job, 2, nil
}

// Original fetches the uploaded image record by its database ID.
func (m ImageModel) Original(ctx context.Context, imageID string) (Image, error) {
	var image Image
	err := m.DB.QueryRowContext(ctx, `SELECT id,original_filename,stored_filename,media_type,size_bytes FROM images WHERE id=$1`, imageID).Scan(&image.ID, &image.OriginalFilename, &image.StoredFilename, &image.MediaType, &image.Size)
	if errors.Is(err, sql.ErrNoRows) {
		return image, ErrRecordNotFound
	}
	return image, err
}

// AddVariantAndProgress stores real variant metadata and the matching versioned
// milestone atomically. The last variant and completed state share one commit.
func (m ImageModel) AddVariantAndProgress(ctx context.Context, jobID, imageID, name, filename string, width, height int, size int64, final bool) (JobCommit, error) {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return JobCommit{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO image_variants (image_id,name,stored_filename,width,height,size_bytes) VALUES ($1,$2,$3,$4,$5,$6)`, imageID, name, filename, width, height, size); err != nil {
		return JobCommit{}, err
	}
	stage := name + "_ready"
	query := `UPDATE image_jobs SET variants_completed=variants_completed+1, stage=$2, version=version+1 WHERE id=$1 AND status='processing' RETURNING public_id,version,status,stage,variants_completed,simulation_step,completed_at,failed_at`
	if final {
		query = `UPDATE image_jobs SET variants_completed=variants_completed+1, stage=$2, status='completed', completed_at=now(), version=version+1 WHERE id=$1 AND status='processing' RETURNING public_id,version,status,stage,variants_completed,simulation_step,completed_at,failed_at`
	}
	var commit JobCommit
	if err := tx.QueryRowContext(ctx, query, jobID, stage).Scan(&commit.PublicID, &commit.Version, &commit.Status, &commit.Stage, &commit.VariantsCompleted, &commit.SimulationStep, &commit.CompletedAt, &commit.FailedAt); err != nil {
		return JobCommit{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobCommit{}, err
	}
	return commit, nil
}

// AdvanceSimulation persists one experimental-only progress change.
func (m ImageModel) AdvanceSimulation(ctx context.Context, jobID string) (JobCommit, error) {
	var commit JobCommit
	err := m.DB.QueryRowContext(ctx, `UPDATE image_jobs SET simulation_step=simulation_step+1, version=version+1 WHERE id=$1 AND status='processing' RETURNING public_id,version,status,stage,variants_completed,simulation_step,completed_at,failed_at`, jobID).Scan(&commit.PublicID, &commit.Version, &commit.Status, &commit.Stage, &commit.VariantsCompleted, &commit.SimulationStep, &commit.CompletedAt, &commit.FailedAt)
	return commit, err
}

// Fail records a safe terminal failure and increments the representation version.
func (m ImageModel) Fail(ctx context.Context, id, message string) (JobCommit, error) {
	var commit JobCommit
	err := m.DB.QueryRowContext(ctx, `UPDATE image_jobs SET status='failed',error_message=$2,failed_at=now(),version=version+1 WHERE id=$1 AND status IN ('queued','processing') RETURNING public_id,version,status,stage,variants_completed,simulation_step,completed_at,failed_at`, id, message).Scan(&commit.PublicID, &commit.Version, &commit.Status, &commit.Stage, &commit.VariantsCompleted, &commit.SimulationStep, &commit.CompletedAt, &commit.FailedAt)
	return commit, err
}

// Variant looks up a single generated output by image ID and variant name. It
// only finds outputs of completed jobs, so partial results of a processing or
// failed job are not served even though their files already exist on disk.
func (m ImageModel) Variant(ctx context.Context, imageID, name string) (Variant, error) {
	var v Variant
	err := m.DB.QueryRowContext(ctx, `SELECT v.name,v.width,v.height,v.stored_filename,v.size_bytes
		FROM image_variants AS v
		JOIN image_jobs AS j ON j.image_id=v.image_id
		WHERE v.image_id=$1 AND v.name=$2 AND j.status='completed'`, imageID, name).Scan(&v.Name, &v.Width, &v.Height, &v.Filename, &v.Size)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrRecordNotFound
	}
	return v, err
}
