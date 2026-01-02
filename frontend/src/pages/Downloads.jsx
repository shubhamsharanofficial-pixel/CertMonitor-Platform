import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Download, Server, Terminal, Shield, ArrowLeft, Cpu, Loader2, FileText, FileCode, Container } from 'lucide-react';
import { useAuth } from '../context';

export default function Downloads() {
  const { user } = useAuth();
  const [guideTab, setGuideTab] = useState('binary'); // 'binary' | 'docker'
  
  // FIXED: In Docker/Nginx setup, the API is always accessible relative to the current domain/port.
  const backendOrigin = window.location.origin; 

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
                For manual installations, air-gapped systems, or containerized environments.
            </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            
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
                arch="ARM64 (Pi / Graviton)" 
                filename="agent-linux-arm64"
                url={`${backendOrigin}/api/downloads/agent-linux-arm64`}
                icon={Cpu}
            />

            {/* macOS */}
            <DownloadCard 
                os="macOS" 
                arch="Apple Silicon (M1+)" 
                filename="agent-darwin-arm64"
                url={`${backendOrigin}/api/downloads/agent-darwin-arm64`}
                icon={Terminal}
            />

            {/* Config Template */}
            <ResourceCard 
                title="Config Template"
                subtitle="Required for all setups"
                filename="config.yaml"
                url={`${backendOrigin}/api/downloads/config.yaml`}
                icon={FileCode}
                color="yellow"
            />

            {/* Docker Compose */}
            <ResourceCard 
                title="Docker Compose"
                subtitle="Containerized Setup"
                filename="docker-compose.yml"
                url={`${backendOrigin}/api/downloads/docker-compose.agent.yml`}
                icon={Container}
                color="blue"
            />

        </div>

        {/* Installation Guide */}
        <div className="mt-16 bg-white rounded-xl border border-slate-200 p-8 shadow-sm">
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-slate-900 flex items-center gap-2">
                    <Terminal className="w-5 h-5 text-slate-400" /> 
                    Installation Guide
                </h2>
                <div className="flex bg-slate-100 p-1 rounded-lg">
                    <button onClick={() => setGuideTab('binary')} className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all ${guideTab === 'binary' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}>Binary</button>
                    <button onClick={() => setGuideTab('docker')} className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all ${guideTab === 'docker' ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700'}`}>Docker</button>
                </div>
            </div>

            {guideTab === 'binary' ? (
                <div className="space-y-6 animate-in fade-in duration-300">
                    <Step number="1" title="Download & Permissions">
                        <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg font-mono text-sm overflow-x-auto mt-2">
                            wget {backendOrigin}/api/downloads/agent-linux-amd64 -O cert-agent{'\n'}
                            chmod +x cert-agent
                        </pre>
                    </Step>
                    <Step number="2" title="Create Config">
                        <p className="text-sm text-slate-600 mb-2">Download <span className="font-semibold text-slate-900">config.yaml</span> and add your API Key.</p>
                    </Step>
                    <Step number="3" title="Run Agent">
                        <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg font-mono text-sm overflow-x-auto mt-2">
                            sudo ./cert-agent -config config.yaml
                        </pre>
                    </Step>
                </div>
            ) : (
                <div className="space-y-6 animate-in fade-in duration-300">
                    <Step number="1" title="Download Files">
                        <p className="text-sm text-slate-600 mb-2">Download both <span className="font-semibold text-slate-900">docker-compose.yml</span> and <span className="font-semibold text-slate-900">config.yaml</span>.</p>
                    </Step>
                    <Step number="2" title="Update Docker Compose">
                        <p className="text-sm text-slate-600 mb-2">Edit <span className="font-semibold text-slate-900">docker-compose.yml</span> to map the host directories you want to scan (volumes).</p>
                        <p className="text-xs text-slate-500">Example: Map <code>/etc/ssl/certs</code> (Host) to <code>/app/scan_target</code> (Container).</p>
                    </Step>
                    <Step number="3" title="Update Config">
                        <p className="text-sm text-slate-600 mb-2">Edit <code>config.yaml</code> with your API Key and Backend URL.</p>
                        <p className="text-sm text-slate-600 mb-2 font-medium">Important: Use paths internal to the container:</p>
                        <pre className="bg-slate-50 text-slate-700 p-4 rounded-lg font-mono text-sm border border-slate-200">
state_path: "/app/data"
cert_paths:
  - "/app/scan_target"
                        </pre>
                    </Step>
                    <Step number="4" title="Run Container">
                        <pre className="bg-slate-900 text-slate-300 p-4 rounded-lg font-mono text-sm overflow-x-auto mt-2">
                            docker compose up -d
                        </pre>
                    </Step>
                </div>
            )}
        </div>
      </main>
    </div>
  );
}

// Binary Download Card
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
        <div className="bg-white p-6 rounded-xl border border-slate-200 hover:border-slate-300 hover:shadow-md transition-all group flex flex-col justify-between">
            <div>
                <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-slate-100 rounded-lg group-hover:bg-slate-200 transition-colors">
                            <Icon className="w-6 h-6 text-slate-600 group-hover:text-slate-800" />
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
                className="flex items-center justify-center gap-2 w-full py-2.5 bg-slate-900 hover:bg-slate-800 text-white text-sm font-bold rounded-lg transition-colors cursor-pointer"
            >
                {downloading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                {downloading ? "Downloading..." : "Download"}
            </button>
        </div>
    );
}

function ResourceCard({ title, subtitle, filename, url, icon: Icon, color }) {
    const [downloading, setDownloading] = useState(false);

    const colors = {
        yellow: 'hover:border-yellow-400 group-hover:bg-yellow-50 text-yellow-600',
        blue: 'hover:border-blue-400 group-hover:bg-blue-50 text-blue-600',
    };
    
    const iconColor = colors[color] || colors.blue;

    const handleDownload = async () => {
        setDownloading(true);
        try {
            const response = await fetch(url);
            if (!response.ok) {
                if (response.status === 404) alert(`Resource Not Found: ${filename}`);
                else alert(`Download Failed: ${response.status}`);
                return;
            }
            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = downloadUrl;
            link.download = filename;
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
        <div className={`bg-white p-6 rounded-xl border border-slate-200 hover:shadow-md transition-all group flex flex-col justify-between ${color === 'yellow' ? 'hover:border-yellow-300' : 'hover:border-blue-300'}`}>
            <div>
                <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <div className={`p-2 bg-slate-100 rounded-lg transition-colors ${iconColor.replace('text-', 'group-hover:bg-')}`}>
                            <Icon className={`w-6 h-6 text-slate-600 ${iconColor}`} />
                        </div>
                        <div>
                            <h3 className="font-bold text-slate-900">{title}</h3>
                            <p className="text-xs text-slate-500 font-medium">{subtitle}</p>
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
                className="flex items-center justify-center gap-2 w-full py-2.5 bg-white border-2 border-slate-200 text-slate-700 text-sm font-bold rounded-lg transition-all cursor-pointer hover:bg-slate-50"
            >
                {downloading ? <Loader2 className="w-4 h-4 animate-spin" /> : <FileText className="w-4 h-4" />}
                {downloading ? "Fetching..." : "Download"}
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