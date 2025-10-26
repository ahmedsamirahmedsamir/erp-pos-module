// Service Worker for LinearBits POS Terminal
// Provides offline capabilities and caching for POS operations

const CACHE_NAME = 'pos-terminal-v1';
const OFFLINE_CACHE_NAME = 'pos-offline-v1';
const API_CACHE_NAME = 'pos-api-v1';

// Files to cache for offline functionality
const STATIC_CACHE_URLS = [
  '/pos/terminal',
  '/pos/static/css/main.css',
  '/pos/static/js/main.js',
  '/pos/static/js/pos-terminal.js',
  '/pos/static/js/offline-sync.js',
  '/pos/icons/icon-192x192.png',
  '/pos/icons/icon-512x512.png',
  '/pos/manifest.json'
];

// API endpoints to cache
const API_CACHE_URLS = [
  '/api/v1/pos/products',
  '/api/v1/pos/registers',
  '/api/v1/pos/taxes',
  '/api/v1/pos/discounts',
  '/api/v1/customers',
  '/api/v1/inventory/products'
];

// Install event - cache static resources
self.addEventListener('install', (event) => {
  console.log('POS Service Worker: Installing...');
  
  event.waitUntil(
    Promise.all([
      caches.open(CACHE_NAME).then((cache) => {
        console.log('POS Service Worker: Caching static files');
        return cache.addAll(STATIC_CACHE_URLS);
      }),
      caches.open(API_CACHE_NAME).then((cache) => {
        console.log('POS Service Worker: Caching API endpoints');
        return cache.addAll(API_CACHE_URLS);
      })
    ]).then(() => {
      console.log('POS Service Worker: Installation complete');
      return self.skipWaiting();
    })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('POS Service Worker: Activating...');
  
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME && 
              cacheName !== OFFLINE_CACHE_NAME && 
              cacheName !== API_CACHE_NAME) {
            console.log('POS Service Worker: Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => {
      console.log('POS Service Worker: Activation complete');
      return self.clients.claim();
    })
  );
});

// Fetch event - serve from cache or network
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);
  
  // Handle API requests
  if (url.pathname.startsWith('/api/v1/pos/')) {
    event.respondWith(handleAPIRequest(request));
    return;
  }
  
  // Handle static file requests
  if (url.pathname.startsWith('/pos/') || url.pathname.startsWith('/static/')) {
    event.respondWith(handleStaticRequest(request));
    return;
  }
  
  // Handle navigation requests
  if (request.mode === 'navigate') {
    event.respondWith(handleNavigationRequest(request));
    return;
  }
});

// Handle API requests with offline support
async function handleAPIRequest(request) {
  const url = new URL(request.url);
  
  try {
    // Try network first for API requests
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(API_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('POS Service Worker: Network failed, trying cache:', url.pathname);
    
    // Fallback to cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // For POST requests (transactions), store for later sync
    if (request.method === 'POST') {
      return handleOfflineTransaction(request);
    }
    
    // Return offline response
    return new Response(
      JSON.stringify({ 
        error: 'Offline', 
        message: 'No internet connection. Data will sync when online.' 
      }),
      { 
        status: 503, 
        headers: { 'Content-Type': 'application/json' } 
      }
    );
  }
}

// Handle static file requests
async function handleStaticRequest(request) {
  try {
    // Try cache first for static files
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Fallback to network
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('POS Service Worker: Failed to fetch static file:', request.url);
    
    // Return offline page for HTML requests
    if (request.headers.get('accept').includes('text/html')) {
      return caches.match('/pos/offline.html');
    }
    
    throw error;
  }
}

// Handle navigation requests
async function handleNavigationRequest(request) {
  try {
    const networkResponse = await fetch(request);
    return networkResponse;
  } catch (error) {
    // Return cached offline page
    const cachedResponse = await caches.match('/pos/offline.html');
    return cachedResponse || new Response('Offline', { status: 503 });
  }
}

// Handle offline transactions
async function handleOfflineTransaction(request) {
  try {
    const requestData = await request.clone().json();
    
    // Store transaction in IndexedDB for later sync
    await storeOfflineTransaction(requestData);
    
    // Return success response
    return new Response(
      JSON.stringify({ 
        success: true, 
        message: 'Transaction saved offline. Will sync when online.',
        offline: true 
      }),
      { 
        status: 200, 
        headers: { 'Content-Type': 'application/json' } 
      }
    );
  } catch (error) {
    console.error('POS Service Worker: Failed to store offline transaction:', error);
    return new Response(
      JSON.stringify({ 
        error: 'Failed to save offline transaction' 
      }),
      { 
        status: 500, 
        headers: { 'Content-Type': 'application/json' } 
      }
    );
  }
}

// Store offline transaction in IndexedDB
async function storeOfflineTransaction(transactionData) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('POSOfflineDB', 1);
    
    request.onerror = () => reject(request.error);
    
    request.onsuccess = () => {
      const db = request.result;
      const transaction = db.transaction(['offlineTransactions'], 'readwrite');
      const store = transaction.objectStore('offlineTransactions');
      
      const offlineTransaction = {
        id: Date.now() + Math.random(),
        data: transactionData,
        timestamp: new Date().toISOString(),
        synced: false
      };
      
      const addRequest = store.add(offlineTransaction);
      
      addRequest.onsuccess = () => {
        console.log('POS Service Worker: Offline transaction stored');
        resolve();
      };
      
      addRequest.onerror = () => reject(addRequest.error);
    };
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      
      if (!db.objectStoreNames.contains('offlineTransactions')) {
        const store = db.createObjectStore('offlineTransactions', { 
          keyPath: 'id', 
          autoIncrement: true 
        });
        store.createIndex('timestamp', 'timestamp', { unique: false });
        store.createIndex('synced', 'synced', { unique: false });
      }
    };
  });
}

// Background sync for offline transactions
self.addEventListener('sync', (event) => {
  if (event.tag === 'pos-sync') {
    console.log('POS Service Worker: Background sync triggered');
    event.waitUntil(syncOfflineTransactions());
  }
});

// Sync offline transactions when back online
async function syncOfflineTransactions() {
  try {
    const db = await openIndexedDB();
    const transaction = db.transaction(['offlineTransactions'], 'readwrite');
    const store = transaction.objectStore('offlineTransactions');
    const index = store.index('synced');
    
    const unsyncedRequest = index.getAll(false);
    
    unsyncedRequest.onsuccess = async () => {
      const unsyncedTransactions = unsyncedRequest.result;
      
      for (const offlineTransaction of unsyncedTransactions) {
        try {
          // Attempt to sync the transaction
          const response = await fetch('/api/v1/pos/transactions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(offlineTransaction.data)
          });
          
          if (response.ok) {
            // Mark as synced
            offlineTransaction.synced = true;
            store.put(offlineTransaction);
            console.log('POS Service Worker: Transaction synced:', offlineTransaction.id);
          }
        } catch (error) {
          console.error('POS Service Worker: Failed to sync transaction:', error);
        }
      }
    };
  } catch (error) {
    console.error('POS Service Worker: Failed to sync offline transactions:', error);
  }
}

// Helper function to open IndexedDB
function openIndexedDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('POSOfflineDB', 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      
      if (!db.objectStoreNames.contains('offlineTransactions')) {
        const store = db.createObjectStore('offlineTransactions', { 
          keyPath: 'id', 
          autoIncrement: true 
        });
        store.createIndex('timestamp', 'timestamp', { unique: false });
        store.createIndex('synced', 'synced', { unique: false });
      }
    };
  });
}

// Push notification handling for POS alerts
self.addEventListener('push', (event) => {
  console.log('POS Service Worker: Push notification received');
  
  const options = {
    body: event.data ? event.data.text() : 'New POS notification',
    icon: '/pos/icons/icon-192x192.png',
    badge: '/pos/icons/badge-72x72.png',
    vibrate: [100, 50, 100],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'explore',
        title: 'View Details',
        icon: '/pos/icons/checkmark.png'
      },
      {
        action: 'close',
        title: 'Close',
        icon: '/pos/icons/xmark.png'
      }
    ]
  };
  
  event.waitUntil(
    self.registration.showNotification('POS Terminal', options)
  );
});

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('POS Service Worker: Notification clicked');
  
  event.notification.close();
  
  if (event.action === 'explore') {
    event.waitUntil(
      clients.openWindow('/pos/terminal')
    );
  }
});

// Message handling from main thread
self.addEventListener('message', (event) => {
  console.log('POS Service Worker: Message received:', event.data);
  
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
  
  if (event.data && event.data.type === 'SYNC_TRANSACTIONS') {
    syncOfflineTransactions();
  }
  
  if (event.data && event.data.type === 'CACHE_PRODUCT') {
    cacheProductOffline(event.data.product);
  }
  
  if (event.data && event.data.type === 'CACHE_CUSTOMER') {
    cacheCustomerOffline(event.data.customer);
  }
});

// Cache product data for offline use
async function cacheProductOffline(product) {
  try {
    const cache = await caches.open(API_CACHE_NAME);
    const url = `/api/v1/products/${product.id}`;
    const response = new Response(JSON.stringify(product), {
      headers: { 'Content-Type': 'application/json' }
    });
    await cache.put(url, response);
    console.log('POS Service Worker: Product cached offline:', product.id);
  } catch (error) {
    console.error('POS Service Worker: Failed to cache product:', error);
  }
}

// Cache customer data for offline use
async function cacheCustomerOffline(customer) {
  try {
    const cache = await caches.open(API_CACHE_NAME);
    const url = `/api/v1/customers/${customer.id}`;
    const response = new Response(JSON.stringify(customer), {
      headers: { 'Content-Type': 'application/json' }
    });
    await cache.put(url, response);
    console.log('POS Service Worker: Customer cached offline:', customer.id);
  } catch (error) {
    console.error('POS Service Worker: Failed to cache customer:', error);
  }
}

// Periodic background sync for data updates
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'pos-data-refresh') {
    console.log('POS Service Worker: Periodic sync triggered');
    event.waitUntil(refreshCachedData());
  }
});

// Refresh cached data periodically
async function refreshCachedData() {
  try {
    const cache = await caches.open(API_CACHE_NAME);
    
    // Refresh product data
    const productResponse = await fetch('/api/v1/products');
    if (productResponse.ok) {
      await cache.put('/api/v1/products', productResponse.clone());
      console.log('POS Service Worker: Product data refreshed');
    }
    
    // Refresh customer data
    const customerResponse = await fetch('/api/v1/customers');
    if (customerResponse.ok) {
      await cache.put('/api/v1/customers', customerResponse.clone());
      console.log('POS Service Worker: Customer data refreshed');
    }
    
    // Refresh register data
    const registerResponse = await fetch('/api/v1/pos/registers');
    if (registerResponse.ok) {
      await cache.put('/api/v1/pos/registers', registerResponse.clone());
      console.log('POS Service Worker: Register data refreshed');
    }
    
  } catch (error) {
    console.error('POS Service Worker: Failed to refresh cached data:', error);
  }
}

// Handle barcode scanning offline
async function handleOfflineBarcodeScan(barcode) {
  try {
    // Try to find product in cached data
    const cache = await caches.open(API_CACHE_NAME);
    const cachedResponse = await cache.match('/api/v1/products');
    
    if (cachedResponse) {
      const products = await cachedResponse.json();
      const product = products.find(p => p.barcode === barcode);
      
      if (product) {
        return new Response(JSON.stringify({
          success: true,
          product: product,
          offline: true
        }), {
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }
    
    return new Response(JSON.stringify({
      success: false,
      message: 'Product not found in offline cache',
      offline: true
    }), {
      status: 404,
      headers: { 'Content-Type': 'application/json' }
    });
    
  } catch (error) {
    console.error('POS Service Worker: Failed to handle offline barcode scan:', error);
    return new Response(JSON.stringify({
      success: false,
      message: 'Offline barcode scan failed'
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

// Enhanced offline transaction handling with receipt generation
async function handleOfflineTransactionWithReceipt(request) {
  try {
    const requestData = await request.clone().json();
    
    // Store transaction in IndexedDB
    await storeOfflineTransaction(requestData);
    
    // Generate offline receipt
    const receiptData = {
      transactionId: `OFFLINE_${Date.now()}`,
      transactionNumber: requestData.transactionNumber || `OFF-${Date.now()}`,
      timestamp: new Date().toISOString(),
      items: requestData.items || [],
      total: requestData.totalAmount || 0,
      offline: true
    };
    
    // Store receipt for later printing
    await storeOfflineReceipt(receiptData);
    
    return new Response(JSON.stringify({
      success: true,
      message: 'Transaction saved offline with receipt',
      offline: true,
      receiptData: receiptData
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' }
    });
    
  } catch (error) {
    console.error('POS Service Worker: Failed to handle offline transaction with receipt:', error);
    return new Response(JSON.stringify({
      error: 'Failed to save offline transaction with receipt'
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

// Store offline receipt in IndexedDB
async function storeOfflineReceipt(receiptData) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('POSOfflineDB', 1);
    
    request.onerror = () => reject(request.error);
    
    request.onsuccess = () => {
      const db = request.result;
      const transaction = db.transaction(['offlineReceipts'], 'readwrite');
      const store = transaction.objectStore('offlineReceipts');
      
      const offlineReceipt = {
        id: Date.now() + Math.random(),
        data: receiptData,
        timestamp: new Date().toISOString(),
        printed: false
      };
      
      const addRequest = store.add(offlineReceipt);
      
      addRequest.onsuccess = () => {
        console.log('POS Service Worker: Offline receipt stored');
        resolve();
      };
      
      addRequest.onerror = () => reject(addRequest.error);
    };
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      
      if (!db.objectStoreNames.contains('offlineReceipts')) {
        const store = db.createObjectStore('offlineReceipts', { 
          keyPath: 'id', 
          autoIncrement: true 
        });
        store.createIndex('timestamp', 'timestamp', { unique: false });
        store.createIndex('printed', 'printed', { unique: false });
      }
    };
  });
}

// Enhanced push notifications with POS-specific actions
self.addEventListener('push', (event) => {
  console.log('POS Service Worker: Enhanced push notification received');
  
  let notificationData = {
    title: 'POS Terminal',
    body: 'New notification from POS system',
    icon: '/pos/icons/icon-192x192.png',
    badge: '/pos/icons/badge-72x72.png',
    vibrate: [200, 100, 200],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'view',
        title: 'View Details',
        icon: '/pos/icons/eye.png'
      },
      {
        action: 'print',
        title: 'Print Receipt',
        icon: '/pos/icons/printer.png'
      },
      {
        action: 'close',
        title: 'Close',
        icon: '/pos/icons/xmark.png'
      }
    ]
  };
  
  // Parse push data if available
  if (event.data) {
    try {
      const data = event.data.json();
      notificationData.title = data.title || notificationData.title;
      notificationData.body = data.body || notificationData.body;
      notificationData.data = { ...notificationData.data, ...data };
    } catch (error) {
      notificationData.body = event.data.text();
    }
  }
  
  event.waitUntil(
    self.registration.showNotification(notificationData.title, notificationData)
  );
});

// Enhanced notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('POS Service Worker: Enhanced notification clicked:', event.action);
  
  event.notification.close();
  
  switch (event.action) {
    case 'view':
      event.waitUntil(
        clients.openWindow('/pos/terminal')
      );
      break;
    case 'print':
      event.waitUntil(
        clients.matchAll().then(clients => {
          clients.forEach(client => {
            client.postMessage({
              type: 'PRINT_RECEIPT',
              data: event.notification.data
            });
          });
        })
      );
      break;
    case 'close':
      // Just close the notification
      break;
    default:
      // Default action - open the app
      event.waitUntil(
        clients.openWindow('/pos/terminal')
      );
  }
});

console.log('POS Service Worker: Enhanced version loaded successfully');