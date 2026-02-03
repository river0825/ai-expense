import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#F59E0B', // Amber 500
          foreground: '#FFFFFF',
        },
        secondary: {
          DEFAULT: '#FBBF24', // Amber 400
          foreground: '#0F172A',
        },
        cta: {
          DEFAULT: '#8B5CF6', // Violet 500
          foreground: '#FFFFFF',
        },
        background: '#0F172A', // Slate 900
        surface: '#1E293B', // Slate 800 (for cards/glass)
        text: {
          DEFAULT: '#F8FAFC', // Slate 50
          muted: '#94A3B8', // Slate 400
        },
        border: 'rgba(255, 255, 255, 0.1)',
      },
      fontFamily: {
        sans: ['"IBM Plex Sans"', 'sans-serif'],
        mono: ['"Fira Code"', 'monospace'],
      },
      boxShadow: {
        'glass-sm': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        'glass-md': '0 8px 30px rgba(0, 0, 0, 0.12)',
        'glass-lg': '0 30px 60px -12px rgba(0, 0, 0, 0.25)',
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'glass': 'linear-gradient(145deg, rgba(30, 41, 59, 0.7) 0%, rgba(15, 23, 42, 0.6) 100%)',
      },
    },
  },
  plugins: [],
}
export default config
