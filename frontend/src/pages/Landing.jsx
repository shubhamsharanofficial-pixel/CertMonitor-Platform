import React from 'react';
import { Link } from 'react-router-dom';
import { Shield, Server, Zap, ArrowRight, Download, Terminal, Lock, Activity, CheckCircle } from 'lucide-react';

export default function Landing() {
  return (
    <div className="min-h-screen bg-white font-sans text-slate-900">
      
      {/* 1. Navigation */}
      <nav className="border-b border-slate-100 bg-white/80 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-blue-600 p-1.5 rounded-lg">
                <Shield className="text-white w-5 h-5" />
            </div>
            <span className="text-lg font-bold tracking-tight">CertMonitor</span>
          </div>
          <div className="flex items-center gap-6">
            <Link to="/downloads" className="text-sm font-medium text-slate-600 hover:text-slate-900 hidden md:block">Downloads</Link>
            <Link to="/login" className="text-sm font-medium text-slate-600 hover:text-slate-900">Log In</Link>
            <Link to="/signup" className="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white text-sm font-medium rounded-lg transition-colors">
                Get Started
            </Link>
          </div>
        </div>
      </nav>

      {/* 2. Hero Section */}
      <section className="pt-20 pb-32 px-6 relative overflow-hidden">
        <div className="max-w-7xl mx-auto text-center relative z-10">
            <div className="inline-flex items-center gap-3 px-4 py-1.5 rounded-full bg-blue-50 text-blue-700 text-sm md:text-base font-bold mb-8 border border-blue-100 shadow-sm">
                <span className="relative flex h-3 w-3">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
                </span>
                Now with Ghost Detection
            </div>
            <h1 className="text-5xl md:text-6xl font-extrabold tracking-tight text-slate-900 mb-6 max-w-4xl mx-auto leading-tight">
                Never Let an SSL Certificate <br className="hidden md:block" />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-indigo-600">Expire Again.</span>
            </h1>
            <p className="text-xl text-slate-500 mb-10 max-w-2xl mx-auto leading-relaxed">
                Automated discovery, centralized monitoring, and instant alerts for your entire infrastructure. Run a simple agent, and we handle the rest.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                <Link to="/signup" className="w-full sm:w-auto px-8 py-4 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl shadow-lg hover:shadow-blue-500/20 transition-all flex items-center justify-center gap-2">
                    Create Free Account <ArrowRight className="w-5 h-5" />
                </Link>
                <Link to="/downloads" className="w-full sm:w-auto px-8 py-4 bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 font-bold rounded-xl transition-all flex items-center justify-center gap-2">
                    <Download className="w-5 h-5" /> Download Agent
                </Link>
            </div>
        </div>
        
        {/* Background Decoration */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full max-w-7xl pointer-events-none opacity-30">
            <div className="absolute top-20 left-10 w-72 h-72 bg-blue-400 rounded-full mix-blend-multiply filter blur-3xl animate-blob"></div>
            <div className="absolute top-20 right-10 w-72 h-72 bg-purple-400 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-2000"></div>
        </div>
      </section>

      {/* 3. Features Grid */}
      <section className="py-24 bg-slate-50 border-t border-slate-200">
        <div className="max-w-7xl mx-auto px-6">
            <div className="text-center mb-16">
                <h2 className="text-3xl font-bold text-slate-900">Why DevOps Teams Choose Us</h2>
                <p className="text-slate-500 mt-4 text-lg">Monitoring shouldn't be complicated. We keep it simple and effective.</p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <FeatureCard 
                    icon={Terminal} 
                    title="One-Line Install" 
                    desc="Deploy our lightweight Go binary in seconds. Works on Linux, macOS, and standard Docker containers."
                />
                <FeatureCard 
                    icon={Zap} 
                    title="Smart Discovery" 
                    desc="We automatically find certificates in file paths (/etc/ssl) and scan listening ports (443, 8443)."
                />
                <FeatureCard 
                    icon={Lock} 
                    title="Private & Secure" 
                    desc="Agent-based architecture means you never open firewall ports. Your private keys never leave your server."
                />
            </div>
        </div>
      </section>

      {/* 4. Demo Showcase (NEW SECTION) */}
      <section className="py-24 bg-white">
        <div className="max-w-7xl mx-auto px-6">
            <div className="text-center mb-16">
                <h2 className="text-3xl font-bold text-slate-900">See It In Action</h2>
                <p className="text-slate-500 mt-4 text-lg">A single pane of glass for all your certificates.</p>
            </div>

            {/* Dashboard Screenshot */}
            <div className="relative rounded-xl bg-slate-900 p-2 shadow-2xl border border-slate-800 mb-20 transform hover:scale-[1.01] transition-transform duration-500">
                {/* Browser Window Chrome */}
                <div className="flex items-center gap-2 px-2 py-2 mb-2">
                    <div className="w-3 h-3 rounded-full bg-red-500"></div>
                    <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                    <div className="w-3 h-3 rounded-full bg-green-500"></div>
                    <div className="ml-4 px-4 py-1 bg-slate-800 rounded text-xs text-slate-400 font-mono w-full max-w-sm">certmonitor.io/dashboard</div>
                </div>
                {/* Image */}
                <img 
                    src="/demo/dashboard.jpeg" 
                    alt="CertMonitor Dashboard" 
                    className="rounded-lg w-full h-auto border border-slate-700/50"
                />
            </div>

            {/* Feature Split Layout */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
                <div className="order-2 lg:order-1">
                    <h3 className="text-2xl font-bold text-slate-900 mb-4">Detailed Agent Inventory</h3>
                    <p className="text-slate-600 leading-relaxed mb-6">
                        Know exactly which server is reporting which certificate. 
                        Track agent health, last seen status, and network configuration 
                        from a centralized view.
                    </p>
                    <ul className="space-y-3">
                        <li className="flex items-center gap-3 text-slate-700">
                            <CheckCircle className="w-5 h-5 text-green-500" />
                            <span>Track Online/Offline status</span>
                        </li>
                        <li className="flex items-center gap-3 text-slate-700">
                            <CheckCircle className="w-5 h-5 text-green-500" />
                            <span>View IP addresses and Hostnames</span>
                        </li>
                        <li className="flex items-center gap-3 text-slate-700">
                            <CheckCircle className="w-5 h-5 text-green-500" />
                            <span>Drill down into individual certs</span>
                        </li>
                    </ul>
                </div>
                <div className="order-1 lg:order-2 relative rounded-xl bg-slate-900 p-2 shadow-xl border border-slate-800 transform rotate-1 hover:rotate-0 transition-transform duration-500">
                     {/* Browser Window Chrome */}
                    <div className="flex items-center gap-2 px-2 py-2 mb-2">
                        <div className="w-3 h-3 rounded-full bg-red-500"></div>
                        <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                        <div className="w-3 h-3 rounded-full bg-green-500"></div>
                    </div>
                    <img 
                        src="/demo/agents.jpeg" 
                        alt="Agent Inventory" 
                        className="rounded-lg w-full h-auto border border-slate-700/50"
                    />
                </div>
            </div>

        </div>
      </section>

      {/* 5. Footer */}
      <footer className="bg-slate-50 border-t border-slate-200 py-12">
        <div className="max-w-7xl mx-auto px-6 flex flex-col md:flex-row justify-between items-center gap-6">
            <div className="flex items-center gap-2">
                <div className="bg-slate-900 p-1.5 rounded-lg">
                    <Shield className="text-white w-4 h-4" />
                </div>
                <span className="font-bold text-slate-900">CertMonitor</span>
            </div>
            <div className="flex gap-8 text-sm text-slate-500 font-medium">
                <Link to="/downloads" className="hover:text-blue-600 transition-colors">Downloads</Link>
                <Link to="/login" className="hover:text-blue-600 transition-colors">Login</Link>
                <Link to="/signup" className="hover:text-blue-600 transition-colors">Signup</Link>
            </div>
            <p className="text-slate-400 text-sm">© 2025 CertMonitor. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon: Icon, title, desc }) {
    return (
        <div className="bg-white p-8 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-shadow">
            <div className="w-12 h-12 bg-blue-50 rounded-xl flex items-center justify-center mb-6">
                <Icon className="w-6 h-6 text-blue-600" />
            </div>
            <h3 className="text-xl font-bold text-slate-900 mb-3">{title}</h3>
            <p className="text-slate-500 leading-relaxed">{desc}</p>
        </div>
    );
}