# Assessment 2 experiment procedure and evidence

This directory holds the Assessment 2 evidence: `predictions.md` (written and committed before any measurement), the raw browser and server logs in `raw/`, run notes, and the analyzer output in `results/`. The procedure below is the one that was followed: fill in predictions before each condition, collect actual browser/server logs, then run the analyzer. It creates the CSV, per-run and grouped summary tables, and labelled SVG charts from those logs only. The analyzer needs Python 3.11 or newer (`python3`, or `py` on Windows).

## Before running

1. Record the Assessment 1 baseline: Assessment 2 started from commit `e03682d055cf31a32cfca75103fde4b26a959c52` (tag `assessment1-baseline`).
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

The browser keeps its log in localStorage across page reloads, so it was cleared once before the first measured run and never between conditions. Every server start used `tee -a experiments/raw/server.jsonl` (append) so one server log covers all runs. After the last run, use **Download raw client logs** in the page. The download contains every run recorded so far; do not edit event timestamps. Keep the file as `experiments/raw/client.jsonl`; the browser log includes request starts, response receipt, HTTP status, returned version, render time, cancellations, errors, run metadata, and wall-clock epoch milliseconds. The server log is JSON Lines and can be joined by job ID and terminal version.

Once the run set is complete:

```sh
python3 experiments/analyze.py \
  --client experiments/raw/client.jsonl \
  --server experiments/raw/server.jsonl \
  --out-dir experiments/results
```

Add `--exclude-job JOB_ID_PREFIX` (at least 8 characters, repeatable) to leave a run out of the analysis. It must match exactly one run, the raw log is never modified, and the excluded runs are listed in the Input check of `summary.md`. The exact command used for the submitted results is under **Runs collected** below.

Generated outputs:

- `experiments/results/summary.csv` — every run and the required metrics, plus `observation_db_reads`, `artificial_delay_ms` and `real_processing_ms`.
- `experiments/results/summary.md` — an input check (records read, non-JSON lines skipped, excluded runs), a workload check (each run's server flags and latency note against its condition label), per-run rows, median and minimum–maximum by condition/mode, and run-count warnings.
- `experiments/results/request-count-by-condition.svg` — labelled median request-count chart.
- `experiments/results/detection-delay-by-condition.svg` — labelled median completion-detection-delay chart in milliseconds.

Non-JSON lines in either log (for example a stray `go run` message) are skipped and counted rather than stopping the analysis. A run whose server flags or latency note disagree with its condition label is reported as `CONFIG MISMATCH` in its notes and in the workload check. `artificial_delay_ms` is the configured quiet and image-processing delay plus the measured length of the simulated phase; `real_processing_ms` is `processing_ms` minus that, and still includes the short claim-to-finish overhead.

The browser stamps `Date.now()` as soon as the full response body has been read, before any logging work, and keeps its log in memory (it is written to localStorage after a terminal response, an error, or when the page is hidden), so storage work does not delay the timestamp. Detection delay uses browser `Date.now()` at terminal response receipt minus the server's `post_commit_at` log timestamp for the same job/version. This is valid only for the specified same-computer run; it remains an estimate affected by wall-clock resolution and scheduling. The script leaves detection blank if it cannot join the records and flags negative results instead of silently changing them. It also notes started requests with no response/error/cancellation record. Investigate missing logs, pending requests, version gaps, and unexpected outcomes before writing the report. Never edit raw logs to make runs look complete.

## Runs collected

The evidence is 30 measured runs: 5 conditions × 2 modes × 3 repetitions, one observer, the same image (`picture1.jpg`), on one computer. The mode order alternated by repetition: short then long (repetition 1), long then short (repetition 2), short then long (repetition 3). Server flags per condition: A `-quiet-delay=30s`; B none; C1 `-simulation-duration=10s -simulation-interval=2s`; C2 and C3 `-simulation-duration=10s -simulation-interval=200ms`. C3 also used a Chrome DevTools custom network profile with 200 ms latency and no bandwidth limit (latency note: `Chrome custom profile, 200 ms latency, no bandwidth limit`). Measured prompt-GET duration in C3 was 205–220 ms (median 215 ms) against 13–22 ms (median 16 ms) with no throttling in B, an added delay of about 199 ms.

Two earlier attempts at condition A are in `client.jsonl` but are excluded from the analysis: A short repetition 1 (job `e9c04638-e6f5-4b3a-934f-dee8b1eee02e`, 38 client events) and A long repetition 1 (job `28d43985-136f-4b46-9db7-46455e1ce3c8`, 16 client events). They were recorded before server log capture began, so `server.jsonl` has no records for either job, and their client traces are shorter than the completed A runs of the same mode (94 and 18 events). Both were repeated as the measured A repetition 1 runs. The raw log is unchanged.

The submitted results were produced with:

```sh
python3 experiments/analyze.py \
  --client experiments/raw/client.jsonl \
  --server experiments/raw/server.jsonl \
  --out-dir experiments/results \
  --exclude-job e9c04638 \
  --exclude-job 28d43985
```

## Checking the actual C3 latency

The 200 ms profile is nominal, so report what was measured. This prints the duration of each C3 run's first (prompt) request from the raw client log; a normal localhost request takes only a few milliseconds:

```sh
python3 -c "import json;[print(e['mode'],e['repetition'],round(e['request_duration_ms'])) for e in map(json.loads,open('experiments/raw/client.jsonl')) if e.get('condition')=='C3' and e.get('event')=='response_received' and e.get('request_kind')=='initial']"
```