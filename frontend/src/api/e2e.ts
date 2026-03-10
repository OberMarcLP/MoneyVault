/**
 * E2E encryption interceptors for the API client.
 * Encrypts sensitive fields before sending and decrypts after receiving.
 */
import { useCryptoStore } from '@/stores/crypto';
import { encryptField, decryptField } from '@/utils/crypto';

interface FieldMap {
  [key: string]: string[];
}

// Maps endpoint patterns to their encrypted fields
const ENCRYPTED_FIELDS: FieldMap = {
  '/accounts': ['name', 'balance'],
  '/transactions': ['amount', 'description'],
};

function matchEndpoint(endpoint: string): string[] | null {
  for (const [pattern, fields] of Object.entries(ENCRYPTED_FIELDS)) {
    if (endpoint.startsWith(pattern)) {
      return fields;
    }
  }
  return null;
}

export async function encryptRequestBody(endpoint: string, body: string | undefined): Promise<string | undefined> {
  if (!body) return body;
  const { dek, isE2EEnabled } = useCryptoStore.getState();
  if (!isE2EEnabled || !dek) return body;

  const fields = matchEndpoint(endpoint);
  if (!fields) return body;

  try {
    const parsed = JSON.parse(body);
    for (const field of fields) {
      if (parsed[field] && typeof parsed[field] === 'string') {
        parsed[field] = await encryptField(parsed[field], dek);
      }
    }
    return JSON.stringify(parsed);
  } catch {
    return body;
  }
}

async function decryptObject(obj: Record<string, unknown>, fields: string[], dek: Uint8Array): Promise<void> {
  for (const field of fields) {
    if (obj[field] && typeof obj[field] === 'string') {
      try {
        obj[field] = await decryptField(obj[field] as string, dek);
      } catch {
        // Field may not be encrypted (e.g. newly created while E2E was being set up)
      }
    }
  }
}

export async function decryptResponseData<T>(endpoint: string, data: T): Promise<T> {
  const { dek, isE2EEnabled } = useCryptoStore.getState();
  if (!isE2EEnabled || !dek) return data;

  const fields = matchEndpoint(endpoint);
  if (!fields) return data;

  try {
    if (Array.isArray(data)) {
      for (const item of data) {
        if (item && typeof item === 'object') {
          await decryptObject(item as Record<string, unknown>, fields, dek);
        }
      }
    } else if (data && typeof data === 'object') {
      // Handle paginated responses (e.g. { transactions: [...], total: 10 })
      const obj = data as Record<string, unknown>;
      for (const key of Object.keys(obj)) {
        if (Array.isArray(obj[key])) {
          for (const item of obj[key] as Record<string, unknown>[]) {
            if (item && typeof item === 'object') {
              await decryptObject(item, fields, dek);
            }
          }
        }
      }
      // Also decrypt the top-level object itself
      await decryptObject(obj, fields, dek);
    }
  } catch {
    // Silently fail decryption errors
  }

  return data;
}
