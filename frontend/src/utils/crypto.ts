/**
 * Client-side E2E encryption using Web Crypto API.
 * Uses PBKDF2 for key derivation and AES-256-GCM for encryption.
 */

const PBKDF2_ITERATIONS = 600_000;

function toBase64(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
  let binary = '';
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function fromBase64(b64: string): Uint8Array {
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

export function generateSalt(): string {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  return toBase64(salt);
}

export async function generateDEK(): Promise<Uint8Array> {
  return crypto.getRandomValues(new Uint8Array(32));
}

async function deriveKEK(password: string, saltB64: string): Promise<CryptoKey> {
  const encoder = new TextEncoder();
  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    encoder.encode(password),
    'PBKDF2',
    false,
    ['deriveKey'],
  );

  const salt = fromBase64(saltB64);

  return crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt: salt as BufferSource, iterations: PBKDF2_ITERATIONS, hash: 'SHA-256' },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt'],
  );
}

export async function encryptDEK(dek: Uint8Array, password: string, saltB64: string): Promise<string> {
  const kek = await deriveKEK(password, saltB64);
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    kek,
    dek as BufferSource,
  );

  const result = new Uint8Array(iv.length + encrypted.byteLength);
  result.set(iv);
  result.set(new Uint8Array(encrypted), iv.length);
  return toBase64(result);
}

export async function decryptDEK(encryptedDEKB64: string, password: string, saltB64: string): Promise<Uint8Array> {
  const kek = await deriveKEK(password, saltB64);
  const data = fromBase64(encryptedDEKB64);
  const iv = data.slice(0, 12);
  const ciphertext = data.slice(12);

  const decrypted = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv },
    kek,
    ciphertext,
  );

  return new Uint8Array(decrypted);
}

async function importDEK(dek: Uint8Array): Promise<CryptoKey> {
  return crypto.subtle.importKey(
    'raw',
    dek as BufferSource,
    'AES-GCM',
    false,
    ['encrypt', 'decrypt'],
  );
}

export async function encryptField(plaintext: string, dek: Uint8Array): Promise<string> {
  if (!plaintext) return '';
  const key = await importDEK(dek);
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const encoder = new TextEncoder();
  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    encoder.encode(plaintext),
  );

  const result = new Uint8Array(iv.length + encrypted.byteLength);
  result.set(iv);
  result.set(new Uint8Array(encrypted), iv.length);
  return toBase64(result);
}

export async function decryptField(ciphertextB64: string, dek: Uint8Array): Promise<string> {
  if (!ciphertextB64) return '';
  const key = await importDEK(dek);
  const data = fromBase64(ciphertextB64);
  const iv = data.slice(0, 12);
  const ciphertext = data.slice(12);

  const decrypted = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv },
    key,
    ciphertext,
  );

  return new TextDecoder().decode(decrypted);
}
