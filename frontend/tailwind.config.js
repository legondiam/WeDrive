import tailwindcssAnimate from 'tailwindcss-animate'
import { setupInspiraUI } from '@inspira-ui/plugins'

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{js,vue,ts,tsx}',
  ],
  theme: {
    extend: {
      borderRadius: {
        lg: '10px',
        md: '10px',
        sm: '10px',
      },
      colors: {
        border: '#e5e5e5',
        input: '#e5e5e5',
        ring: '#18181b',
        background: '#fafafa',
        foreground: '#18181b',
        primary: {
          DEFAULT: '#18181b',
          foreground: '#ffffff',
        },
        secondary: {
          DEFAULT: '#f5f5f5',
          foreground: '#18181b',
        },
        muted: {
          DEFAULT: '#f5f5f5',
          foreground: '#71717a',
        },
        accent: {
          DEFAULT: '#f5f5f5',
          foreground: '#18181b',
        },
        card: {
          DEFAULT: '#ffffff',
          foreground: '#18181b',
        },
      },
      boxShadow: {
        sm: '0 1px 2px rgba(0,0,0,0.06)',
        md: '0 8px 24px rgba(0,0,0,0.08)',
      },
      keyframes: {
        'accordion-down': {
          from: { height: 0 },
          to: { height: 'var(--radix-accordion-content-height)' },
        },
        'accordion-up': {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: 0 },
        },
      },
      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out',
      },
    },
  },
  plugins: [tailwindcssAnimate, setupInspiraUI],
}
