import { Button } from '@/components/ui/button';
import { Select } from '@/components/ui/select';
import { FileText, Upload, Loader2 } from 'lucide-react';
import type { CSVImportMapping, CSVPreview as CSVPreviewType, Account } from '@/types';

export type ImportTemplate = {
  name: string;
  mapping: CSVImportMapping;
  postedOnly: boolean;
};

interface ColumnMapping {
  date: string;
  amount: string;
  description: string;
  merchant: string;
  category: string;
  subCategory: string;
  type: string;
  status: string;
  currency: string;
}

interface CSVPreviewProps {
  file: File;
  preview: CSVPreviewType;
  accounts: Account[];
  accountId: string;
  onAccountIdChange: (id: string) => void;
  isStructuredFile: boolean;
  mapping: ColumnMapping;
  onMappingChange: (field: keyof ColumnMapping, value: string) => void;
  postedOnly: boolean;
  onPostedOnlyChange: (val: boolean) => void;
  suggestedTemplate: string;
  selectedTemplate: string;
  savedTemplates: ImportTemplate[];
  templateName: string;
  onTemplateNameChange: (name: string) => void;
  onApplyTemplate: (name: string) => void;
  onSaveTemplate: () => void;
  onResetImport: () => void;
  onImport: () => void;
  isImporting: boolean;
}

export function CSVPreview({
  file,
  preview,
  accounts,
  accountId,
  onAccountIdChange,
  isStructuredFile,
  mapping,
  onMappingChange,
  postedOnly,
  onPostedOnlyChange,
  suggestedTemplate,
  selectedTemplate,
  savedTemplates,
  templateName,
  onTemplateNameChange,
  onApplyTemplate,
  onSaveTemplate,
  onResetImport,
  onImport,
  isImporting,
}: CSVPreviewProps) {
  const mappedColumns = [
    mapping.date,
    mapping.amount,
    mapping.description,
    mapping.merchant,
    mapping.category,
    mapping.subCategory,
    mapping.type,
    mapping.status,
    mapping.currency,
  ];

  function getHeaderSuffix(h: string): string {
    if (h === mapping.date) return ' (Date)';
    if (h === mapping.amount) return ' (Amount)';
    if (h === mapping.description) return ' (Desc)';
    if (h === mapping.merchant) return ' (Merchant)';
    if (h === mapping.category) return ' (Category)';
    if (h === mapping.subCategory) return ' (SubCat)';
    if (h === mapping.type) return ' (Type)';
    if (h === mapping.status) return ' (Status)';
    if (h === mapping.currency) return ' (Currency)';
    return '';
  }

  return (
    <>
      <div className="flex items-center justify-between rounded-lg border p-3">
        <div className="flex items-center gap-3">
          <FileText className="h-5 w-5 text-primary" />
          <div>
            <p className="text-sm font-medium">{file.name}</p>
            <p className="text-xs text-muted-foreground">
              {preview.total} rows &middot; {preview.headers.length} columns detected
            </p>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={onResetImport}>Change file</Button>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Import to account *</label>
        <Select value={accountId} onChange={(e) => onAccountIdChange(e.target.value)} required>
          <option value="">Select account</option>
          {accounts.map((a) => (
            <option key={a.id} value={a.id}>{a.name} ({a.currency})</option>
          ))}
        </Select>
        {accounts.length === 0 && (
          <p className="text-xs text-destructive">No accounts found. Please create an account first.</p>
        )}
      </div>

      {isStructuredFile && (
        <div className="rounded-lg border border-primary/30 bg-primary/5 p-3">
          <p className="text-sm font-medium">
            {file.name.toLowerCase().endsWith('.qif') ? 'QIF' : 'OFX'} file detected
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            Column mapping is handled automatically for this file format. Select an account and click Import.
          </p>
        </div>
      )}

      {!isStructuredFile && (
        <>
          <div className="rounded-lg border p-3 space-y-3">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-sm font-medium">Template</p>
                <p className="text-xs text-muted-foreground">Detected: {suggestedTemplate}</p>
              </div>
              <Select value={selectedTemplate} onChange={(e) => onApplyTemplate(e.target.value)}>
                <option value="">Apply template</option>
                <option value="Swisscards preset">Swisscards preset</option>
                {savedTemplates.map((t) => (
                  <option key={t.name} value={t.name}>{t.name}</option>
                ))}
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <input
                value={templateName}
                onChange={(e) => onTemplateNameChange(e.target.value)}
                placeholder="Template name"
                className="h-9 w-full rounded-md border bg-background px-3 text-sm"
              />
              <Button type="button" variant="outline" onClick={onSaveTemplate}>Save</Button>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Date column *</label>
              <Select value={mapping.date} onChange={(e) => onMappingChange('date', e.target.value)} required>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Amount column *</label>
              <Select value={mapping.amount} onChange={(e) => onMappingChange('amount', e.target.value)} required>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Description column</label>
              <Select value={mapping.description} onChange={(e) => onMappingChange('description', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Merchant column</label>
              <Select value={mapping.merchant} onChange={(e) => onMappingChange('merchant', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Category column</label>
              <Select value={mapping.category} onChange={(e) => onMappingChange('category', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Sub-category column</label>
              <Select value={mapping.subCategory} onChange={(e) => onMappingChange('subCategory', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Debit/Credit column</label>
              <Select value={mapping.type} onChange={(e) => onMappingChange('type', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Status column</label>
              <Select value={mapping.status} onChange={(e) => onMappingChange('status', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Currency column</label>
              <Select value={mapping.currency} onChange={(e) => onMappingChange('currency', e.target.value)}>
                <option value="">None</option>
                {preview.headers.map((h) => (
                  <option key={h} value={h}>{h}</option>
                ))}
              </Select>
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={postedOnly}
              onChange={(e) => onPostedOnlyChange(e.target.checked)}
            />
            Import posted transactions only
          </label>
        </>
      )}

      <div className="rounded-lg border overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              {preview.headers.map((h) => (
                <th key={h} className={`px-3 py-2 text-left font-medium ${
                  mappedColumns.includes(h)
                    ? 'text-primary'
                    : 'text-muted-foreground'
                }`}>
                  {h}
                  {getHeaderSuffix(h)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {preview.rows.slice(0, 5).map((row, i) => (
              <tr key={i} className="border-b">
                {preview.headers.map((h) => (
                  <td key={h} className="px-3 py-2 whitespace-nowrap">{row.values[h] || ''}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        {preview.total > 5 && (
          <p className="px-3 py-2 text-xs text-muted-foreground">
            Showing 5 of {preview.total} rows
          </p>
        )}
      </div>

      <Button onClick={onImport} className="w-full" disabled={isImporting || !accountId}>
        {isImporting ? (
          <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Importing...</>
        ) : (
          <><Upload className="mr-2 h-4 w-4" /> Import {preview.total} transactions</>
        )}
      </Button>
    </>
  );
}
