import {
  Briefcase, Laptop, TrendingUp, Home, Car,
  Utensils, ShoppingCart, Zap, Heart, Film,
  ShoppingBag, BookOpen, Shield, Repeat, Plane,
  Smile, Gift, FileText, PlusCircle, MoreHorizontal,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@/lib/utils';

const ICON_MAP: Record<string, LucideIcon> = {
  'briefcase': Briefcase,
  'laptop': Laptop,
  'trending-up': TrendingUp,
  'home': Home,
  'car': Car,
  'utensils': Utensils,
  'shopping-cart': ShoppingCart,
  'zap': Zap,
  'heart': Heart,
  'film': Film,
  'shopping-bag': ShoppingBag,
  'book-open': BookOpen,
  'shield': Shield,
  'repeat': Repeat,
  'plane': Plane,
  'smile': Smile,
  'gift': Gift,
  'file-text': FileText,
  'plus-circle': PlusCircle,
  'more-horizontal': MoreHorizontal,
};

export const ICON_NAMES = Object.keys(ICON_MAP);

interface CategoryIconProps {
  icon: string;
  color?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizeMap = {
  sm: { container: 'h-6 w-6', icon: 'h-3 w-3' },
  md: { container: 'h-8 w-8', icon: 'h-4 w-4' },
  lg: { container: 'h-10 w-10', icon: 'h-5 w-5' },
};

export function CategoryIcon({ icon, color, size = 'md', className }: CategoryIconProps) {
  const IconComp = ICON_MAP[icon];
  const s = sizeMap[size];

  return (
    <div
      className={cn('flex items-center justify-center rounded-full shrink-0', s.container, className)}
      style={{ backgroundColor: color ? `${color}20` : undefined }}
    >
      {IconComp ? (
        <IconComp className={s.icon} style={{ color: color || 'currentColor' }} />
      ) : (
        <span className="text-xs font-bold" style={{ color: color || 'currentColor' }}>
          {icon.charAt(0).toUpperCase()}
        </span>
      )}
    </div>
  );
}
