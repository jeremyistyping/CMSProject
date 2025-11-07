import { useLanguage } from '@/contexts/LanguageContext';

/**
 * Custom hook for accessing translation functionality
 * @returns Translation function and current language
 */
export const useTranslation = () => {
  const { t, language, setLanguage } = useLanguage();
  
  return {
    t,
    language,
    setLanguage
  };
};
