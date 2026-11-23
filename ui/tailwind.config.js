/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        opti: {
          DEFAULT: 'rgb(var(--color-primary-rgb, 11 86 148) / <alpha-value>)',
          dark: 'rgb(var(--color-primary-dark-rgb, 8 63 110) / <alpha-value>)',
          light: 'var(--color-primary-light, #e6f0f9)',
          soft: 'var(--color-primary-soft, #77a3c7)'
        },
        assist: {
          page: {
            light: '#f5f7fb',
            dark: '#0a0f1d'
          },
          sidebar: {
            light: '#ffffff',
            dark: '#0e1626'
          },
          card: {
            light: '#ffffff',
            dark: '#121b2d'
          },
          border: {
            light: '#e2e8f0',
            dark: '#1e2d47'
          }
        },
        slate: {
          700: '#1e293b',
          750: '#182438',
          800: '#111928',
          850: '#0e1626',
          900: '#0a0f1d',
          950: '#060a14',
        }
      }
    },
  },
  plugins: [],
}
