import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select } from '@/components/ui/select';
import { useAdminUsers, useAdminUpdateUser, useAdminDeleteUser, useAuditLogs } from '@/api/hooks';
import { useAuthStore } from '@/stores/auth';
import { useToast } from '@/components/ui/toast';
import { formatDate } from '@/lib/utils';
import { Trash2, Users, ShieldCheck, ScrollText, ChevronLeft, ChevronRight } from 'lucide-react';

export function AdminPage() {
  const [tab, setTab] = useState<'users' | 'audit'>('users');

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h1 className="text-3xl font-bold">Admin</h1>
        <p className="text-muted-foreground">Manage users and system settings</p>
      </div>

      <div className="flex gap-1 border-b">
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === 'users' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
          onClick={() => setTab('users')}
        >
          <Users className="inline h-4 w-4 mr-1.5 -mt-0.5" />
          Users
        </button>
        <button
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === 'audit' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
          onClick={() => setTab('audit')}
        >
          <ScrollText className="inline h-4 w-4 mr-1.5 -mt-0.5" />
          Audit Log
        </button>
      </div>

      {tab === 'users' ? <UsersTab /> : <AuditLogTab />}
    </div>
  );
}

function UsersTab() {
  const { data, isLoading } = useAdminUsers();
  const updateUser = useAdminUpdateUser();
  const deleteUser = useAdminDeleteUser();
  const currentUser = useAuthStore((s) => s.user);
  const { toast } = useToast();

  const users = data?.users ?? [];

  function handleRoleChange(userId: string, role: string) {
    updateUser.mutate(
      { id: userId, role },
      {
        onSuccess: () => toast('Role updated successfully', 'success'),
        onError: (err) => toast(err.message, 'error'),
      },
    );
  }

  function handleDelete(userId: string, email: string) {
    if (!confirm(`Delete user ${email}? This cannot be undone.`)) return;
    deleteUser.mutate(userId, {
      onSuccess: () => toast('User deleted', 'success'),
      onError: (err) => toast(err.message, 'error'),
    });
  }

  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{users.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Admins</CardTitle>
            <ShieldCheck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{users.filter((u) => u.role === 'admin').length}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Users</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-16 rounded-lg bg-muted animate-pulse" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {users.map((user) => (
                <div key={user.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div className="flex items-center gap-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary font-medium">
                      {user.email.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <p className="font-medium">{user.email}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant={user.totp_enabled ? 'success' : 'outline'}>
                          {user.totp_enabled ? '2FA On' : '2FA Off'}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          Joined {formatDate(user.created_at)}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <Select
                      value={user.role}
                      onChange={(e) => handleRoleChange(user.id, e.target.value)}
                      className="w-28"
                      disabled={user.id === currentUser?.id}
                    >
                      <option value="admin">Admin</option>
                      <option value="user">User</option>
                    </Select>

                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(user.id, user.email)}
                      disabled={user.id === currentUser?.id}
                      title={user.id === currentUser?.id ? "Can't delete yourself" : 'Delete user'}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </>
  );
}

const ACTION_LABELS: Record<string, string> = {
  login: 'Login',
  logout: 'Logout',
  register: 'Register',
  password_reset_request: 'Password Reset Request',
  password_reset_confirm: 'Password Reset',
  verify_email: 'Email Verified',
  role_change: 'Role Changed',
  user_delete: 'User Deleted',
};

function AuditLogTab() {
  const [page, setPage] = useState(1);
  const [actionFilter, setActionFilter] = useState('');
  const { data, isLoading } = useAuditLogs({ action: actionFilter || undefined, page, limit: 20 });

  const logs = data?.logs ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Audit Log</CardTitle>
        <Select
          value={actionFilter}
          onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
          className="w-48"
        >
          <option value="">All Actions</option>
          {Object.entries(ACTION_LABELS).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </Select>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-12 rounded-lg bg-muted animate-pulse" />
            ))}
          </div>
        ) : logs.length === 0 ? (
          <p className="text-center text-sm text-muted-foreground py-8">No audit logs found</p>
        ) : (
          <div className="space-y-1">
            {logs.map((log) => (
              <div key={log.id} className="flex items-center gap-4 rounded-lg border px-4 py-3 text-sm">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="shrink-0">
                      {ACTION_LABELS[log.action] || log.action}
                    </Badge>
                    {log.resource_type && (
                      <span className="text-xs text-muted-foreground truncate">
                        {log.resource_type}{log.resource_id ? `: ${log.resource_id.slice(0, 8)}...` : ''}
                      </span>
                    )}
                  </div>
                  {log.details && Object.keys(log.details).length > 0 && (
                    <p className="text-xs text-muted-foreground mt-1 truncate">
                      {Object.entries(log.details).map(([k, v]) => `${k}: ${v}`).join(', ')}
                    </p>
                  )}
                </div>
                <div className="text-right shrink-0">
                  <p className="text-xs text-muted-foreground">
                    {new Date(log.created_at).toLocaleString()}
                  </p>
                  {log.ip_address && (
                    <p className="text-[10px] text-muted-foreground">{log.ip_address}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm text-muted-foreground">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
