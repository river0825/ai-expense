import { redirect } from 'next/navigation';

import { routing } from '@/i18n/routing';

type DashboardRedirectPageProps = {
  searchParams?: Record<string, string | string[] | undefined>;
};

export default function DashboardRedirectPage({
  searchParams = {},
}: DashboardRedirectPageProps) {
  const params = new URLSearchParams();

  Object.entries(searchParams).forEach(([key, value]) => {
    if (Array.isArray(value)) {
      value.forEach((entry) => {
        if (typeof entry === 'string') {
          params.append(key, entry);
        }
      });
    } else if (typeof value === 'string') {
      params.append(key, value);
    }
  });

  const queryString = params.toString();
  const locale = routing.defaultLocale ?? 'en';
  const target = `/${locale}/dashboard${queryString ? `?${queryString}` : ''}`;

  redirect(target);
}
