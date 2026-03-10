const CACHE_NAME = 'moneyvault-v2';
const API_CACHE = 'moneyvault-api-v1';
const STATIC_ASSETS = [
  '/',
  '/manifest.json',
  '/icons/icon-192.svg',
  '/icons/icon-512.svg',
];

// API endpoints to cache for offline access (read-only GET endpoints)
const CACHEABLE_API = [
  '/api/v1/accounts',
  '/api/v1/categories',
  '/api/v1/transactions',
  '/api/v1/budgets',
  '/api/v1/investments',
  '/api/v1/investments/summary',
  '/api/v1/crypto/summary',
  '/api/v1/notifications',
  '/api/v1/notifications/count',
  '/api/v1/analytics/net-worth',
  '/api/v1/analytics/spending',
  '/api/v1/analytics/asset-allocation',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((k) => k !== CACHE_NAME && k !== API_CACHE)
          .map((k) => caches.delete(k))
      )
    )
  );
  self.clients.claim();
});

// Push notification handler
self.addEventListener('push', (event) => {
  let data = { title: 'MoneyVault', body: 'You have a new notification', url: '/' };
  try {
    if (event.data) {
      data = { ...data, ...event.data.json() };
    }
  } catch (e) {
    // Use defaults
  }

  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body,
      icon: '/icons/icon-192.svg',
      badge: '/icons/icon-192.svg',
      data: { url: data.url },
      vibrate: [100, 50, 100],
    })
  );
});

// Notification click handler
self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  const url = event.notification.data?.url || '/';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      for (const client of windowClients) {
        if (new URL(client.url).origin === self.location.origin) {
          client.navigate(url);
          return client.focus();
        }
      }
      return clients.openWindow(url);
    })
  );
});

// Background sync — replay queued mutations when back online
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-mutations') {
    event.waitUntil(replayMutations());
  }
});

async function replayMutations() {
  try {
    const cache = await caches.open('moneyvault-mutations');
    const requests = await cache.keys();
    for (const req of requests) {
      const response = await cache.match(req);
      if (!response) continue;
      const body = await response.text();
      try {
        await fetch(req.url, {
          method: req.method,
          headers: Object.fromEntries(req.headers.entries()),
          body: body || undefined,
        });
        await cache.delete(req);
      } catch {
        // Will retry on next sync
        break;
      }
    }
  } catch {
    // Ignore errors
  }
}

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Non-GET requests: try network, queue failures for sync
  if (request.method !== 'GET') {
    if (url.pathname.startsWith('/api/')) {
      event.respondWith(
        fetch(request.clone()).catch(async () => {
          if ('sync' in self.registration) {
            const cache = await caches.open('moneyvault-mutations');
            await cache.put(request, new Response(await request.clone().text()));
            await self.registration.sync.register('sync-mutations');
          }
          return new Response(JSON.stringify({ error: 'Offline. Changes will sync when you reconnect.' }), {
            status: 503,
            headers: { 'Content-Type': 'application/json' },
          });
        })
      );
    }
    return;
  }

  // API calls: network-first with cache fallback for cacheable endpoints
  if (url.pathname.startsWith('/api/')) {
    const isCacheable = CACHEABLE_API.some((path) => url.pathname.startsWith(path));
    if (!isCacheable) return;

    event.respondWith(
      fetch(request.clone())
        .then((response) => {
          if (response.ok) {
            const clone = response.clone();
            caches.open(API_CACHE).then((cache) => cache.put(request, clone));
          }
          return response;
        })
        .catch(() => caches.match(request).then((cached) => {
          if (cached) return cached;
          return new Response(JSON.stringify({ error: 'Offline' }), {
            status: 503,
            headers: { 'Content-Type': 'application/json' },
          });
        }))
    );
    return;
  }

  // Static assets: cache-first with offline fallback
  event.respondWith(
    caches.match(request).then((cached) => {
      if (cached) return cached;
      return fetch(request).then((response) => {
        if (response.ok && (url.pathname.match(/\.(js|css|svg|png|jpg|woff2?)$/) || url.pathname === '/')) {
          const clone = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(request, clone));
        }
        return response;
      }).catch(() => {
        // Serve cached index.html for navigation requests when offline
        if (request.mode === 'navigate') {
          return caches.match('/');
        }
        return new Response('Offline', { status: 503 });
      });
    })
  );
});
