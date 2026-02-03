'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { LanguageSwitcher } from './LanguageSwitcher';
import { 
  MagnifyingGlassIcon, 
  BellIcon,
  ChevronDownIcon,
  Bars3Icon
} from '@heroicons/react/24/outline';

interface TopBarProps {
  isSidebarCollapsed: boolean;
  isMobile: boolean;
  onMenuClick: () => void;
}

export function TopBar({ isSidebarCollapsed, isMobile, onMenuClick }: TopBarProps) {
  const t = useTranslations('TopBar');

  return (
    <header 
      className={`
        fixed top-0 right-0 h-20 glass-panel border-b border-white/10 z-40 px-4 sm:px-8 flex items-center justify-between transition-all duration-300
        ${isMobile ? 'left-0' : (isSidebarCollapsed ? 'left-20' : 'left-64')}
      `}
    >
      <div className="flex items-center gap-4 flex-1 max-w-xl">
        {/* Mobile Menu Button */}
        {isMobile && (
          <button 
            onClick={onMenuClick}
            className="p-2 -ml-2 rounded-lg hover:bg-white/5 text-text/60 hover:text-text"
          >
            <Bars3Icon className="w-6 h-6" />
          </button>
        )}

        {/* Search */}
        <div className="flex-1">
          <div className="relative group">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text/40 group-focus-within:text-primary transition-colors" />
            <input 
              type="text" 
              placeholder={t('searchPlaceholder')}
              className="w-full bg-white/5 border border-white/10 rounded-xl py-2.5 pl-10 pr-4 text-sm text-text placeholder:text-text/30 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all font-sans"
            />
          </div>
        </div>
      </div>

      {/* Right Actions */}
      <div className="flex items-center gap-3 sm:gap-6 pl-2">
        <LanguageSwitcher />

        {/* Notifications */}
        <button className="relative w-10 h-10 rounded-full flex items-center justify-center hover:bg-white/5 text-text/60 hover:text-text transition-colors">
          <BellIcon className="w-6 h-6" />
          <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-cta shadow-[0_0_8px_#8B5CF6]"></span>
        </button>

        {/* User Profile */}
        <div className="h-8 w-px bg-white/10 hidden sm:block"></div>
        
        <button className="flex items-center gap-3 pl-2 pr-1 rounded-lg hover:bg-white/5 transition-colors group">
          <div className="text-right hidden sm:block">
            <p className="text-sm font-medium text-text group-hover:text-primary transition-colors">Alex Morgan</p>
            <p className="text-xs text-text/40">{t('userRole')}</p>
          </div>
          <div className="w-8 h-8 sm:w-10 sm:h-10 rounded-full bg-gradient-to-tr from-primary to-cta p-[2px]">
            <div className="w-full h-full rounded-full bg-surface border-2 border-surface flex items-center justify-center overflow-hidden">
               <span className="font-bold text-xs text-text">AM</span>
            </div>
          </div>
          <ChevronDownIcon className="w-4 h-4 text-text/40 group-hover:text-text transition-colors hidden sm:block" />
        </button>
      </div>
    </header>
  );
}
