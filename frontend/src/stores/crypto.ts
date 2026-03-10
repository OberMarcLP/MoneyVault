import { create } from 'zustand';

const SESSION_KEY = 'moneyvault-e2e-dek';

function toBase64(buf: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < buf.length; i++) {
    binary += String.fromCharCode(buf[i]);
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

function persistDEK(dek: Uint8Array | null) {
  try {
    if (dek) {
      sessionStorage.setItem(SESSION_KEY, toBase64(dek));
    } else {
      sessionStorage.removeItem(SESSION_KEY);
    }
  } catch {
    // sessionStorage not available
  }
}

function restoreDEK(): Uint8Array | null {
  try {
    const stored = sessionStorage.getItem(SESSION_KEY);
    if (stored) return fromBase64(stored);
  } catch {
    // sessionStorage not available
  }
  return null;
}

interface CryptoState {
  dek: Uint8Array | null;
  isE2EEnabled: boolean;
  setDEK: (dek: Uint8Array) => void;
  clearDEK: () => void;
  setE2EEnabled: (enabled: boolean) => void;
}

const restoredDEK = restoreDEK();

export const useCryptoStore = create<CryptoState>((set) => ({
  dek: restoredDEK,
  isE2EEnabled: restoredDEK !== null,
  setDEK: (dek) => {
    persistDEK(dek);
    set({ dek, isE2EEnabled: true });
  },
  clearDEK: () => {
    persistDEK(null);
    set({ dek: null, isE2EEnabled: false });
  },
  setE2EEnabled: (enabled) => set({ isE2EEnabled: enabled }),
}));
