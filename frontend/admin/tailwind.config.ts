import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        text: 'hsl(var(--text))',
        surface: 'hsl(var(--surface))',
        border: 'hsl(var(--border))',
        card: 'hsl(var(--card))',
        primary: 'hsl(var(--primary))',
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
      },
      boxShadow: {
        'glass-sm': '0 8px 20px rgba(2, 6, 23, 0.2)',
        'glass-md': '0 14px 34px rgba(2, 6, 23, 0.28)',
      },
    },
  },
  plugins: [],
};

export default config;
