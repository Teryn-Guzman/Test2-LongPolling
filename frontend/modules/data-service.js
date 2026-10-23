const acceptedStatus = 202;

export async function uploadImage(file) {
	const body = new FormData();
	body.append('image', file);

	const response = await fetch('/v1/images', { method: 'POST', body });
	const data = await readJson(response);
	if (!response.ok || response.status !== acceptedStatus) {
		throw new Error(data.error || 'The image could not be accepted.');
	}

	return data;
}

export async function getJob(statusUrl, signal, after = null) {
	const url = new URL(statusUrl, window.location.origin);
	if (after !== null) url.searchParams.set('after', String(after));
	const response = await fetch(url, { signal, cache: 'no-store' });
	if (response.status === 204) {
		return { status: 204, job: null };
	}
	if (!response.ok) {
		const error = new Error(`Status request failed with HTTP ${response.status}.`);
		error.status = response.status;
		throw error;
	}
	try {
		return { status: response.status, job: await response.json() };
	} catch {
		throw new Error('The status response was not valid JSON.');
	}
}

async function readJson(response) {
	return response.json().catch(() => ({}));
}
