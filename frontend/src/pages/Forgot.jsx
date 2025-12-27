import React, { useState } from 'react';
import { Mail, ArrowRight, ArrowLeft, CheckCircle, Shield } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context';

export default function Forgot() {
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState('idle'); // idle | loading | success | error
  const { forgotPassword } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('loading');
    try {
      await forgotPassword(email);
      setStatus('success');
    } catch (err) {
      // We don't want to reveal if email exists or not for security, 
      // but if API fails due to network, we can show generic error.
      // Ideally backend returns 200 OK even if email not found.
      setStatus('success'); 
    }
  };

  if (status === 'success') {
      return (
        <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
            <div className="max-w-md w-full bg-white rounded-xl shadow-lg border border-slate-100 p-8 text-center animate-in fade-in zoom-in duration-300">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4 mx-auto">
                    <CheckCircle className="w-8 h-8 text-green-600" />
                </div>
                <h2 className="text-2xl font-bold text-slate-900">Check Your Email</h2>
                <p className="text-slate-500 mt-2 mb-6">
                    If an account exists for <strong>{email}</strong>, we have sent password reset instructions.
                </p>
                <Link to="/login" className="text-blue-600 font-semibold hover:text-blue-700 flex items-center justify-center gap-2">
                    <ArrowLeft className="w-4 h-4" /> Back to Login
                </Link>
            </div>
        </div>
      );
  }

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-xl shadow-lg border border-slate-100 overflow-hidden">
        <div className="bg-slate-900 p-8 text-center">
          <div className="mx-auto bg-blue-600 w-12 h-12 rounded-lg flex items-center justify-center mb-4">
            <Shield className="text-white w-7 h-7" />
          </div>
          <h2 className="text-2xl font-bold text-white">Reset Password</h2>
          <p className="text-slate-400 mt-2 text-sm">Enter your email to receive instructions</p>
        </div>

        <div className="p-8">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Email Address</label>
              <div className="relative">
                <Mail className="absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
                <input 
                  type="email" 
                  className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all"
                  placeholder="you@company.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
            </div>

            <button 
              type="submit" 
              disabled={status === 'loading'}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-semibold py-2.5 rounded-lg transition-colors flex items-center justify-center gap-2 group"
            >
              {status === 'loading' ? 'Sending...' : 'Send Reset Link'}
              {!status === 'loading' && <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />}
            </button>
          </form>
            
          <div className="mt-6 text-center">
            <Link to="/login" className="text-slate-600 text-sm font-medium hover:text-slate-900 flex items-center justify-center gap-2">
              <ArrowLeft className="w-4 h-4" /> Back to Login
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}