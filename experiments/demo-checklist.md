# Assessment 2 demonstration checklist (20–30 minutes)

- [ ] State baseline commit/tag, system topology, and that PostgreSQL owns job state while the worker performs processing.
- [ ] Show short-poll mode selected before upload; show job ID, status, version, real variant progress, and separate GETs in DevTools.
- [ ] Show long-poll mode selected before a new run; demonstrate an update waking a held request.
- [ ] Demonstrate a committed version change between requests and show the next `after=<last version>` request returns promptly with the newer snapshot.
- [ ] Let an unchanged long poll reach normal 204; show no JSON parse/error and immediate follow-up request.
- [ ] Show completed and failed jobs stop observation; reopening a terminal job returns promptly.
- [ ] Stop observation or switch modes while a request is pending; show that the original worker continues progressing and no second POST is made.
- [ ] Simulate a retrieval failure; show last known state preserved and **Try again** resumes the same job without uploading again.
- [ ] Enable a simulated-update run visibly and show the simulated label/count is separate from real variant progress; then disable controls for ordinary use.
- [ ] Show two tabs both discovering the same committed update (correctness check, outside the single-observer measurement runs).
- [ ] Show client raw-log download, server JSON Lines records, generated CSV/tables/charts, predictions, and report.
- [ ] Explain request count versus database activity, completion-detection delay versus processing duration, version gaps, limitations, and recommendation using actual run IDs/results.
