# Assessment 2 experiment procedure and evidence

No measurement results are included in the repository yet. Fill in predictions before each condition, collect actual browser/server logs, then run the analyzer. It creates the CSV, per-run and grouped summary tables, and labelled SVG charts from those logs only.

## Before running

1. Record and verify the Assessment 1 baseline commit/tag. The current pre-Assessment-2 source tree began at `e03682d`; verify it is the submitted Assessment 1 state before describing it as the official baseline.
2. Fill in the five predictions in `predictions.md` before taking measurements.
3. Use the same image file and keep image transformations unchanged. Clear the worker queue before every run; use one browser tab/observer and keep it foregrounded.
4. Alternate mode order across repetitions (for example, short then long on repetitions 1 and 3, long then short on repetition 2).
5. Record the latency configuration in the page before each run. In Chrome DevTools, open **Network**, open the throttling dropdown, choose **Add…/Custom**, set download/upload bandwidth to **No throttling** and latency to **200 ms**, save/select that profile, then make a prompt GET and record its measured client request duration from the downloaded log. If your browser version/tool cannot set that exact profile, document the actual setting and measured prompt-GET timing; do not report a nominal 200 ms that was not achieved.
6. Start a fresh accepted job for every measured run. The observation window begins with the first GET, not the upload request. Do not reuse a completed job as a run.

## Workload controls

The Go server workload flags default to zero/disabled. Restart the API for each condition and append stdout JSON logs to a single file:

- **A — Quiet job:** run with `quiet_delay=30s`; no simulated updates. Real transformations still run after the delay.
- **B — Real progress:** run with all delay/simulation settings left at zero.
- **C1 — Slow simulated updates:** run with `simulation_duration=10s simulation_interval=2s`.
- **C2 — Fast simulated updates:** run with `simulation_duration=10s simulation_interval=200ms`.
- **C3 — Fast updates with latency:** same server settings as C2 plus the recorded client/network latency configuration.

For each job, the simulation (if enabled) runs before real transformations. It only increments `simulation_step` and job `version`; it does not change `variants_completed`, create files, or complete the job. The UI labels this as simulated progress. Do not use simulated progress for ordinary-use demonstrations. For C3, the browser-side 200 ms delay must be active while the server is configured for the same 10-second/200 ms simulated phase used in C2.

Example capture workflow from the repository root:

```sh
mkdir -p experiments/raw
source .envrc
go run ./cmd/api -db-dsn="$GATEKEEPER_DB_DSN" -quiet-delay=30s 2>&1 | tee -a experiments/raw/server.jsonl
```

Use that command in the server terminal and leave it running while you collect condition A runs. For other conditions, stop and restart with the corresponding `-quiet-delay`, `-simulation-duration`, and `-simulation-interval` flags above, keeping the same append-only `experiments/raw/server.jsonl` file. Normal use uses `make run/api` with defaults. Use direct `go run` for evidence capture so Make's echoed shell command does not contaminate the JSON Lines log.

The `experiment workload configuration` server log records the actual settings at startup. `job committed` records contain job ID, version, status, real/simulated progress and the post-commit wall-clock timestamp. `observation database reads` records count SQL statements for the GET observation handler. Long-poll waiting itself uses in-process broadcast notifications and does not periodically query PostgreSQL; it reads the durable snapshot initially, after a wake-up, and again at the holding deadline before a 204. The worker's independent queue polling is not an observation query.

## Capture and analyze

After a run, use **Download raw client logs** in the page. Combine/download traces for all runs without editing event timestamps. Keep the file as `experiments/raw/client.jsonl`; the browser log includes request starts, response receipt, HTTP status, returned version, render time, cancellations, errors, run metadata, and wall-clock epoch milliseconds. The server log is JSON Lines and can be joined by job ID and terminal version.

Once the run set is complete:

```sh
python3 experiments/analyze.py \
  --client experiments/raw/client.jsonl \
  --server experiments/raw/server.jsonl \
  --out-dir experiments/results
```

Generated outputs:

- `experiments/results/summary.csv` — every run and the required metrics (plus `observation_db_reads`).
- `experiments/results/summary.md` — per-run rows, median and minimum–maximum by condition/mode, and run-count warnings.
- `experiments/results/request-count-by-condition.svg` — labelled median request-count chart.
- `experiments/results/detection-delay-by-condition.svg` — labelled median completion-detection-delay chart in milliseconds.

Detection delay uses browser `Date.now()` at terminal response receipt minus the server's `post_commit_at` log timestamp for the same job/version. This is valid only for the specified same-computer run; it remains an estimate affected by wall-clock resolution and scheduling. The script leaves detection blank if it cannot join the records and flags negative results instead of silently changing them. It also notes started requests with no response/error/cancellation record. Investigate missing logs, pending requests, version gaps, and unexpected outcomes before writing the report. Never edit raw logs to make runs look complete.
