/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          dark: 'rgb(var(--brand-dark) / <alpha-value>)',
          primary: 'rgb(var(--brand-primary) / <alpha-value>)',
          secondary: 'rgb(var(--brand-secondary) / <alpha-value>)',
          accent: 'rgb(var(--brand-accent) / <alpha-value>)',
          'status-green': 'rgb(var(--brand-status-green) / <alpha-value>)',
          'status-red': 'rgb(var(--brand-status-red) / <alpha-value>)',
        }
      }
    },
  },
  plugins: [],
}