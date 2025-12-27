import React, { useState, useEffect } from 'react';
import api from '../services/api';
import { AuthContext } from './AuthContextUtils';

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [loading, setLoading] = useState(true);

  // 1. Check for existing session
  useEffect(() => {
    const storedUser = localStorage.getItem('user');
    if (token && storedUser) {
      setUser(JSON.parse(storedUser));
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    }
    setLoading(false);
  }, []);

  // 2. Login Action
  const login = async (email, password) => {
    const response = await api.login(email, password);
    const { token: newToken, user: userData } = response.data;
    
    setToken(newToken);
    setUser(userData);
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(userData));
    
    api.defaults.headers.common['Authorization'] = `Bearer ${newToken}`;
    return userData;
  };

  // 3. Signup Action (UPDATED: No longer logs in automatically)
  const signup = async (email, password, orgName) => {
    // This now sends a verification email and returns success message
    await api.signup(email, password, orgName);
    return true; 
  };

  // 4. Logout Action
  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    delete api.defaults.headers.common['Authorization'];
  };

  // 5. Generate API Key Action
  const generateApiKey = async () => {
    const response = await api.regenerateKey();
    const { api_key } = response.data;

    if (user) {
        const updatedUser = { ...user, has_api_key: true };
        setUser(updatedUser);
        localStorage.setItem('user', JSON.stringify(updatedUser));
    }

    return api_key;
  };

  // 6. Update Profile
  const updateUser = (updates) => {
    if (!user) return;
    const updatedUser = { ...user, ...updates };
    setUser(updatedUser);
    localStorage.setItem('user', JSON.stringify(updatedUser));
  };

  // 7. NEW: Account Recovery Wrappers
  const verifyEmail = async (token) => {
      await api.verifyEmail(token);
  };

  const forgotPassword = async (email) => {
      await api.forgotPassword(email);
  };

  const resetPassword = async (token, newPassword) => {
      await api.resetPassword(token, newPassword);
  };

  return (
    <AuthContext.Provider value={{ 
      user, 
      token, 
      login, 
      signup, 
      logout, 
      generateApiKey, 
      updateUser,
      verifyEmail,
      forgotPassword,
      resetPassword,
      loading 
    }}>
      {!loading && children}
    </AuthContext.Provider>
  );
};