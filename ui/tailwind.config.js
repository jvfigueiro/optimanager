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
          DEFAULT: 'var(--color-primary, #0B5694)',
          dark: 'var(--color-primary-dark, #083f6e)',
          light: 'var(--color-primary-light, #e6f0f9)'
        }
      }
    },
  },
  plugins: [],
}
