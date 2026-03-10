import { useRef } from 'react';
import { Upload, Loader2, AlertCircle } from 'lucide-react';

interface CSVUploaderProps {
  onFileSelect: (file: File) => void;
  isPending: boolean;
  previewError: string;
  dragOver: boolean;
  onDragOver: () => void;
  onDragLeave: () => void;
}

export function CSVUploader({
  onFileSelect,
  isPending,
  previewError,
  dragOver,
  onDragOver,
  onDragLeave,
}: CSVUploaderProps) {
  const fileRef = useRef<HTMLInputElement>(null);

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0];
    if (!f) return;
    onFileSelect(f);
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    onDragLeave();
    const f = e.dataTransfer.files?.[0];
    if (f) onFileSelect(f);
  }

  return (
    <>
      <label
        className={`flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 cursor-pointer transition-colors ${
          dragOver ? 'border-primary bg-primary/5' : 'hover:bg-accent/50'
        } ${previewError ? 'border-destructive' : ''}`}
        onDragOver={(e) => { e.preventDefault(); onDragOver(); }}
        onDragLeave={onDragLeave}
        onDrop={handleDrop}
      >
        {isPending ? (
          <>
            <Loader2 className="h-10 w-10 text-primary mb-3 animate-spin" />
            <p className="text-sm font-medium">Analyzing file...</p>
          </>
        ) : (
          <>
            <Upload className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="text-sm font-medium">
              {dragOver ? 'Drop your file here' : 'Drop file here or click to browse'}
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              Supports CSV, OFX, QFX, and QIF files up to 10MB
            </p>
          </>
        )}
        <input
          ref={fileRef}
          type="file"
          accept=".csv,.tsv,.txt,.ofx,.qfx,.qif"
          className="hidden"
          onChange={handleFileChange}
        />
      </label>
      {previewError && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/5 p-3">
          <div className="flex items-start gap-2">
            <AlertCircle className="h-4 w-4 text-destructive mt-0.5 shrink-0" />
            <div>
              <p className="text-sm font-medium text-destructive">Preview failed</p>
              <p className="text-xs text-muted-foreground mt-1">{previewError}</p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
