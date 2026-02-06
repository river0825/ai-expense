'use client';

import React, { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { DashboardLayout } from '@/components/DashboardLayout';
import RepositoryFactory from '@/infrastructure/RepositoryFactory';
import { Currency } from '@/domain/models/Currency';
import { User } from '@/domain/models/User';
import { Category } from '@/domain/models/Category';
import { CheckIcon, PencilSquareIcon, TrashIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { MergeResult } from '@/domain/models/Category';

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
  const [isAddingCategory, setIsAddingCategory] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newCategoryDescription, setNewCategoryDescription] = useState('');
  const [addingCategoryError, setAddingCategoryError] = useState('');
  const [savingCategory, setSavingCategory] = useState(false);
  const [deletingCategoryId, setDeletingCategoryId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Edit category state
  const [editingCategoryId, setEditingCategoryId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editError, setEditError] = useState('');
  const [savingEdit, setSavingEdit] = useState(false);

  // Merge confirmation state
  const [showMergeConfirm, setShowMergeConfirm] = useState(false);
  const [mergeTargetId, setMergeTargetId] = useState<string | null>(null);
  const [mergeResult, setMergeResult] = useState<MergeResult | null>(null);
  const [mergingCategory, setMergingCategory] = useState(false);

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

   const handleAddCategory = async () => {
     if (!newCategoryName.trim()) {
       setAddingCategoryError('Category name is required');
       return;
     }

     if (newCategoryName.length > 50) {
       setAddingCategoryError('Category name cannot exceed 50 characters');
       return;
     }

     if (newCategoryDescription.length > 200) {
       setAddingCategoryError('Description cannot exceed 200 characters');
       return;
     }

     setSavingCategory(true);
     setAddingCategoryError('');

     try {
       const categoryRepo = RepositoryFactory.getCategoryRepository();
       const newCategory = await categoryRepo.create(
         token,
         newCategoryName.trim(),
         newCategoryDescription.trim() || undefined
       );

       setCategories([...categories, newCategory]);
       setIsAddingCategory(false);
       setNewCategoryName('');
       setNewCategoryDescription('');
     } catch (error) {
       console.error('Failed to create category', error);
       const errorMessage = error instanceof Error ? error.message : 'Failed to create category';
       
       if (errorMessage.includes('duplicate') || errorMessage.includes('already exists')) {
         setAddingCategoryError('A category with this name already exists');
       } else {
         setAddingCategoryError(errorMessage || 'Failed to create category');
       }
     } finally {
       setSavingCategory(false);
      }
    };

  const handleDeleteClick = (categoryId: string) => {
    setDeletingCategoryId(categoryId);
    setDeleteError(null);
  };

  const handleCancelDelete = () => {
    setDeletingCategoryId(null);
    setDeleteError(null);
  };

  const handleConfirmDelete = async (categoryId: string) => {
    try {
      const categoryRepo = RepositoryFactory.getCategoryRepository();
      await categoryRepo.delete(token, categoryId);
      
      setCategories(categories.filter(c => c.id !== categoryId));
      setDeletingCategoryId(null);
      setDeleteError(null);
    } catch (error) {
      console.error('Failed to delete category', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to delete category';
      
      if (errorMessage.includes('in use') || errorMessage.includes('foreign key')) {
        setDeleteError('Cannot delete: Category is used by expenses. Merge categories instead.');
      } else {
        setDeleteError(errorMessage || 'Failed to delete category');
      }
    }
  };

  const handleEditClick = (category: Category) => {
    setEditingCategoryId(category.id);
    setEditName(category.name);
    setEditDescription(category.description || '');
    setEditError('');
    setShowMergeConfirm(false);
    setMergeTargetId(null);
    setMergeResult(null);
  };

  const handleCancelEdit = () => {
    setEditingCategoryId(null);
    setEditName('');
    setEditDescription('');
    setEditError('');
    setShowMergeConfirm(false);
    setMergeTargetId(null);
  };

  const handleSaveEdit = async () => {
    if (!editingCategoryId) return;

    if (!editName.trim()) {
      setEditError('Category name is required');
      return;
    }

    if (editName.length > 50) {
      setEditError('Category name cannot exceed 50 characters');
      return;
    }

    if (editDescription.length > 200) {
      setEditError('Description cannot exceed 200 characters');
      return;
    }

    setSavingEdit(true);
    setEditError('');

    try {
      const categoryRepo = RepositoryFactory.getCategoryRepository();
      const updatedCategory = await categoryRepo.update(
        token,
        editingCategoryId,
        editName.trim(),
        editDescription.trim() || undefined
      );

      setCategories(categories.map(c => 
        c.id === editingCategoryId ? updatedCategory : c
      ));
      handleCancelEdit();
    } catch (error) {
      console.error('Failed to update category', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to update category';
      
      if (errorMessage.includes('already exists') || errorMessage.includes('duplicate')) {
        const existingCategory = categories.find(
          c => c.name.toLowerCase() === editName.trim().toLowerCase() && c.id !== editingCategoryId
        );
        if (existingCategory) {
          setMergeTargetId(existingCategory.id);
          setShowMergeConfirm(true);
          setEditError('');
        } else {
          setEditError('A category with this name already exists');
        }
      } else {
        setEditError(errorMessage || 'Failed to update category');
      }
    } finally {
      setSavingEdit(false);
    }
  };

  const handleConfirmMerge = async () => {
    if (!editingCategoryId || !mergeTargetId) return;

    setMergingCategory(true);
    try {
      const categoryRepo = RepositoryFactory.getCategoryRepository();
      const result = await categoryRepo.merge(token, editingCategoryId, mergeTargetId);
      
      setMergeResult(result);
      setCategories(categories.filter(c => c.id !== editingCategoryId));
      
      setTimeout(() => {
        handleCancelEdit();
        setMergeResult(null);
      }, 3000);
    } catch (error) {
      console.error('Failed to merge categories', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to merge categories';
      setEditError(errorMessage);
      setShowMergeConfirm(false);
    } finally {
      setMergingCategory(false);
    }
  };

  const handleCancelMerge = () => {
    setShowMergeConfirm(false);
    setMergeTargetId(null);
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
                 {deleteError && (
                   <div className="text-red-400 text-sm p-3 bg-red-400/10 rounded-lg border border-red-400/20">
                     {deleteError}
                   </div>
                 )}
                 {categories.map((category) => (
                  <div 
                    key={category.id}
                    className={`p-4 rounded-xl border ${
                      category.is_default 
                        ? 'bg-white/5 border-white/5' 
                        : 'bg-black/20 border-white/10'
                    }`}
                  >
                    {editingCategoryId === category.id ? (
                      <div className="space-y-3">
                        {mergeResult ? (
                          <div className="text-green-400 text-sm p-3 bg-green-400/10 rounded-lg border border-green-400/20">
                            {mergeResult.message}
                          </div>
                        ) : showMergeConfirm ? (
                          <div className="space-y-3">
                            <div className="text-amber-400 text-sm p-3 bg-amber-400/10 rounded-lg border border-amber-400/20">
                              Category &apos;{editName}&apos; already exists. Merge expenses into &apos;{categories.find(c => c.id === mergeTargetId)?.name}&apos;?
                            </div>
                            <div className="flex items-center gap-2">
                              <button
                                onClick={handleConfirmMerge}
                                disabled={mergingCategory}
                                data-testid="merge-confirm-button"
                                className="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-primary/20 text-primary hover:bg-primary/30 transition-colors disabled:opacity-50"
                              >
                                <CheckIcon className="w-4 h-4" />
                                {mergingCategory ? 'Merging...' : 'Confirm Merge'}
                              </button>
                              <button
                                onClick={handleCancelMerge}
                                disabled={mergingCategory}
                                data-testid="merge-cancel-button"
                                className="flex-1 px-3 py-2 rounded-lg bg-white/10 text-text/60 hover:bg-white/20 transition-colors disabled:opacity-50"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        ) : (
                          <>
                            <div className="flex flex-col sm:flex-row gap-2">
                              <input
                                type="text"
                                value={editName}
                                onChange={(e) => setEditName(e.target.value)}
                                placeholder="Category name"
                                maxLength={50}
                                className="flex-1 bg-black/20 border border-white/10 rounded-lg px-3 py-2 text-sm text-text focus:border-primary/50 outline-none"
                                autoFocus
                              />
                              <input
                                type="text"
                                value={editDescription}
                                onChange={(e) => setEditDescription(e.target.value)}
                                placeholder="Description (optional)"
                                maxLength={200}
                                className="flex-1 bg-black/20 border border-white/10 rounded-lg px-3 py-2 text-sm text-text focus:border-primary/50 outline-none"
                              />
                            </div>
                            {editError && (
                              <div className="text-red-400 text-sm p-2 bg-red-400/10 rounded-lg border border-red-400/20">
                                {editError}
                              </div>
                            )}
                            <div className="flex items-center gap-2">
                              <button
                                onClick={handleSaveEdit}
                                disabled={savingEdit || !editName.trim()}
                                data-testid="save-edit-button"
                                className="p-2 rounded-lg bg-primary/20 text-primary hover:bg-primary/30 transition-colors disabled:opacity-50"
                              >
                                <CheckIcon className="w-4 h-4" />
                              </button>
                              <button
                                onClick={handleCancelEdit}
                                disabled={savingEdit}
                                data-testid="cancel-edit-button"
                                className="p-2 rounded-lg bg-white/10 text-text/60 hover:bg-white/20 transition-colors disabled:opacity-50"
                              >
                                <XMarkIcon className="w-4 h-4" />
                              </button>
                            </div>
                          </>
                        )}
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
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
                        {deletingCategoryId === category.id ? (
                          <div className="flex items-center gap-2 ml-4">
                            <span className="text-sm text-text/70">Delete?</span>
                            <button 
                              onClick={() => handleConfirmDelete(category.id)}
                              className="p-2 rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
                              data-testid="confirm-delete-button"
                            >
                              <CheckIcon className="w-4 h-4" />
                            </button>
                            <button 
                              onClick={handleCancelDelete}
                              className="p-2 rounded-lg bg-white/10 text-text/60 hover:bg-white/20 transition-colors"
                              data-testid="cancel-delete-button"
                            >
                              <XMarkIcon className="w-4 h-4" />
                            </button>
                          </div>
                        ) : (
                          !category.is_default && (
                            <div className="flex items-center gap-2 ml-4">
                              <button
                                onClick={() => handleEditClick(category)}
                                className="p-2 rounded-lg text-text/40 hover:text-primary hover:bg-primary/10 transition-colors"
                                data-testid="edit-category-button"
                              >
                                <PencilSquareIcon className="w-4 h-4" />
                              </button>
                              <button
                                onClick={() => handleDeleteClick(category.id)}
                                className="p-2 rounded-lg text-text/40 hover:text-red-400 hover:bg-red-500/10 transition-colors"
                                data-testid="delete-category-button"
                              >
                                <TrashIcon className="w-4 h-4" />
                              </button>
                            </div>
                          )
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Add Category Button and Form */}
            {!isAddingCategory && (
              <button
                onClick={() => setIsAddingCategory(true)}
                data-testid="add-category-button"
                className="w-full py-3 px-4 rounded-xl border-2 border-dashed border-white/20 text-text/60 hover:text-text hover:border-white/40 transition-all font-medium"
              >
                + Add Category
              </button>
            )}

            {/* Inline Add Category Form */}
            {isAddingCategory && (
              <div className="p-4 rounded-xl bg-black/20 border border-white/10 space-y-4">
                <div>
                  <label className="text-sm font-medium text-text mb-2 block">
                    Category Name <span className="text-red-400">*</span>
                  </label>
                  <input
                    type="text"
                    value={newCategoryName}
                    onChange={(e) => setNewCategoryName(e.target.value)}
                    placeholder="Enter category name"
                    maxLength={50}
                    data-testid="category-name-input"
                    className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-text focus:outline-none focus:ring-2 focus:ring-primary/50 placeholder-text/40"
                  />
                  <p className="text-xs text-text/40 mt-1">
                    {newCategoryName.length}/50
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium text-text mb-2 block">
                    Description <span className="text-text/40">(optional)</span>
                  </label>
                  <textarea
                    value={newCategoryDescription}
                    onChange={(e) => setNewCategoryDescription(e.target.value)}
                    placeholder="Enter category description"
                    maxLength={200}
                    data-testid="category-description-input"
                    className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-text focus:outline-none focus:ring-2 focus:ring-primary/50 placeholder-text/40 resize-none"
                    rows={3}
                  />
                  <p className="text-xs text-text/40 mt-1">
                    {newCategoryDescription.length}/200
                  </p>
                </div>

                {addingCategoryError && (
                  <div className="text-red-400 text-sm p-3 bg-red-400/10 rounded-lg border border-red-400/20">
                    {addingCategoryError}
                  </div>
                )}

                <div className="flex gap-3">
                  <button
                    onClick={handleAddCategory}
                    disabled={savingCategory || !newCategoryName.trim()}
                    data-testid="save-category-button"
                    className={`flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-xl font-medium transition-all ${
                      savingCategory || !newCategoryName.trim()
                        ? 'bg-white/5 text-text/30 cursor-not-allowed'
                        : 'bg-primary text-white hover:bg-primary/90 shadow-lg shadow-primary/20'
                    }`}
                  >
                    {savingCategory ? 'Saving...' : (
                      <>
                        <CheckIcon className="w-5 h-5" />
                        Save Category
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => {
                      setIsAddingCategory(false);
                      setNewCategoryName('');
                      setNewCategoryDescription('');
                      setAddingCategoryError('');
                    }}
                    disabled={savingCategory}
                    className="flex-1 px-4 py-3 rounded-xl font-medium transition-all bg-white/10 text-text hover:bg-white/20 disabled:opacity-50"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
         </div>
       </div>
     </DashboardLayout>
   );
 }
