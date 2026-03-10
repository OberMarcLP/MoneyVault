import { useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { useCSVPreview, useCSVImport, useImportHistory, useAccounts } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import type { CSVImportMapping, CSVPreview as CSVPreviewType } from '@/types';
import { CSVUploader } from './CSVUploader';
import { CSVPreview, type ImportTemplate } from './CSVPreview';
import { ImportHistory } from './ImportHistory';

const IMPORT_TEMPLATES_KEY = 'moneyvault-import-templates-v1';

function isSwisscardsHeaders(headers: string[]): boolean {
  const set = new Set(headers.map((h) => h.toLowerCase().trim()));
  return set.has('transaction date')
    && set.has('debit/credit')
    && set.has('merchant category');
}

function findHeader(headers: string[], candidates: RegExp): string {
  return headers.find((h) => candidates.test(h)) || '';
}

function buildAutoMapping(headers: string[]): { mapping: CSVImportMapping; postedOnly: boolean; suggestedTemplate: string } {
  if (isSwisscardsHeaders(headers)) {
    return {
      mapping: {
        date: 'Transaction date',
        amount: 'Amount',
        description: 'Description',
        merchant: 'Merchant',
        category: 'Merchant Category',
        sub_category: 'Registered Category',
        type: 'Debit/Credit',
        status: 'Status',
        currency: 'Currency',
      },
      postedOnly: true,
      suggestedTemplate: 'Swisscards',
    };
  }

  return {
    mapping: {
      date: findHeader(headers, /date|datum|zeit/i) || headers[0] || '',
      amount: findHeader(headers, /amount|sum|total|betrag|saldo/i) || headers[1] || '',
      description: findHeader(headers, /desc|memo|note|narr|text|verwendung|buchung/i) || headers[2] || '',
      merchant: findHeader(headers, /merchant|vendor|shop|payee|h[äa]ndler/i),
      category: findHeader(headers, /category|kategorie|merchant category/i),
      sub_category: findHeader(headers, /sub.?category|registered category|mcc/i),
      type: findHeader(headers, /type|debit.?credit|art|direction/i),
      status: findHeader(headers, /status|state/i),
      currency: findHeader(headers, /currency|w[äa]hrung/i),
    },
    postedOnly: false,
    suggestedTemplate: 'General CSV',
  };
}

export function ImportPage() {
  const { data: accountsData } = useAccounts();
  const { data: historyData, refetch: refetchHistory } = useImportHistory();
  const previewMutation = useCSVPreview();
  const importMutation = useCSVImport();
  const { toast } = useToast();

  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<CSVPreviewType | null>(null);
  const [accountId, setAccountId] = useState('');
  const [mapDate, setMapDate] = useState('');
  const [mapAmount, setMapAmount] = useState('');
  const [mapDescription, setMapDescription] = useState('');
  const [mapMerchant, setMapMerchant] = useState('');
  const [mapCategory, setMapCategory] = useState('');
  const [mapSubCategory, setMapSubCategory] = useState('');
  const [mapType, setMapType] = useState('');
  const [mapStatus, setMapStatus] = useState('');
  const [mapCurrency, setMapCurrency] = useState('');
  const [postedOnly, setPostedOnly] = useState(false);
  const [templateName, setTemplateName] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState('');
  const [suggestedTemplate, setSuggestedTemplate] = useState('General CSV');
  const [dragOver, setDragOver] = useState(false);
  const [previewError, setPreviewError] = useState('');

  const accounts = accountsData?.accounts ?? [];
  const history = historyData?.imports ?? [];
  const [savedTemplates, setSavedTemplates] = useState<ImportTemplate[]>(() => {
    try {
      const raw = localStorage.getItem(IMPORT_TEMPLATES_KEY);
      if (!raw) return [];
      const parsed = JSON.parse(raw) as ImportTemplate[];
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  });

  const applyMapping = useCallback((mapping: CSVImportMapping, posted: boolean) => {
    setMapDate(mapping.date || '');
    setMapAmount(mapping.amount || '');
    setMapDescription(mapping.description || '');
    setMapMerchant(mapping.merchant || '');
    setMapCategory(mapping.category || '');
    setMapSubCategory(mapping.sub_category || '');
    setMapType(mapping.type || '');
    setMapStatus(mapping.status || '');
    setMapCurrency(mapping.currency || '');
    setPostedOnly(posted);
  }, []);

  const processFile = useCallback((f: File) => {
    const name = f.name.toLowerCase();
    const validExts = ['.csv', '.tsv', '.txt', '.ofx', '.qfx', '.qif'];
    if (!validExts.some((ext) => name.endsWith(ext)) && f.type !== 'text/csv') {
      toast('Please select a CSV, OFX, or QIF file', 'error');
      return;
    }
    if (f.size > 10 * 1024 * 1024) {
      toast('File is too large (max 10MB)', 'error');
      return;
    }
    setFile(f);
    setPreviewError('');
    previewMutation.mutate(f, {
      onSuccess: (data) => {
        setPreview(data.preview);
        setPreviewError('');
        if (data.preview.headers.length > 0) {
          const auto = buildAutoMapping(data.preview.headers);
          applyMapping(auto.mapping, auto.postedOnly);
          setSuggestedTemplate(auto.suggestedTemplate);
        }
      },
      onError: (err) => {
        setPreviewError(err.message);
        toast(err.message, 'error');
      },
    });
  }, [previewMutation, toast, applyMapping]);

  function handleMappingChange(field: string, value: string) {
    const setters: Record<string, (v: string) => void> = {
      date: setMapDate,
      amount: setMapAmount,
      description: setMapDescription,
      merchant: setMapMerchant,
      category: setMapCategory,
      subCategory: setMapSubCategory,
      type: setMapType,
      status: setMapStatus,
      currency: setMapCurrency,
    };
    setters[field]?.(value);
  }

  function handleSaveTemplate() {
    if (!templateName.trim()) {
      toast('Enter a template name first', 'error');
      return;
    }
    const next: ImportTemplate = {
      name: templateName.trim(),
      mapping: {
        date: mapDate,
        amount: mapAmount,
        description: mapDescription,
        merchant: mapMerchant,
        category: mapCategory,
        sub_category: mapSubCategory,
        type: mapType,
        status: mapStatus,
        currency: mapCurrency,
      },
      postedOnly,
    };
    const deduped = savedTemplates.filter((t) => t.name.toLowerCase() !== next.name.toLowerCase());
    const updated = [next, ...deduped].slice(0, 20);
    localStorage.setItem(IMPORT_TEMPLATES_KEY, JSON.stringify(updated));
    setSavedTemplates(updated);
    setTemplateName('');
    toast(`Saved template "${next.name}"`, 'success');
  }

  function handleApplyTemplate(name: string) {
    setSelectedTemplate(name);
    if (name === 'Swisscards preset') {
      applyMapping({
        date: 'Transaction date',
        amount: 'Amount',
        description: 'Description',
        merchant: 'Merchant',
        category: 'Merchant Category',
        sub_category: 'Registered Category',
        type: 'Debit/Credit',
        status: 'Status',
        currency: 'Currency',
      }, true);
      return;
    }
    const tpl = savedTemplates.find((t) => t.name === name);
    if (tpl) applyMapping(tpl.mapping, tpl.postedOnly);
  }

  const isStructuredFile = file ? /\.(ofx|qfx|qif)$/i.test(file.name) : false;

  function handleImport() {
    if (!file || !accountId) {
      toast('Please select a file and account', 'error');
      return;
    }
    if (!isStructuredFile && (!mapDate || !mapAmount)) {
      toast('Please map date and amount columns', 'error');
      return;
    }
    importMutation.mutate(
      {
        file,
        accountId,
        mapping: isStructuredFile
          ? { date: 'Date', amount: 'Amount', description: 'Description', merchant: '', category: '', sub_category: '', type: 'Type', status: '', currency: '' }
          : {
              date: mapDate,
              amount: mapAmount,
              description: mapDescription,
              merchant: mapMerchant,
              category: mapCategory,
              sub_category: mapSubCategory,
              type: mapType,
              status: mapStatus,
              currency: mapCurrency,
            },
        postedOnly: isStructuredFile ? false : postedOnly,
      },
      {
        onSuccess: (data) => {
          const job = data.import;
          if (job.status === 'failed') {
            toast(job.error_message || 'Import failed — check your CSV format', 'error');
          } else {
            toast(`Imported ${job.imported_rows} transactions (${job.duplicate_rows} duplicates skipped)`, 'success');
          }
          setFile(null);
          setPreview(null);
          setPreviewError('');
          refetchHistory();
        },
        onError: (err) => toast(err.message, 'error'),
      },
    );
  }

  function resetImport() {
    setFile(null);
    setPreview(null);
    setPreviewError('');
    setMapDate('');
    setMapAmount('');
    setMapDescription('');
    setMapMerchant('');
    setMapCategory('');
    setMapSubCategory('');
    setMapType('');
    setMapStatus('');
    setMapCurrency('');
    setPostedOnly(false);
    previewMutation.reset();
  }

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h1 className="text-3xl font-bold">Import Transactions</h1>
        <p className="text-muted-foreground">Upload a CSV file to import transactions</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Upload CSV</CardTitle>
          <CardDescription>Select a CSV file from your bank or financial institution</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!preview ? (
            <CSVUploader
              onFileSelect={processFile}
              isPending={previewMutation.isPending}
              previewError={previewError}
              dragOver={dragOver}
              onDragOver={() => setDragOver(true)}
              onDragLeave={() => setDragOver(false)}
            />
          ) : file && (
            <CSVPreview
              file={file}
              preview={preview}
              accounts={accounts}
              accountId={accountId}
              onAccountIdChange={setAccountId}
              isStructuredFile={isStructuredFile}
              mapping={{
                date: mapDate,
                amount: mapAmount,
                description: mapDescription,
                merchant: mapMerchant,
                category: mapCategory,
                subCategory: mapSubCategory,
                type: mapType,
                status: mapStatus,
                currency: mapCurrency,
              }}
              onMappingChange={handleMappingChange}
              postedOnly={postedOnly}
              onPostedOnlyChange={setPostedOnly}
              suggestedTemplate={suggestedTemplate}
              selectedTemplate={selectedTemplate}
              savedTemplates={savedTemplates}
              templateName={templateName}
              onTemplateNameChange={setTemplateName}
              onApplyTemplate={handleApplyTemplate}
              onSaveTemplate={handleSaveTemplate}
              onResetImport={resetImport}
              onImport={handleImport}
              isImporting={importMutation.isPending}
            />
          )}
        </CardContent>
      </Card>

      <ImportHistory history={history} />
    </div>
  );
}
