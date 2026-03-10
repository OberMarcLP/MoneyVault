import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ConfirmDialog } from '@/components/ui/ConfirmDialog';
import { ListSkeleton } from '@/components/ui/skeleton';
import { CategoryForm } from '@/components/forms/CategoryForm';
import { useCategories, useDeleteCategory } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import { CategoryIcon } from '@/components/ui/CategoryIcon';
import { Plus, Pencil, Trash2 } from 'lucide-react';
import type { Category } from '@/types';

export function CategoriesPage() {
  const { data, isLoading } = useCategories();
  const deleteCategory = useDeleteCategory();
  const { toast } = useToast();
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [deleting, setDeleting] = useState<Category | null>(null);

  const categories = data?.categories ?? [];
  const incomeCategories = categories.filter((c) => c.type === 'income');
  const expenseCategories = categories.filter((c) => c.type === 'expense');

  function handleDeleteConfirm() {
    if (!deleting) return;
    deleteCategory.mutate(deleting.id, {
      onSuccess: () => { toast('Category deleted', 'success'); setDeleting(null); },
      onError: (err) => toast(err.message, 'error'),
    });
  }

  function renderCategory(cat: Category) {
    return (
      <div key={cat.id} className="flex items-center justify-between rounded-lg border p-3">
        <div className="flex items-center gap-3">
          <CategoryIcon icon={cat.icon} color={cat.color} />
          <span className="font-medium">{cat.name}</span>
        </div>
        <div className="flex gap-1">
          <Button variant="ghost" size="icon" onClick={() => setEditing(cat)}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => setDeleting(cat)}>
            <Trash2 className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Categories</h1>
          <p className="text-muted-foreground">Organize your income and expenses</p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Category
        </Button>
      </div>

      {isLoading ? (
        <div className="grid gap-6 lg:grid-cols-2">
          <ListSkeleton rows={4} />
          <ListSkeleton rows={4} />
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                Income
                <Badge variant="success">{incomeCategories.length}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {incomeCategories.length === 0 ? (
                <p className="text-muted-foreground text-sm">No income categories</p>
              ) : (
                incomeCategories.map(renderCategory)
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                Expenses
                <Badge variant="destructive">{expenseCategories.length}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {expenseCategories.length === 0 ? (
                <p className="text-muted-foreground text-sm">No expense categories</p>
              ) : (
                expenseCategories.map(renderCategory)
              )}
            </CardContent>
          </Card>
        </div>
      )}

      <Dialog open={showCreate} onClose={() => setShowCreate(false)}>
        <DialogHeader>
          <DialogTitle>Add Category</DialogTitle>
        </DialogHeader>
        <CategoryForm onSuccess={() => setShowCreate(false)} />
      </Dialog>

      <Dialog open={!!editing} onClose={() => setEditing(null)}>
        <DialogHeader>
          <DialogTitle>Edit Category</DialogTitle>
        </DialogHeader>
        {editing && <CategoryForm category={editing} onSuccess={() => setEditing(null)} />}
      </Dialog>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={handleDeleteConfirm}
        title="Delete Category"
        message={`Delete category "${deleting?.name}"? Transactions using it will become uncategorized.`}
        isPending={deleteCategory.isPending}
      />
    </div>
  );
}
