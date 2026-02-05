'use client';

import React, { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { DashboardLayout } from '@/components/DashboardLayout';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Currency } from '@/domain/models/Currency';
import { User } from '@/domain/models/User';
import { Category } from '@/domain/models/Category';
import { CheckIcon } from '@heroicons/react/24/outline';

import { useSearchParams } from 'next/navigation';

export default function SettingsPage() {
  const t = useTranslations('Settings');
  const searchParams = useSearchParams();
  const token = searchParams.get('token') || searchParams.get('user_id') || 'test-user';

  const [currencies, setCurrencies] = useState<Currency[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [selectedCurrency, setSelectedCurrency] = useState('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoriesLoading, setCategoriesLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const currencyRepo = RepositoryFactory.getCurrencyRepository();
        const userRepo = RepositoryFactory.getUserRepository();

        // Token is now derived from searchParams
        
        const [currencyData, userData] = await Promise.all([
          currencyRepo.getCurrencies(),
          userRepo.getUser(token)
        ]);

        setCurrencies(currencyData);
        setUser(userData);
        setSelectedCurrency(userData.home_currency);

        // Fetch categories
        const categoryRepo = RepositoryFactory.getCategoryRepository();
        const categoryData = await categoryRepo.list(token);
        setCategories(categoryData);
      } catch (error) {
        console.error('Failed to fetch settings data', error);
      } finally {
        setLoading(false);
        setCategoriesLoading(false);
      }
    };

    fetchData();
  }, [token]);

  const handleSave = async () => {
    if (!user) return;
    setSaving(true);
    try {
      const userRepo = RepositoryFactory.getUserRepository();

      // Update local state first for responsiveness
      const updatedSettings = {
         home_currency: selectedCurrency,
         locale: user.locale || 'zh-TW'
      };

      await userRepo.updateSettings(token, updatedSettings);
      
      // Update local user object
      setUser({ ...user, home_currency: selectedCurrency });
      
    } catch (error) {
       console.error('Failed to save settings', error);
       // Revert on failure if needed, or show toast
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center p-12">
           <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout>
      <div className="p-4 sm:p-8 max-w-4xl mx-auto space-y-8">
        <div>
          <h1 className="text-3xl font-mono font-bold text-text tracking-tight mb-1">
            Settings
          </h1>
          <p className="text-text/60">Manage your preferences and configurations.</p>
        </div>

        {/* Currency Settings */}
        <div className="glass-panel p-6 rounded-2xl border border-white/5 space-y-6">
          <div>
            <h2 className="text-xl font-bold text-text mb-2">Currency Preferences</h2>
            <p className="text-sm text-text/60">
              Select your home currency. All expenses will be converted to this currency for aggregation.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center">
             <div className="relative w-full sm:w-64">
               <select
                 value={selectedCurrency}
                 onChange={(e) => setSelectedCurrency(e.target.value)}
                 className="w-full appearance-none bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-text focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all cursor-pointer hover:bg-black/30"
               >
                 {currencies.map((c) => (
                   <option key={c.code} value={c.code} className="bg-surface text-text">
                     {c.code} - {c.name} ({c.symbol})
                   </option>
                 ))}
               </select>
               <div className="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-text/40">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-4 h-4">
                   <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                 </svg>
               </div>
             </div>

             <button
               onClick={handleSave}
               disabled={saving || selectedCurrency === user?.home_currency}
               className={`
                 flex items-center gap-2 px-6 py-3 rounded-xl font-medium transition-all
                 ${selectedCurrency === user?.home_currency 
                    ? 'bg-white/5 text-text/30 cursor-not-allowed' 
                    : 'bg-primary text-white hover:bg-primary/90 shadow-lg shadow-primary/20'}
               `}
             >
               {saving ? 'Saving...' : (
                 <>
                   <CheckIcon className="w-5 h-5" />
                   Save Changes
                 </>
               )}
             </button>
          </div>
         </div>

         {/* Category Management */}
         <div className="glass-panel p-6 rounded-2xl border border-white/5 space-y-6">
           <div>
             <h2 className="text-xl font-bold text-text mb-2">Category Management</h2>
             <p className="text-sm text-text/60">
               Manage your expense categories. Default categories cannot be edited or deleted.
             </p>
           </div>

           {categoriesLoading ? (
             <div className="flex items-center justify-center p-8">
               <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
             </div>
           ) : categories.length === 0 ? (
             <div className="text-center py-8 text-text/60">
               No categories found.
             </div>
           ) : (
             <div className="space-y-3">
               {categories.map((category) => (
                 <div 
                   key={category.id}
                   className={`flex items-center justify-between p-4 rounded-xl border ${
                     category.is_default 
                       ? 'bg-white/5 border-white/5' 
                       : 'bg-black/20 border-white/10'
                   }`}
                 >
                   <div className="flex-1 min-w-0">
                     <div className="flex items-center gap-2">
                       <span className={`font-medium ${category.is_default ? 'text-text/60' : 'text-text'}`}>
                         {category.name}
                       </span>
                       {category.is_default && (
                         <span className="text-xs px-2 py-0.5 rounded-full bg-white/10 text-text/50">
                           Default
                         </span>
                       )}
                     </div>
                     {category.description && (
                       <p className="text-sm text-text/50 mt-1 truncate">
                         {category.description}
                       </p>
                     )}
                   </div>
                 </div>
               ))}
             </div>
           )}
         </div>
       </div>
     </DashboardLayout>
   );
 }
