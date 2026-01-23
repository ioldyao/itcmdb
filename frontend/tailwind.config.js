/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class', // 使用 class 模式控制深色主题
  theme: {
    extend: {
      colors: {
        // 深色主题配色
        'bg-primary': '#0a0e1a',
        'bg-secondary': '#141824',
        'bg-tertiary': '#1e2330',
        // 浅色主题配色
        'bg-primary-light': '#ffffff',
        'bg-secondary-light': '#f5f7fa',
        'bg-tertiary-light': '#e8ecef',
        // 品牌色
        'brand-primary': '#3a84ff',
        'brand-hover': '#5a9dff',
        'brand-active': '#2a74ef',
        // 文字色
        'text-primary': '#e5e5e5',
        'text-secondary': '#8b8d98',
        'text-tertiary': '#5e5f6b',
        'text-primary-light': '#1f2937',
        'text-secondary-light': '#6b7280',
        'text-tertiary-light': '#9ca3af',
      },
      backgroundImage: {
        'gradient-card': 'linear-gradient(135deg, #1e2330 0%, #141824 100%)',
        'gradient-card-light': 'linear-gradient(135deg, #ffffff 0%, #f9fafb 100%)',
      },
      backdropBlur: {
        'glass': '10px',
      },
      boxShadow: {
        'card': '0 4px 16px rgba(0, 0, 0, 0.4)',
        'card-hover': '0 8px 24px rgba(0, 0, 0, 0.5)',
        'card-light': '0 2px 8px rgba(0, 0, 0, 0.08)',
        'card-hover-light': '0 4px 12px rgba(0, 0, 0, 0.12)',
        'glass': '0 8px 32px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.05)',
      },
    },
  },
  plugins: [],
}
