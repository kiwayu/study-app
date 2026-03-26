const BASE = '/api';

async function request(method, path, body) {
  try {
    const opts = {
      method,
      headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
      body: body !== undefined ? JSON.stringify(body) : undefined,
    };
    const res = await fetch(BASE + path, opts);
    if (res.status === 204) return { data: null, error: null };
    const json = await res.json();
    if (!res.ok) return { data: null, error: json.error ?? `HTTP ${res.status}` };
    return { data: json, error: null };
  } catch (err) {
    return { data: null, error: err.message };
  }
}

export const getTasks        = ()       => request('GET',    '/tasks');
export const createTask      = (body)   => request('POST',   '/tasks', body);
export const updateTask      = (id, b)  => request('PUT',    `/tasks/${id}`, b);
export const deleteTask      = (id)     => request('DELETE', `/tasks/${id}`);
export const getSettings     = ()       => request('GET',    '/settings');
export const updateSettings  = (body)   => request('PUT',    '/settings', body);
export const getSession      = ()       => request('GET',    '/session');
export const startSession    = (body)   => request('POST',   '/session/start', body);
export const pauseSession    = ()       => request('POST',   '/session/pause');
export const stopSession     = ()       => request('POST',   '/session/stop');
export const updateTotals    = (body)   => request('PUT',    '/session/totals', body);
export const getCompletions      = ()       => request('GET',    '/stats/completions');
export const getEstimationStats  = ()       => request('GET',    '/stats/estimation');
