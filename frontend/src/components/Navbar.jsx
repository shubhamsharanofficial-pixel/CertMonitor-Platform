import React, { useState } from 'react'; // 1. Import useState
import { Shield, Key, Server, FileText, Settings, Download, Cloud, HelpCircle } from 'lucide-react'; // 2. Add HelpCircle
import { Link, useLocation } from 'react-router-dom';
import AboutModal from './AboutModal'; // 3. Import the Modal

export default function Navbar({ user, logout, onRotate, isGenerating }) {
  const location = useLocation();
  const [showAbout, setShowAbout] = useState(false); // 4. Initialize State

  // Helper to check active state
  const isActive = (path) => location.pathname === path 
    ? "text-blue-600 bg-blue-50" 
    : "text-slate-600 hover:text-slate-900 hover:bg-slate-50";

  return (
    <>
      <nav className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between sticky top-0 z-10">
        <div className="flex items-center gap-8">
          {/* Logo */}
          <div className="flex items-center gap-3">
            <img src="/favicon.png" alt="CertMonitor Logo" className="h-12 w-auto" />
            <h1 className="text-2xl font-bold tracking-tight text-slate-900">CertMonitor</h1>
          </div>

          {/* Navigation Links */}
          <div className="hidden md:flex items-center gap-2">
            <Link 
              to="/dashboard" 
              className={`px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${isActive('/dashboard')}`}
            >
              <FileText className="w-4 h-4" />
              Certificates
            </Link>

            <Link 
              to="/agents" 
              className={`px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${isActive('/agents')}`}
            >
              <Server className="w-4 h-4" />
              Agents
            </Link>

            <Link 
              to="/cloud" 
              className={`px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${isActive('/cloud')}`}
            >
              <Cloud className="w-4 h-4 text-sky-500" />
              Cloud Monitors
            </Link>

            <Link 
              to="/downloads" 
              className={`px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${isActive('/downloads')}`}
            >
              <Download className="w-4 h-4" />
              Downloads
            </Link>

            <Link 
              to="/settings" 
              className={`px-3 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${isActive('/settings')}`}
            >
              <Settings className="w-4 h-4" />
              Settings
            </Link>
          </div>
        </div>

        <div className="flex items-center gap-4">
          
          {/* NEW: Help / About Button */}
          <button 
            onClick={() => setShowAbout(true)}
            className="p-2 text-slate-500 hover:text-blue-600 hover:bg-slate-50 rounded-full transition-colors"
            title="About Developer"
          >
            <HelpCircle className="w-5 h-5" />
          </button>
          
          <div className="h-6 w-px bg-slate-200 mx-1"></div> {/* Vertical Divider */}

          {/* Rotate Button */}
          {user?.has_api_key && onRotate && (
            <button 
              onClick={onRotate}
              disabled={isGenerating}
              className="hidden md:flex items-center gap-2 px-3 py-1.5 text-xs font-medium text-slate-600 border border-slate-300 rounded hover:bg-slate-50 hover:text-red-600 hover:border-red-200 transition-colors"
              title="Regenerate API Key"
            >
              <Key className="w-3 h-3" />
              {isGenerating ? "Rotating..." : "Rotate Key"}
            </button>
          )}

          <div className="text-right hidden md:block">
            <div className="text-sm font-medium text-slate-900">{user?.organization_name}</div>
            <div className="text-xs text-slate-500">{user?.email}</div>
          </div>
          <button 
            onClick={logout} 
            className="text-sm text-red-600 hover:text-red-700 font-medium"
          >
            Logout
          </button>
        </div>
      </nav>

      {/* 5. Render Modal Outside Nav */}
      <AboutModal isOpen={showAbout} onClose={() => setShowAbout(false)} />
    </>
  );
}