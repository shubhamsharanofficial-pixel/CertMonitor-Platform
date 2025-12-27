import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Download, Server, Terminal, Shield, ArrowLeft, Cpu, Loader2, FileText, FileCode } from 'lucide-react';
import { useAuth } from '../context';

export default function Downloads() {
  const { user } = useAuth();
  
  const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
  const backendOrigin = isLocal ? 'http://localhost:8080' : window.location.origin;

  const backLink = user ? "/dashboard" : "/";
  const backText = user ? "Back to Dashboard" : "Back to Home";

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <nav className="bg-white border-b border-slate-200 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
            <Link to={backLink} className="flex items-center gap-2 text-slate-600 hover:text-slate-900 transition-colors font-medium">
                <ArrowLeft className="w-4 h-4" /> {backText}
            </Link>
            <div className="flex items-center gap-2">
                <Shield className="w-5 h-5 text-blue-600" />
                <span className="font-bold text-slate-900">CertMonitor Agents</span>
            </div>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-6 py-12">
        <div className="text-center mb-16">
            <h1 className="text-3xl md:text-4xl font-bold text-slate-900 mb-4">Download Agent Binaries</h1>
            <p className="text-slate-500 text-lg max-w-2xl mx-auto">
                For manual installations, air-gapped systems, or automation scripts. 
                Simply download the binary, create a config file, and run.
            </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            {/* Linux AMD64 */}
            <DownloadCard 
                os="Linux" 
                arch="x86_64 / AMD64" 
                filename="agent-linux-amd64"
                url={`${backendOrigin}/api/downloads/agent-linux-amd64`}
                icon={Server}
            />

            {/* Linux ARM64 */}
            <DownloadCard 
                os="Linux" 
                arch="ARM64 (Raspberry Pi / AWS Graviton)" 
                filename="agent-linux-arm64"
                url={`${backendOrigin}/api/downloads/agent-linux-arm64`}
                icon={Cpu}
            />

            {/* macOS */}
            <DownloadCard 
                os="macOS" 
                arch="Apple Silicon (M1/M2/M3/M4)" 
                filename="agent-darwin-arm64"
                url={`${backendOrigin}/api/downloads/agent-darwin-arm64`}
                icon={Terminal}
            />

            {/* Config Template (Now fetches from Backend) */}
            <ConfigCard url={`${backendOrigin}/api/downloads/config.yaml`} />

        </div>

        {/* Quick Start Guide */}
        <div className="mt-16 bg-white rounded-xl border border-slate-200 p-8 shadow-sm">
            <h2 className="text-xl font-bold text-slate-900 mb-6 flex items-center gap-2">
                <Terminal className="w-5 h-5 text-slate-400" /> 
                Manual Installation Guide
            </h2>
            <div className="space-y-6">
                <Step number="1" title="Download & Permissions">
                    <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg font-mono text-sm overflow-x-auto mt-2">
                        wget {backendOrigin}/api/downloads/agent-linux-amd64 -O cert-agent{'\n'}
                        chmod +x cert-agent
                    </pre>
                </Step>
                <Step number="2" title="Create Config">
                    <p className="text-sm text-slate-600 mb-2">
                        Download the <span className="font-semibold text-slate-900">config.yaml</span> template above and add your API Key.
                    </p>
                </Step>
                <Step number="3" title="Run Agent">
                    <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg font-mono text-sm overflow-x-auto mt-2">
                        sudo ./cert-agent -config config.yaml
                    </pre>
                </Step>
            </div>
        </div>
      </main>
    </div>
  );
}

// 1. Binary Download Card
function DownloadCard({ os, arch, filename, url, icon: Icon }) {
    const [downloading, setDownloading] = useState(false);

    const handleDownload = async (e) => {
        e.preventDefault();
        setDownloading(true);

        try {
            const response = await fetch(url);
            
            if (!response.ok) {
                if (response.status === 404) {
                    alert(`Resource Not Found\n\nThe file '${filename}' is not available on the server yet.\nPlease check the backend 'public/downloads' directory.`);
                } else {
                    alert(`Download Failed\n\nServer returned status: ${response.status}`);
                }
                return;
            }

            // Convert to blob and trigger download
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = downloadUrl;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(downloadUrl);

        } catch (error) {
            console.error("Download error:", error);
            alert("Network Error\n\nUnable to reach the server.");
        } finally {
            setDownloading(false);
        }
    };

    return (
        <div className="bg-white p-6 rounded-xl border border-slate-200 hover:border-blue-300 hover:shadow-md transition-all group flex flex-col justify-between">
            <div>
                <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-slate-100 rounded-lg group-hover:bg-blue-50 transition-colors">
                            <Icon className="w-6 h-6 text-slate-600 group-hover:text-blue-600" />
                        </div>
                        <div>
                            <h3 className="font-bold text-slate-900">{os}</h3>
                            <p className="text-xs text-slate-500 font-medium">{arch}</p>
                        </div>
                    </div>
                </div>
                <div className="bg-slate-50 rounded-lg p-3 font-mono text-xs text-slate-600 mb-4 break-all border border-slate-100">
                    {filename}
                </div>
            </div>
            <button 
                onClick={handleDownload}
                disabled={downloading}
                className="flex items-center justify-center gap-2 w-full py-2.5 bg-slate-900 hover:bg-blue-600 disabled:bg-slate-400 text-white text-sm font-bold rounded-lg transition-colors cursor-pointer"
            >
                {downloading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                {downloading ? "Downloading..." : "Download Binary"}
            </button>
        </div>
    );
}

// 2. Config Download Card (Fetches from Backend)
function ConfigCard({ url }) {
    const [downloading, setDownloading] = useState(false);

    const handleDownloadConfig = async () => {
        setDownloading(true);
        try {
            const response = await fetch(url);
            if (!response.ok) {
                if (response.status === 404) {
                    alert(`Resource Not Found\n\nPlease ensure 'config.yaml' exists in the backend 'public/downloads' directory.`);
                } else {
                    alert(`Download Failed: ${response.status}`);
                }
                return;
            }
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = downloadUrl;
            link.download = "config.yaml";
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(downloadUrl);
        } catch (error) {
            alert("Network Error");
        } finally {
            setDownloading(false);
        }
    };

    return (
        <div className="bg-white p-6 rounded-xl border border-slate-200 hover:border-blue-300 hover:shadow-md transition-all group flex flex-col justify-between relative overflow-hidden">
            {/* Visual Flair */}
            <div className="absolute top-0 right-0 w-20 h-20 bg-yellow-50 rounded-bl-full -mr-4 -mt-4 opacity-50 pointer-events-none"></div>

            <div>
                <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-slate-100 rounded-lg group-hover:bg-yellow-50 transition-colors">
                            <FileCode className="w-6 h-6 text-slate-600 group-hover:text-yellow-600" />
                        </div>
                        <div>
                            <h3 className="font-bold text-slate-900">Config Template</h3>
                            <p className="text-xs text-slate-500 font-medium">YAML Configuration</p>
                        </div>
                    </div>
                </div>
                <div className="bg-slate-50 rounded-lg p-3 font-mono text-xs text-slate-600 mb-4 break-all border border-slate-100">
                    config.yaml
                </div>
                <p className="text-xs text-slate-500 mb-4 leading-relaxed">
                    A commented template file. <br/> Just add your API Key and paths.
                </p>
            </div>
            
            <button 
                onClick={handleDownloadConfig}
                disabled={downloading}
                className="flex items-center justify-center gap-2 w-full py-2.5 bg-white border-2 border-slate-200 hover:border-yellow-400 hover:bg-yellow-50 text-slate-700 hover:text-yellow-800 text-sm font-bold rounded-lg transition-all cursor-pointer"
            >
                {downloading ? <Loader2 className="w-4 h-4 animate-spin text-yellow-600" /> : <FileText className="w-4 h-4" />}
                {downloading ? "Fetching..." : "Download Template"}
            </button>
        </div>
    );
}

function Step({ number, title, children }) {
    return (
        <div className="flex gap-4">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-sm">
                {number}
            </div>
            <div className="flex-1">
                <h4 className="font-semibold text-slate-900">{title}</h4>
                <div className="mt-1">{children}</div>
            </div>
        </div>
    );
}