-- Drop all initial schema objects in reverse order of creation to respect foreign keys

-- Drop indexes first
DROP INDEX IF EXISTS idx_keywords_category;
DROP INDEX IF EXISTS idx_categories_user;
DROP INDEX IF EXISTS idx_expenses_user_category;
DROP INDEX IF EXISTS idx_expenses_user_date;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS category_keywords;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
