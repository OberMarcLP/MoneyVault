import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { formatDate } from '@/lib/utils';
import type { ImportJob } from '@/types';

interface ImportHistoryProps {
  history: ImportJob[];
}

export function ImportHistory({ history }: ImportHistoryProps) {
  if (history.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Import History</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {history.map((job) => (
            <div key={job.id} className="flex items-center justify-between rounded-lg border p-3">
              <div className="flex items-center gap-3">
                {job.status === 'completed' ? (
                  <CheckCircle className="h-5 w-5 text-success" />
                ) : job.status === 'failed' ? (
                  <AlertCircle className="h-5 w-5 text-destructive" />
                ) : (
                  <Loader2 className="h-5 w-5 animate-spin" />
                )}
                <div>
                  <p className="text-sm font-medium">{job.filename}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatDate(job.created_at)} &middot;{' '}
                    {job.imported_rows} imported, {job.duplicate_rows} duplicates
                  </p>
                  {job.error_message && (
                    <p className="text-xs text-destructive mt-0.5">{job.error_message}</p>
                  )}
                </div>
              </div>
              <Badge variant={job.status === 'completed' ? 'success' : job.status === 'failed' ? 'destructive' : 'secondary'}>
                {job.status}
              </Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
