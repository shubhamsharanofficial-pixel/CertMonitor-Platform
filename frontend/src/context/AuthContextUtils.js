import { createContext, useContext } from 'react';

// 1. Create the Context Object
export const AuthContext = createContext(null);

// 2. Create the Hook
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};