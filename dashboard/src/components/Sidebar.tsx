'use client';

import React from 'react';
import { usePathname } from 'next/navigation';
import { Link } from '../i18n/routing';
import { useTranslations } from 'next-intl';
import {
  HomeIcon,
  ChartBarIcon,
  Cog6ToothIcon,
  UserCircleIcon
} from '@heroicons/react/24/outline';

const NAVIGATION = [
  { name: 'dashboard', href: '/', icon: HomeIcon },
  { name: 'reports', href: '/reports', icon: ChartBarIcon },
  { name: 'my_expenses', href: '/user/reports', icon: UserCircleIcon },
];

export function Sidebar() {
  const pathname = usePathname();
  const t = useTranslations('Sidebar');
  const isUserPage = pathname?.startsWith('/user');

  // Filter navigation items based on current path
  const filteredNavigation = NAVIGATION.filter(item => {
    if (isUserPage) {
      // In user pages, show only 'my_expenses'
      return item.href.startsWith('/user');
    }
    // In admin pages, show everything (or could exclude user links if desired)
    return true;
  });

  return (
    <aside className="fixed left-0 top-0 h-full w-64 glass-panel border-r border-white/10 z-50 flex flex-col transition-all duration-300">
      {/* Logo */}
      <div className="h-20 flex items-center px-8 border-b border-white/5">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-secondary flex items-center justify-center text-white font-bold font-mono">
            AI
          </div>
          <span className="font-mono font-bold text-lg text-text tracking-tight">Expense</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-8 space-y-2">
        {filteredNavigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link 
              key={item.name} 
              href={item.href}
              className={`
                flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group
                ${isActive 
                  ? 'bg-primary/20 text-primary shadow-[0_0_15px_rgba(245,158,11,0.2)]' 
                  : 'text-text/60 hover:text-text hover:bg-white/5'}
              `}
            >
              <item.icon className={`w-5 h-5 ${isActive ? 'text-primary' : 'group-hover:text-primary transition-colors'}`} />
              <span className="font-medium text-sm">{t(item.name)}</span>
            </Link>
          );
        })}
      </nav>

      {/* Bottom Actions */}
      <div className="p-4 border-t border-white/5">
        <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-text/60 hover:text-text hover:bg-white/5 transition-all duration-200">
          <Cog6ToothIcon className="w-5 h-5" />
          <span className="font-medium text-sm">{t('settings')}</span>
        </button>
      </div>
    </aside>
  );
}
