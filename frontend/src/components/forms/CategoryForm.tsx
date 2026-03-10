import { useState, type FormEvent } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { CategoryIcon, ICON_NAMES } from '@/components/ui/CategoryIcon';
import { useCreateCategory, useUpdateCategory } from '@/api/hooks';
import { useToast } from '@/components/ui/toast';
import type { Category, CategoryType } from '@/types';

interface CategoryFormProps {
  category?: Category;
  onSuccess: () => void;
}

const COLORS = [
  '#3B82F6', '#6366F1', '#8B5CF6', '#A855F7', '#D946EF',
  '#EC4899', '#F43F5E', '#EF4444', '#F59E0B', '#10B981',
  '#14B8A6', '#06B6D4', '#0EA5E9', '#FB923C', '#78716C',
];

export function CategoryForm({ category, onSuccess }: CategoryFormProps) {
  const [name, setName] = useState(category?.name ?? '');
  const [type, setType] = useState<CategoryType>(category?.type ?? 'expense');
  const [icon, setIcon] = useState(category?.icon ?? ICON_NAMES[0]);
  const [color, setColor] = useState(category?.color ?? COLORS[0]);

  const createCat = useCreateCategory();
  const updateCat = useUpdateCategory();
  const { toast } = useToast();
  const isEditing = !!category;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const opts = {
      onSuccess: () => { toast(isEditing ? 'Category updated' : 'Category created', 'success'); onSuccess(); },
      onError: (err: Error) => toast(err.message, 'error'),
    };
    if (isEditing) {
      updateCat.mutate({ id: category.id, name, type, icon, color }, opts);
    } else {
      createCat.mutate({ name, type, icon, color }, opts);
    }
  }

  const isPending = createCat.isPending || updateCat.isPending;
  const error = createCat.error || updateCat.error;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="flex items-center gap-3">
        <CategoryIcon icon={icon} color={color} size="lg" />
        <div className="flex-1 space-y-2">
          <label className="text-sm font-medium">Name</label>
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Category name" required />
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Type</label>
        <Select value={type} onChange={(e) => setType(e.target.value as CategoryType)}>
          <option value="expense">Expense</option>
          <option value="income">Income</option>
        </Select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Icon</label>
        <div className="grid gap-1.5" style={{ gridTemplateColumns: 'repeat(10, minmax(0, 1fr))' }}>
          {ICON_NAMES.map((i) => (
            <button
              key={i}
              type="button"
              onClick={() => setIcon(i)}
              className={`flex items-center justify-center rounded-md p-1.5 transition-colors ${
                icon === i ? 'bg-primary/10 ring-2 ring-primary' : 'hover:bg-accent'
              }`}
              title={i}
            >
              <CategoryIcon icon={i} color={icon === i ? color : undefined} size="sm" />
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Color</label>
        <div className="flex flex-wrap gap-2">
          {COLORS.map((c) => (
            <button
              key={c}
              type="button"
              onClick={() => setColor(c)}
              className={`h-8 w-8 rounded-full border-2 transition-transform ${
                color === c ? 'border-foreground scale-110' : 'border-transparent'
              }`}
              style={{ backgroundColor: c }}
            />
          ))}
        </div>
      </div>

      {error && <p className="text-sm text-destructive">{error.message}</p>}

      <Button type="submit" className="w-full" loading={isPending}>
        {isEditing ? 'Update Category' : 'Create Category'}
      </Button>
    </form>
  );
}
