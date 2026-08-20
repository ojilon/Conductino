/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: ['class', '[data-theme="aurora-dark"]'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--bg)',
        elev: 'var(--bg-elev)',
        elev2: 'var(--bg-elev-2)',
        fg: 'var(--fg)',
        muted: 'var(--muted)',
        accent: 'var(--accent)',
        accent2: 'var(--accent2)',
        border: 'var(--border)',
        danger: 'var(--danger)',
        'tab-active': 'var(--tab-active)',
      },
      borderRadius: {
        app: 'var(--radius)',
        'app-sm': 'var(--radius-sm)',
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'sans-serif',
        ],
      },
      spacing: {
        tabstrip: 'var(--tabstrip-h)',
        toolbar: 'var(--toolbar-h)',
        sidebar: 'var(--sidebar-w)',
      },
    },
  },
  plugins: [],
}
