import React from 'react';
import { Mail, X, ExternalLink, Copy, Check } from 'lucide-react';

export default function AboutModal({ isOpen, onClose }) {
  const [copied, setCopied] = React.useState(null);

  if (!isOpen) return null;

  const copyToClipboard = (text, type) => {
    navigator.clipboard.writeText(text);
    setCopied(type);
    setTimeout(() => setCopied(null), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md border border-slate-100 overflow-hidden relative animate-in fade-in zoom-in duration-200">
        
        {/* Header Background */}
        <div className="h-24 bg-gradient-to-r from-blue-600 to-indigo-600 relative">
          <button 
            onClick={onClose}
            className="absolute top-4 right-4 p-1.5 bg-white/10 hover:bg-white/20 text-white rounded-full transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Profile Section */}
        <div className="px-6 pb-8">
          <div className="relative -mt-12 mb-4">
             {/* Initials Avatar */}
            <div className="w-24 h-24 rounded-full bg-white p-1.5 shadow-lg inline-block">
              <div className="w-full h-full rounded-full bg-slate-100 flex items-center justify-center text-2xl font-bold text-slate-600 border border-slate-200">
                <img 
                    src="/profile.png" 
                    alt="SS" 
                    className="w-full h-full rounded-full object-cover border border-slate-200"
                />
              </div>
            </div>
          </div>

          <div className="mb-6">
            <h2 className="text-2xl font-bold text-slate-900">Shubham Sharan</h2>
            <p className="text-slate-500 font-medium">Software Engineer & Creator</p>
            <p className="text-sm text-slate-400 mt-1">
              Building scalable tools for the modern web. CertMonitor is designed to solve real infrastructure visibility problems.
            </p>
          </div>

          {/* Links Section */}
          <div className="space-y-3">
            
            {/* GitHub */}
            <a 
              href="https://github.com/shubhamsharanofficial-pixel/CertMonitor-Platform" 
              target="_blank" 
              rel="noopener noreferrer"
              className="flex items-center justify-between p-3 rounded-xl bg-slate-50 hover:bg-slate-100 border border-slate-200 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <div className="bg-white p-2 rounded-lg shadow-sm">
                  {/* Custom GitHub SVG */}
                  <GithubIcon className="w-5 h-5 text-slate-900" />
                </div>
                <div>
                  <div className="font-semibold text-slate-900">GitHub Project</div>
                  <div className="text-xs text-slate-500">View Source Code</div>
                </div>
              </div>
              <ExternalLink className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
            </a>

            {/* LinkedIn */}
            <a 
              href="https://www.linkedin.com/in/shubham-sharan-56765b226" 
              target="_blank" 
              rel="noopener noreferrer"
              className="flex items-center justify-between p-3 rounded-xl bg-slate-50 hover:bg-slate-100 border border-slate-200 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <div className="bg-white p-2 rounded-lg shadow-sm">
                  {/* Custom LinkedIn SVG */}
                  <LinkedinIcon className="w-5 h-5 text-blue-700" />
                </div>
                <div>
                  <div className="font-semibold text-slate-900">LinkedIn</div>
                  <div className="text-xs text-slate-500">Let's Connect</div>
                </div>
              </div>
              <ExternalLink className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
            </a>

            {/* Email */}
            <button 
              onClick={() => copyToClipboard('shubhamsharanofficial@gmail.com', 'email')}
              className="w-full flex items-center justify-between p-3 rounded-xl bg-slate-50 hover:bg-slate-100 border border-slate-200 transition-colors group text-left"
            >
              <div className="flex items-center gap-3">
                <div className="bg-white p-2 rounded-lg shadow-sm">
                  <Mail className="w-5 h-5 text-red-500" />
                </div>
                <div className="overflow-hidden">
                  <div className="font-semibold text-slate-900">Email Me</div>
                  <div className="text-xs text-slate-500 truncate">shubhamsharanofficial@gmail.com</div>
                </div>
              </div>
              {copied === 'email' ? (
                <Check className="w-4 h-4 text-green-600" />
              ) : (
                <Copy className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
              )}
            </button>

          </div>
        </div>

        {/* Footer */}
        <div className="bg-slate-50 px-6 py-4 border-t border-slate-100 text-center">
            <p className="text-xs text-slate-400">© 2026 CertMonitor v1.0.0</p>
        </div>
      </div>
    </div>
  );
}

// --- Custom Icon Components (Reliable SVGs) ---

function GithubIcon({ className }) {
  return (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
      <path d="M9 18c-4.51 2-5-2-7-2" />
    </svg>
  );
}

function LinkedinIcon({ className }) {
  return (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z" />
      <rect width="4" height="12" x="2" y="9" />
      <circle cx="4" cy="4" r="2" />
    </svg>
  );
}