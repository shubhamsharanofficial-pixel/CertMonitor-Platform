import React, { useState } from 'react';
import { Copy, Check, AlertTriangle, X, Terminal, FileCode } from 'lucide-react';

export default function ApiKeyModal({ isOpen, onClose, apiKey }) {
  const [copiedCommand, setCopiedCommand] = useState(false);
  const [copiedKey, setCopiedKey] = useState(false);
  const [tool, setTool] = useState('curl'); // 'curl' | 'wget'

  if (!isOpen) return null;

  // Determine Backend URL
  const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
  const backendOrigin = isLocal ? 'http://localhost:8080' : window.location.origin;
  
  // Define Commands
  const commands = {
    curl: `curl -sL ${backendOrigin}/api/agent/install | sudo bash -s -- -k ${apiKey}`,
    wget: `wget -qO- ${backendOrigin}/api/agent/install | sudo bash -s -- -k ${apiKey}`
  };

  const activeCommand = commands[tool];

  const handleCopyCommand = () => {
    navigator.clipboard.writeText(activeCommand);
    setCopiedCommand(true);
    setTimeout(() => setCopiedCommand(false), 2000);
  };

  const handleCopyKey = () => {
    navigator.clipboard.writeText(apiKey);
    setCopiedKey(true);
    setTimeout(() => setCopiedKey(false), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4">
      <div className="bg-white rounded-xl shadow-2xl max-w-2xl w-full overflow-hidden border border-slate-200 animate-in fade-in zoom-in duration-200">
        
        {/* Header */}
        <div className="bg-slate-50 px-6 py-4 border-b border-slate-100 flex justify-between items-center">
          <h3 className="font-semibold text-slate-800 text-lg">Connect Your Server</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-6 space-y-6">
          
          {/* Warning */}
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 flex gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0" />
            <div className="text-sm text-yellow-800">
              <p className="font-medium">Save this key immediately!</p>
              <p className="mt-1 opacity-90">
                We only show this once. Once you close this window, the key is gone forever.
              </p>
            </div>
          </div>

          {/* Section 1: Auto-Install Command */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="flex items-center gap-2 text-sm font-medium text-slate-700">
                <Terminal className="w-4 h-4 text-blue-600" />
                Auto-Install Command (Linux)
              </label>
              
              {/* Tool Switcher */}
              <div className="flex bg-slate-100 p-1 rounded-lg">
                <button
                  onClick={() => { setTool('curl'); setCopiedCommand(false); }}
                  className={`px-3 py-1 text-xs font-medium rounded-md transition-all ${
                    tool === 'curl' ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700'
                  }`}
                >
                  cURL
                </button>
                <button
                  onClick={() => { setTool('wget'); setCopiedCommand(false); }}
                  className={`px-3 py-1 text-xs font-medium rounded-md transition-all ${
                    tool === 'wget' ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700'
                  }`}
                >
                  wget
                </button>
              </div>
            </div>
            
            <div className="relative group">
              <div className="bg-slate-900 text-slate-300 font-mono text-sm p-4 rounded-lg break-all border border-slate-800 leading-relaxed pr-12">
                {tool === 'curl' ? (
                  <>
                    <span className="text-yellow-400">curl</span> -sL {backendOrigin}/api/agent/install | <span className="text-yellow-400">sudo bash</span> -s -- -k <span className="text-green-400">{apiKey}</span>
                  </>
                ) : (
                  <>
                    <span className="text-yellow-400">wget</span> -qO- {backendOrigin}/api/agent/install | <span className="text-yellow-400">sudo bash</span> -s -- -k <span className="text-green-400">{apiKey}</span>
                  </>
                )}
              </div>
              <button
                onClick={handleCopyCommand}
                className="absolute top-2 right-2 p-2 bg-white/10 hover:bg-white/20 text-white rounded-md transition-colors backdrop-blur-sm"
                title="Copy Command"
              >
                {copiedCommand ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
              </button>
            </div>
            <p className="text-xs text-slate-500 mt-2">
              Run this on your server to download, configure, and start the agent automatically.
            </p>
          </div>

          {/* Divider */}
          <div className="border-t border-slate-100"></div>

          {/* Section 2: Manual Key */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <FileCode className="w-4 h-4 text-slate-500" />
              <label className="text-sm font-medium text-slate-700">
                Raw API Key
              </label>
            </div>
            <div className="flex items-center gap-2">
              <code className="flex-1 bg-slate-100 text-slate-700 font-mono text-sm p-3 rounded-lg break-all border border-slate-200">
                {apiKey}
              </code>
              <button
                onClick={handleCopyKey}
                className="p-3 rounded-lg border border-slate-200 hover:bg-slate-50 text-slate-600 transition-colors"
                title="Copy Key"
              >
                {copiedKey ? <Check className="w-5 h-5 text-green-600" /> : <Copy className="w-5 h-5" />}
              </button>
            </div>
            <p className="text-xs text-slate-500 mt-2">
              Use this if you are configuring the agent manually via <code>config.yaml</code>.
            </p>
          </div>

        </div>

        {/* Footer */}
        <div className="bg-slate-50 px-6 py-4 border-t border-slate-100 flex justify-end">
          <button
            onClick={onClose}
            className="px-6 py-2 bg-slate-900 hover:bg-slate-800 text-white text-sm font-medium rounded-lg transition-colors"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}