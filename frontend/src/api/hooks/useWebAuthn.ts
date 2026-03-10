import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiFetch, setAccessToken } from '../client';
import type { User, WebAuthnCredential } from '@/types';
import { useAuthStore } from '@/stores/auth';

export function useWebAuthnCredentials() {
  return useQuery({
    queryKey: ['webauthn-credentials'],
    queryFn: () => apiFetch<{ credentials: WebAuthnCredential[] }>('/auth/webauthn/credentials'),
  });
}

export function useWebAuthnRegisterBegin() {
  return useMutation({
    mutationFn: () => apiFetch<{ options: PublicKeyCredentialCreationOptions }>('/auth/webauthn/register/begin', { method: 'POST' }),
  });
}

export function useWebAuthnRegisterFinish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { credential: unknown; name?: string }) =>
      apiFetch('/auth/webauthn/register/finish?name=' + encodeURIComponent(body.name || 'My Passkey'), {
        method: 'POST',
        body: JSON.stringify(body.credential),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webauthn-credentials'] }),
  });
}

export function useWebAuthnLoginBegin() {
  return useMutation({
    mutationFn: (email: string) =>
      apiFetch<{ options: PublicKeyCredentialRequestOptions }>('/auth/webauthn/login/begin', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
  });
}

export function useWebAuthnLoginFinish() {
  const setUser = useAuthStore((s) => s.setUser);
  return useMutation({
    mutationFn: async ({ email, credential }: { email: string; credential: unknown }) => {
      const data = await apiFetch<{ access_token: string; user: User }>(
        '/auth/webauthn/login/finish?email=' + encodeURIComponent(email),
        { method: 'POST', body: JSON.stringify(credential) },
      );
      setAccessToken(data.access_token);
      setUser(data.user);
      return data;
    },
  });
}

export function useDeleteWebAuthnCredential() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiFetch(`/auth/webauthn/credentials/${id}`, { method: 'DELETE' }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['webauthn-credentials'] }),
  });
}
