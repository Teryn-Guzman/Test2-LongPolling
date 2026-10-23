const STORAGE_KEY = 'imagelab-assessment2-client-log-v1';

function readLog() {
	try {
		return JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
	} catch {
		return [];
	}
}

function saveLog(entries) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(entries));
	} catch (error) {
		console.warn('Could not persist the ImageLab measurement log.', error);
	}
}

function append(run, event, fields = {}) {
	const entries = readLog();
	entries.push({
		run_id: run.id,
		condition: run.condition,
		repetition: run.repetition,
		mode: run.mode,
		job_id: run.jobId || '',
		image: run.image,
		latency_setting: run.latencySetting,
		event,
		client_time_ms: Date.now(),
		performance_time_ms: performance.now(),
		...fields,
	});
	saveLog(entries);
}

export function beginRun(metadata) {
	return {
		id: globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`,
		condition: metadata.condition,
		repetition: metadata.repetition,
		mode: metadata.mode,
		image: metadata.image,
		latencySetting: metadata.latencySetting,
		jobId: '',
		baselineVersion: null,
		lastVersion: null,
		terminalVersion: null,
		distinctNewVersions: new Set(),
		requestsStarted: 0,
		initial200: 0,
		newer200: 0,
		unchanged200: 0,
		noChange204: 0,
		retrievalErrors: 0,
		cancellations: 0,
		pendingRequests: new Set(),
		terminalReceiptMs: null,
		processingMs: null,
		requestStartTimes: new Map(),
	};
}

export function setRunJob(run, jobId) {
	run.jobId = jobId;
	append(run, 'job_accepted');
}

export function setRunMode(run, mode) {
	if (!run || run.mode === mode) return;
	const previousMode = run.mode;
	run.mode = mode;
	append(run, 'mode_changed', { previous_mode: previousMode });
}

export function requestStarted(run, requestId, url, kind) {
	run.requestsStarted += 1;
	run.pendingRequests.add(requestId);
	const startedPerformanceMs = performance.now();
	run.requestStartTimes.set(requestId, startedPerformanceMs);
	append(run, 'request_start', { request_id: requestId, url, request_kind: kind, started_performance_ms: startedPerformanceMs });
}

export function responseReceived(run, requestId, status, job, kind) {
	run.pendingRequests.delete(requestId);
	const version = Number.isInteger(job?.version) ? job.version : null;
	const startedPerformanceMs = run.requestStartTimes.get(requestId);
	const durationMs = startedPerformanceMs === undefined ? null : performance.now() - startedPerformanceMs;
	let classification = 'non_200';
	if (status === 204) {
		run.noChange204 += 1;
		classification = 'no_change_204';
	} else if (status === 200 && job) {
		if (run.baselineVersion === null) {
			run.initial200 += 1;
			run.baselineVersion = version;
			classification = 'initial_200';
		} else if (version !== null && run.lastVersion !== null && version > run.lastVersion) {
			run.newer200 += 1;
			run.distinctNewVersions.add(version);
			classification = 'newer_200';
		} else {
			run.unchanged200 += 1;
			classification = 'unchanged_200';
		}
		if (version !== null) run.lastVersion = Math.max(run.lastVersion ?? version, version);
		if (job.status === 'completed' || job.status === 'failed') {
			run.terminalVersion = version;
			run.terminalReceiptMs = Date.now();
			const terminalAt = job.completed_at || job.failed_at;
			if (job.started_at && terminalAt) {
				run.processingMs = Date.parse(terminalAt) - Date.parse(job.started_at);
			}
		}
	}
	append(run, 'response_received', {
		request_id: requestId,
		http_status: status,
		received_version: version,
		job_status: job?.status || '',
		stage: job?.stage || '',
		variants_completed: job?.variants_completed ?? null,
		classification,
		request_kind: kind,
		request_duration_ms: durationMs,
		started_at: job?.started_at || null,
		completed_at: job?.completed_at || null,
		failed_at: job?.failed_at || null,
	});
}

export function responseRendered(run, requestId, version) {
	append(run, 'rendered', { request_id: requestId, rendered_version: version });
}

export function retrievalFailed(run, requestId, detail, httpStatus = null) {
	run.pendingRequests.delete(requestId);
	run.retrievalErrors += 1;
	const startedPerformanceMs = run.requestStartTimes.get(requestId);
	append(run, 'retrieval_error', {
		request_id: requestId,
		detail: String(detail),
		http_status: httpStatus,
		request_duration_ms: startedPerformanceMs === undefined ? null : performance.now() - startedPerformanceMs,
	});
}

export function requestCancelled(run, requestId, reason) {
	if (!run.pendingRequests.has(requestId)) return;
	run.pendingRequests.delete(requestId);
	run.cancellations += 1;
	const startedPerformanceMs = run.requestStartTimes.get(requestId);
	append(run, 'request_cancelled', {
		request_id: requestId,
		reason,
		request_duration_ms: startedPerformanceMs === undefined ? null : performance.now() - startedPerformanceMs,
	});
}

export function downloadClientLog() {
	const entries = readLog();
	const contents = entries.map(entry => JSON.stringify(entry)).join('\n');
	const blob = new Blob([contents ? `${contents}\n` : ''], { type: 'application/x-ndjson' });
	const url = URL.createObjectURL(blob);
	const anchor = document.createElement('a');
	anchor.href = url;
	anchor.download = 'imagelab-client-observations.jsonl';
	anchor.click();
	URL.revokeObjectURL(url);
}