# ImageLab Assessment 2 — Long-polling implementation and comparison

_Replace prompts only with verified implementation details and measured evidence. Do not claim a conclusion before completing the runs._

## 1. Predictions and method

**Predictions (written before measurement):** Summarize the five predictions recorded in `predictions.md`, noting the date they were written.

**Setup:** Record the Assessment 1 baseline commit/tag; Go/PostgreSQL/browser versions; OS; exact input image and dimensions; database and server location; number of observers; how the worker queue was confirmed empty; tab-foreground procedure; server flags; and how the actual C3 latency was configured and prompt-GET checked.

**Procedure:** State the run order, alternation of short/long first, repetitions per condition, and observation start/end rules. Define short polling as a one-second delay after each response (or update this to actual verified behavior). Long polling uses the default 20-second holding deadline and starts the next request immediately after valid active-state `200` or `204`.

**Timestamp/method:** The reported detection delay is browser `Date.now()` at full terminal-representation receipt minus the same job/version's server `post_commit_at` wall-clock log. Runs are on the same computer. Discuss log timestamp resolution, parse/join procedure, and scheduling uncertainty. Explain that processing duration uses the job's persisted `started_at` to `completed_at`/`failed_at`. State that long-poll waiting is notification-based, list observation SQL reads from logs, and distinguish worker queue polling from observation queries.

## 2. Results

Insert every row from `experiments/results/summary.csv` and the grouped median/minimum–maximum table from `summary.md`. Include request classifications, pending/canceled/error requests, processing duration, detection delay, version coverage/gaps, and observation DB reads.

![Request count by condition](results/request-count-by-condition.svg)

![Completion-detection delay by condition (ms)](results/detection-delay-by-condition.svg)

All source measurements are retained in raw JSONL logs. Do not remove outliers; explain unexpected or missing joins.

## 3. Interpretation and recommendation

Answer with evidence references to run IDs/conditions:

- Where did long polling help, and where did it not?
- Did image processing become faster, or did observation return the terminal state sooner?
- Did fewer HTTP requests correspond to fewer PostgreSQL reads? Include held handlers and notification bookkeeping.
- What do version gaps mean for this snapshot API? Why can intermediate snapshots be skipped without losing the final durable state? Would that be acceptable if every event had to be retained?
- What must the browser do after every long-poll response? Explain how a continuing response could send multiple updates and what trade-off it changes.
- Recommendation for ordinary ImageLab use, with limitations and why the evidence supports it.

## 4. Limitations and conclusion

Discuss the small sample size (classroom comparison only), single-computer clock/scheduling uncertainty, foreground-tab/browser timer effects, single observer in measured runs, controlled artificial workload versus real image processing, and any actual latency deviation from 200 ms. Report only measurements actually collected.
