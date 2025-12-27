import React, { useState, useEffect } from 'react';
import { User, Bell, Building, Save, CheckCircle, AlertCircle } from 'lucide-react';
import api from '../services/api';
import { useAuth } from '../context';
import Navbar from '../components/Navbar';

export default function Settings() {
  // Destructure updateUser from context
  const { user, logout, updateUser } = useAuth();
  
  const [formData, setFormData] = useState({
    organization_name: '',
    email_enabled: true
  });
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);

  // 1. Fetch Profile on Mount
  useEffect(() => {
    const loadProfile = async () => {
      try {
        const response = await api.getProfile();
        setFormData({
            organization_name: response.data.organization_name,
            email_enabled: response.data.email_enabled
        });
      } catch (err) {
        console.error("Failed to load profile", err);
        setMessage({ type: 'error', text: 'Failed to load profile settings. Please refresh.' });
      } finally {
        setLoading(false);
      }
    };
    loadProfile();
  }, []);

  // 2. Handle Save
  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);

    try {
      await api.updateProfile(formData);
      
      // NEW: Update Global Context immediately
      updateUser({ 
        organization_name: formData.organization_name, 
        email_enabled: formData.email_enabled 
      });

      setMessage({ type: 'success', text: 'Settings saved successfully.' });
      
      setTimeout(() => setMessage(null), 3000);
    } catch (err) {
      console.error(err);
      setMessage({ type: 'error', text: 'Failed to save settings.' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <Navbar user={user} logout={logout} />

      <main className="max-w-3xl mx-auto px-6 py-8">
        
        <div className="mb-8">
            <h2 className="text-2xl font-bold text-slate-900">Account Settings</h2>
            <p className="text-slate-500 mt-1">Manage your organization profile and notification preferences.</p>
        </div>

        {message && (
            <div className={`mb-6 px-4 py-3 rounded-lg flex items-center gap-2 text-sm font-medium ${
                message.type === 'success' 
                    ? 'bg-green-50 text-green-700 border border-green-200' 
                    : 'bg-red-50 text-red-700 border border-red-200'
            }`}>
                {message.type === 'success' ? <CheckCircle className="w-5 h-5"/> : <AlertCircle className="w-5 h-5"/>}
                {message.text}
            </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
            
            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
                <div className="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex items-center gap-2">
                    <User className="w-5 h-5 text-slate-500" />
                    <h3 className="font-semibold text-slate-800">Organization Profile</h3>
                </div>
                <div className="p-6 space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">Organization Name</label>
                        <div className="relative">
                            <Building className="absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
                            <input 
                                type="text"
                                value={formData.organization_name}
                                onChange={(e) => setFormData({...formData, organization_name: e.target.value})}
                                className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                required
                            />
                        </div>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">Email Address</label>
                        <input 
                            type="email" 
                            value={user.email} 
                            disabled 
                            className="w-full px-4 py-2 border border-slate-200 bg-slate-50 text-slate-500 rounded-lg cursor-not-allowed"
                        />
                        <p className="text-xs text-slate-400 mt-1">Email cannot be changed manually.</p>
                    </div>
                </div>
            </div>

            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
                <div className="px-6 py-4 border-b border-slate-100 bg-slate-50/50 flex items-center gap-2">
                    <Bell className="w-5 h-5 text-slate-500" />
                    <h3 className="font-semibold text-slate-800">Notifications</h3>
                </div>
                <div className="p-6">
                    <div className="flex items-start gap-4">
                        <div className="flex items-center h-5 mt-1">
                            <input
                                id="email_enabled"
                                type="checkbox"
                                checked={formData.email_enabled}
                                onChange={(e) => setFormData({...formData, email_enabled: e.target.checked})}
                                className="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500 cursor-pointer"
                            />
                        </div>
                        <div>
                            <label htmlFor="email_enabled" className="font-medium text-slate-900 cursor-pointer">
                                Enable Email Alerts
                            </label>
                            <p className="text-slate-500 text-sm mt-1">
                                Receive daily digests about certificates expiring within 30 days. 
                                We filter duplicates so you won't get spammed.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex justify-end">
                <button 
                    type="submit" 
                    disabled={saving}
                    className="flex items-center gap-2 px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-sm disabled:opacity-70 disabled:cursor-not-allowed transition-all"
                >
                    {saving ? (
                        <>Saving...</>
                    ) : (
                        <>
                            <Save className="w-4 h-4" />
                            Save Changes
                        </>
                    )}
                </button>
            </div>

        </form>
      </main>
    </div>
  );
}