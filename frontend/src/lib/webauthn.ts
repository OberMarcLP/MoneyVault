// WebAuthn helper utilities for encoding/decoding between
// ArrayBuffer (browser API) and base64url (server API).

function bufferToBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let str = '';
  for (const b of bytes) str += String.fromCharCode(b);
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function base64urlToBuffer(base64url: string): ArrayBuffer {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4);
  const binary = atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes.buffer;
}

// Recursively decode base64url fields in WebAuthn server options to ArrayBuffers.
function decodeCreationOptions(options: Record<string, unknown>): PublicKeyCredentialCreationOptions {
  const publicKey = options as Record<string, unknown>;

  // Decode challenge
  if (typeof publicKey.challenge === 'string') {
    publicKey.challenge = base64urlToBuffer(publicKey.challenge);
  }

  // Decode user.id
  const user = publicKey.user as Record<string, unknown> | undefined;
  if (user && typeof user.id === 'string') {
    user.id = base64urlToBuffer(user.id);
  }

  // Decode excludeCredentials[].id
  const exclude = publicKey.excludeCredentials as Array<Record<string, unknown>> | undefined;
  if (exclude) {
    for (const cred of exclude) {
      if (typeof cred.id === 'string') cred.id = base64urlToBuffer(cred.id);
    }
  }

  return { publicKey } as unknown as PublicKeyCredentialCreationOptions;
}

function decodeRequestOptions(options: Record<string, unknown>): PublicKeyCredentialRequestOptions {
  const publicKey = options as Record<string, unknown>;

  if (typeof publicKey.challenge === 'string') {
    publicKey.challenge = base64urlToBuffer(publicKey.challenge);
  }

  const allow = publicKey.allowCredentials as Array<Record<string, unknown>> | undefined;
  if (allow) {
    for (const cred of allow) {
      if (typeof cred.id === 'string') cred.id = base64urlToBuffer(cred.id);
    }
  }

  return { publicKey } as unknown as PublicKeyCredentialRequestOptions;
}

function encodeRegistrationCredential(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAttestationResponse;
  return {
    id: credential.id,
    rawId: bufferToBase64url(credential.rawId),
    type: credential.type,
    response: {
      attestationObject: bufferToBase64url(response.attestationObject),
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
    },
  };
}

function encodeAuthenticationCredential(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAssertionResponse;
  return {
    id: credential.id,
    rawId: bufferToBase64url(credential.rawId),
    type: credential.type,
    response: {
      authenticatorData: bufferToBase64url(response.authenticatorData),
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
      signature: bufferToBase64url(response.signature),
      userHandle: response.userHandle ? bufferToBase64url(response.userHandle) : null,
    },
  };
}

export async function registerPasskey(
  serverOptions: Record<string, unknown>,
): Promise<unknown> {
  // The server sends { publicKey: { ... } } inside options
  const publicKeyOptions = (serverOptions as { publicKey?: Record<string, unknown> }).publicKey ?? serverOptions;
  const createOptions = decodeCreationOptions(publicKeyOptions);
  const credential = await navigator.credentials.create(createOptions as CredentialCreationOptions) as PublicKeyCredential;
  if (!credential) throw new Error('Passkey registration was cancelled');
  return encodeRegistrationCredential(credential);
}

export async function authenticatePasskey(
  serverOptions: Record<string, unknown>,
): Promise<unknown> {
  const publicKeyOptions = (serverOptions as { publicKey?: Record<string, unknown> }).publicKey ?? serverOptions;
  const requestOptions = decodeRequestOptions(publicKeyOptions);
  const credential = await navigator.credentials.get(requestOptions as CredentialRequestOptions) as PublicKeyCredential;
  if (!credential) throw new Error('Passkey authentication was cancelled');
  return encodeAuthenticationCredential(credential);
}

export function isWebAuthnSupported(): boolean {
  return !!window.PublicKeyCredential;
}
