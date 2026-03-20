import tailwindcssAnimate from 'tailwindcss-animate'

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
        ring: '#171717',
        background: '#fafafa',
        foreground: '#171717',
        primary: {
          DEFAULT: '#171717',
          foreground: '#ffffff',
        },
        secondary: {
          DEFAULT: '#f5f5f5',
          foreground: '#171717',
        },
        muted: {
          DEFAULT: '#f5f5f5',
          foreground: '#737373',
        },
        accent: {
          DEFAULT: '#f5f5f5',
          foreground: '#171717',
        },
        card: {
          DEFAULT: '#ffffff',
          foreground: '#171717',
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
  plugins: [tailwindcssAnimate],
}

