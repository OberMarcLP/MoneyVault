import { describe, it, expect, beforeAll } from 'vitest';
import {
  generateSalt,
  generateDEK,
  encryptDEK,
  decryptDEK,
  encryptField,
  decryptField,
} from './crypto';

describe('crypto utilities', () => {
  describe('generateSalt', () => {
    it('returns a non-empty base64 string', () => {
      const salt = generateSalt();
      expect(salt).toBeTruthy();
      expect(typeof salt).toBe('string');
      // Should be valid base64
      expect(() => atob(salt)).not.toThrow();
    });

    it('generates unique salts', () => {
      const salt1 = generateSalt();
      const salt2 = generateSalt();
      expect(salt1).not.toBe(salt2);
    });
  });

  describe('generateDEK', () => {
    it('returns a 32-byte Uint8Array', async () => {
      const dek = await generateDEK();
      expect(dek).toBeInstanceOf(Uint8Array);
      expect(dek.length).toBe(32);
    });

    it('generates unique DEKs', async () => {
      const dek1 = await generateDEK();
      const dek2 = await generateDEK();
      expect(dek1).not.toEqual(dek2);
    });
  });

  describe('encryptDEK / decryptDEK', () => {
    it('round-trips DEK encryption with password', async () => {
      const dek = await generateDEK();
      const salt = generateSalt();
      const password = 'TestPassword123!';

      const encrypted = await encryptDEK(dek, password, salt);
      expect(encrypted).toBeTruthy();
      expect(typeof encrypted).toBe('string');

      const decrypted = await decryptDEK(encrypted, password, salt);
      expect(decrypted).toEqual(dek);
    });

    it('fails with wrong password', async () => {
      const dek = await generateDEK();
      const salt = generateSalt();

      const encrypted = await encryptDEK(dek, 'correct-password', salt);

      await expect(
        decryptDEK(encrypted, 'wrong-password', salt),
      ).rejects.toThrow();
    });

    it('fails with wrong salt', async () => {
      const dek = await generateDEK();
      const salt1 = generateSalt();
      const salt2 = generateSalt();

      const encrypted = await encryptDEK(dek, 'password', salt1);

      await expect(
        decryptDEK(encrypted, 'password', salt2),
      ).rejects.toThrow();
    });
  });

  describe('encryptField / decryptField', () => {
    let dek: Uint8Array;

    beforeAll(async () => {
      dek = await generateDEK();
    });

    it('round-trips field encryption', async () => {
      const plaintext = 'Hello World';
      const encrypted = await encryptField(plaintext, dek);
      expect(encrypted).not.toBe(plaintext);
      expect(encrypted).toBeTruthy();

      const decrypted = await decryptField(encrypted, dek);
      expect(decrypted).toBe(plaintext);
    });

    it('handles empty strings', async () => {
      expect(await encryptField('', dek)).toBe('');
      expect(await decryptField('', dek)).toBe('');
    });

    it('handles numbers as strings', async () => {
      const encrypted = await encryptField('12345.67', dek);
      const decrypted = await decryptField(encrypted, dek);
      expect(decrypted).toBe('12345.67');
    });

    it('handles unicode text', async () => {
      const text = 'Привет 日本語 🎉';
      const encrypted = await encryptField(text, dek);
      const decrypted = await decryptField(encrypted, dek);
      expect(decrypted).toBe(text);
    });

    it('produces different ciphertexts for same plaintext', async () => {
      const plaintext = 'same text';
      const enc1 = await encryptField(plaintext, dek);
      const enc2 = await encryptField(plaintext, dek);
      expect(enc1).not.toBe(enc2); // Random IV
    });

    it('fails to decrypt with wrong DEK', async () => {
      const wrongDEK = await generateDEK();
      const encrypted = await encryptField('secret', dek);

      await expect(
        decryptField(encrypted, wrongDEK),
      ).rejects.toThrow();
    });
  });
});
