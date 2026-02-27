import { defineStore } from 'pinia';

export type ThemeType = 'oak' | 'mcp';

export const useThemeStore = defineStore('theme', {
  state: () => ({
    currentTheme: (localStorage.getItem('theme') as ThemeType) || 'oak',
  }),
  actions: {
    setTheme(theme: ThemeType) {
      this.currentTheme = theme;
      localStorage.setItem('theme', theme);
      this.applyTheme();
    },
    applyTheme() {
      document.documentElement.setAttribute('data-theme', this.currentTheme);
    }
  },
});
