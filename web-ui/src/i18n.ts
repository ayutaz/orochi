import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import enTranslations from './locales/en.json';
import jaTranslations from './locales/ja.json';

const savedLanguage = localStorage.getItem('language') || 'ja';

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: enTranslations },
    ja: { translation: jaTranslations },
  },
  lng: savedLanguage,
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
