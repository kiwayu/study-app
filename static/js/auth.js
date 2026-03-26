// ── Authentication module ────────────────────────────────────

let _user = null;
let _refreshing = null;

/**
 * Check if user is authenticated by calling GET /api/me.
 * On 401, attempt a single token refresh. Returns user object or null.
 */
export async function checkAuth() {
  try {
    const res = await fetch('/api/me', { credentials: 'same-origin' });
    if (res.ok) {
      _user = await res.json();
      return _user;
    }
    if (res.status === 401) {
      // Try refreshing the token
      const refreshed = await tryRefresh();
      if (refreshed) {
        const retry = await fetch('/api/me', { credentials: 'same-origin' });
        if (retry.ok) {
          _user = await retry.json();
          return _user;
        }
      }
    }
  } catch {
    // Network error — treat as not authenticated
  }
  _user = null;
  return null;
}

/**
 * Attempt to refresh the auth token. Returns true on success.
 * Deduplicates concurrent refresh attempts.
 */
export async function tryRefresh() {
  if (_refreshing) return _refreshing;
  _refreshing = (async () => {
    try {
      const res = await fetch('/auth/refresh', {
        method: 'POST',
        credentials: 'same-origin',
      });
      return res.ok;
    } catch {
      return false;
    } finally {
      _refreshing = null;
    }
  })();
  return _refreshing;
}

/**
 * Return the cached user object or null.
 */
export function getUser() {
  return _user;
}

/**
 * Log out: POST /auth/logout then reload the page.
 */
export async function logout() {
  try {
    await fetch('/auth/logout', {
      method: 'POST',
      credentials: 'same-origin',
    });
  } catch {
    // Ignore errors — we reload regardless
  }
  window.location.reload();
}

/**
 * Show the login screen, hide the app shell.
 */
export function showLoginPage() {
  const login = document.getElementById('login-screen');
  const app = document.querySelector('.app-shell');
  if (login) login.style.display = '';
  if (app) app.style.display = 'none';
}

/**
 * Hide the login screen, show the app shell and user profile.
 */
export function hideLoginPage() {
  const login = document.getElementById('login-screen');
  const app = document.querySelector('.app-shell');
  if (login) login.style.display = 'none';
  if (app) app.style.display = '';
}

/**
 * Populate the user profile UI (avatar + name).
 */
export function renderUserProfile(user) {
  const profile = document.getElementById('user-profile');
  const avatar = document.getElementById('user-avatar');
  if (!profile || !user) return;
  if (avatar && user.picture) {
    avatar.src = user.picture;
    avatar.alt = user.name || 'Profile';
  }
  profile.style.display = '';
}
