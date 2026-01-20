'use client';

import React, { useTransition } from 'react';
import { useLocale } from 'next-intl';
import { usePathname, useRouter } from '../i18n/routing';
import { GlobeAltIcon } from '@heroicons/react/24/outline';

export function LanguageSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const [isPending, startTransition] = useTransition();

  const handleLocaleChange = (newLocale: string) => {
    startTransition(() => {
      router.replace(pathname, { locale: newLocale });
    });
  };

  return (
    <div className="relative group">
       <button className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-white/5 transition-colors text-text/60 hover:text-text">
         <GlobeAltIcon className="w-5 h-5" />
         <span className="text-sm font-medium uppercase">{locale}</span>
       </button>
       <div className="absolute right-0 top-full mt-2 w-32 bg-surface border border-white/10 rounded-xl shadow-glass-md opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50 overflow-hidden">
         <button 
           onClick={() => handleLocaleChange('en')}
           className={`w-full text-left px-4 py-2 text-sm hover:bg-white/5 transition-colors ${locale === 'en' ? 'text-primary' : 'text-text'}`}
         >
           English
         </button>
         <button 
           onClick={() => handleLocaleChange('es')}
           className={`w-full text-left px-4 py-2 text-sm hover:bg-white/5 transition-colors ${locale === 'es' ? 'text-primary' : 'text-text'}`}
         >
           Español
         </button>
         <button 
           onClick={() => handleLocaleChange('zh-TW')}
           className={`w-full text-left px-4 py-2 text-sm hover:bg-white/5 transition-colors ${locale === 'zh-TW' ? 'text-primary' : 'text-text'}`}
         >
           繁體中文
         </button>
       </div>
    </div>
  );
}
