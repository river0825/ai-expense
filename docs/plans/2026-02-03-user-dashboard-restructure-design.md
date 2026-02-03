# User Dashboard Restructure Design

**Date:** 2026-02-03  
**Domain:** `my.aiexpense.net`  
**Purpose:** Remove admin features from user dashboard, restructure URLs for user-facing app

## URL Structure

| Path | Content |
|------|---------|
| `/` | Redirect to `/dashboard` |
| `/dashboard` | Expense list (user's expenses view) |
| `/dashboard/settings` | User settings |

## Changes Required

### Pages to Delete (Admin Features)

- `src/app/[locale]/page.tsx` - Admin overview page
- `src/app/[locale]/reports/` - Admin reports page

### Pages to Move

| From | To |
|------|-----|
| `src/app/[locale]/user/reports/page.tsx` | `src/app/[locale]/dashboard/page.tsx` |
| `src/app/[locale]/user/settings/page.tsx` | `src/app/[locale]/dashboard/settings/page.tsx` |

### Folders to Delete

- `src/app/[locale]/user/` - After moving contents

### Root Redirect

Create `src/app/[locale]/page.tsx` that redirects to `/dashboard`.

### Sidebar Updates

Current navigation:
```ts
const NAVIGATION = [
  { name: 'dashboard', href: '/', icon: HomeIcon },
  { name: 'reports', href: '/reports', icon: ChartBarIcon },
  { name: 'my_expenses', href: '/user/reports', icon: UserCircleIcon },
];
```

New navigation:
```ts
const NAVIGATION = [
  { name: 'my_expenses', href: '/dashboard', icon: HomeIcon },
];
```

Settings link: Update from `/user/settings` to `/dashboard/settings`.

Remove path-based filtering logic (no longer needed - single user context).

## Future Considerations

- Admin dashboard will be recreated separately in `frontend/admin/`
- User dashboard at `my.aiexpense.net` serves only authenticated users
- Group features planned for `frontend/group/`
