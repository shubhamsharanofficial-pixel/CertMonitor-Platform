import React, { useState, useEffect, useCallback } from 'react';
import { 
  Server, RefreshCw, ChevronLeft, ChevronRight, Search, Filter, 
  AlertCircle, AlertTriangle, ShieldCheck, ShieldAlert,
  SignalHigh, SignalLow, ChevronDown, ChevronUp, Copy, Check, 
  Globe, Cloud, FileText, Ghost, Key, Trash2, X, Plus
} from 'lucide-react';
import { Link } from 'react-router-dom';
import api from '../services/api';
import { useAuth } from '../context'; 
import StatusBadge from '../components/StatusBadge';
import ApiKeyModal from '../components/ApiKeyModal';
import Navbar from '../components/Navbar';
import StatsCard from '../components/StatsCard';

// POLLING INTERVAL (in milliseconds)
const POLL_INTERVAL = 30000; // 30 seconds

export default function Dashboard() {
  const { user, logout, generateApiKey } = useAuth();
  
  // Data State
  const [certs, setCerts] = useState([]);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState(null);
  const [agents, setAgents] = useState([]);
  
  // UI State
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [expandedRowId, setExpandedRowId] = useState(null);
  const [copiedId, setCopiedId] = useState(null);

  // SEARCH STATE
  const [localSearch, setLocalSearch] = useState(''); 

  // Filter State
  const [filters, setFilters] = useState({
    page: 1,
    limit: 10,
    search: '',
    agentId: '',
    status: 'ALL',
    startDate: '',
    endDate: '',
    trust: 'ALL',
    presence: 'ALL'
  });

  // Helper to check if filters are active
  const hasActiveFilters = 
    filters.search !== '' || 
    filters.agentId !== '' || 
    filters.status !== 'ALL' || 
    filters.startDate !== '' || 
    filters.endDate !== '' || 
    filters.trust !== 'ALL' || 
    filters.presence !== 'ALL';

  // --- API CALLS ---

  const fetchMetadata = useCallback(async () => {
    try {
      const [statsRes, agentsRes] = await Promise.all([
        api.getStats(),
        api.getAgents()
      ]);
      setStats(statsRes.data);
      setAgents(agentsRes.data || []);
    } catch (err) {
      console.error("Failed to load dashboard metadata", err);
    }
  }, []);

  const fetchCerts = useCallback(async (isBackground = false) => {
    if (!isBackground) setLoading(true);
    else setIsRefreshing(true); 

    if (!isBackground) setExpandedRowId(null);

    try {
      const params = new URLSearchParams({
        page: filters.page.toString(),
        limit: filters.limit.toString(),
      });

      if (filters.search) params.append('search', filters.search);
      if (filters.agentId) params.append('agent_id', filters.agentId);

      if (filters.trust === 'TRUSTED') params.append('trusted', 'true');
      if (filters.trust === 'UNTRUSTED') params.append('trusted', 'false');
      if (filters.presence !== 'ALL') params.append('status', filters.presence);

      const now = new Date();
      if (filters.status === 'EXPIRING') {
        const nextMonth = new Date();
        nextMonth.setDate(now.getDate() + 30);
        params.append('valid_after', now.toISOString());
        params.append('valid_before', nextMonth.toISOString());
       
      } else if (filters.status === 'EXPIRED') {
        params.append('valid_before', now.toISOString());
      } else {
        if (filters.startDate) params.append('valid_after', filters.startDate);
        if (filters.endDate) params.append('valid_before', filters.endDate);
      }

      const response = await api.getCertificates(params);
      
      setCerts(response.data.data || []);
      setTotal(response.data.total || 0);
      setError(null);
    } catch (err) {
      console.error(err);
      if (!isBackground) setError("Failed to fetch data.");
    } finally {
      if (!isBackground) setLoading(false);
      setIsRefreshing(false);
    }
  }, [filters]); 

  // Initial Load
  useEffect(() => { 
      fetchMetadata(); 
      fetchCerts(false); 
  }, [fetchMetadata, fetchCerts]);

  // --- POLLING STRATEGY ---
  useEffect(() => {
    const intervalId = setInterval(() => {
        fetchCerts(true);
        fetchMetadata();
    }, POLL_INTERVAL);

    return () => clearInterval(intervalId);
  }, [fetchCerts, fetchMetadata]);


  // --- HANDLERS ---
  const handleManualRefresh = () => { fetchCerts(false); fetchMetadata(); };
  
  const triggerSearch = () => {
    setFilters(prev => ({ ...prev, search: localSearch, page: 1 }));
  };

  const handleSearchKeyDown = (e) => {
    if (e.key === 'Enter') { triggerSearch(); }
  };
  
  const handleFilterChange = (key, value) => { setFilters(prev => ({ ...prev, [key]: value, page: 1 })); };
  
  const handleQuickFilter = (status) => { 
      setLocalSearch(''); 
      setFilters(prev => ({ 
          ...prev, 
          search: '',
          status: status, 
          startDate: '', 
          endDate: '', 
          page: 1,
          trust: 'ALL',
          presence: 'ALL'
      })); 
  };
  
  const handleDateChange = (key, value) => { setFilters(prev => ({ ...prev, [key]: value, status: 'ALL', page: 1 })); };

  const handleResetFilters = () => {
      setLocalSearch('');
      setFilters({
        page: 1, limit: 10, search: '', agentId: '', status: 'ALL', startDate: '', endDate: '', trust: 'ALL', presence: 'ALL'
      });
  };

  const handleGenerateKey = async () => {
    if (user.has_api_key && !window.confirm("WARNING: Generating a new key will invalidate your old one. Continue?")) return;
    setIsGenerating(true);
    try {
      const key = await generateApiKey();
      setNewKey(key);
      setIsModalOpen(true);
      fetchMetadata();
    } catch (err) { alert("Failed to generate key"); } finally { setIsGenerating(false); }
  };

  const handleDeleteInstance = async (e, id, name) => {
    e.stopPropagation(); 
    if (!window.confirm(`Delete certificate record for "${name}"?\n\nThis is a permanent action.`)) return;
    try { await api.deleteCertificate(id); fetchCerts(true); fetchMetadata(); } 
    catch(err) { alert("Failed to delete certificate"); }
  };

  const handlePruneMissing = async () => {
    if (!window.confirm("Are you sure you want to delete ALL 'Missing' certificates?\n\nThis action cannot be undone.")) return;
    try {
        const res = await api.pruneMissingCertificates();
        alert(`Successfully deleted ${res.data.count} missing certificates.`);
        fetchCerts(false);
        fetchMetadata();
    } catch (err) { alert("Failed to prune missing certificates"); }
  };

  const toggleRow = (id) => { setExpandedRowId(expandedRowId === id ? null : id); };
  
  const handleCopy = (e, text, id) => {
    e.stopPropagation();
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const getDateClass = (status) => {
    if (status === 'Expired') return 'text-red-600 font-bold';
    if (['Expiring Today', 'Expiring Tomorrow'].includes(status)) return 'text-orange-600 font-bold';
    if (status === 'Expiring This Week') return 'text-yellow-600 font-bold';
    if (status === 'Expiring Soon') return 'text-yellow-600 font-medium';
    return '';
  };

  const renderSourceIcon = (type) => {
    switch (type) {
      case 'CLOUD': return <Cloud className="w-4 h-4 text-sky-500" />;
      case 'NETWORK': return <Globe className="w-4 h-4 text-blue-500" />;
      default: return <FileText className="w-4 h-4 text-slate-400" />;
    }
  };

  const getDisplaySource = (cert) => {
    if (!cert.source_uid) return '-';
    if (cert.source_type === 'FILE') {
      return cert.source_uid.split(/[/\\]/).pop();
    }
    return cert.source_uid;
  };

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <Navbar user={user} logout={logout} onRotate={handleGenerateKey} isGenerating={isGenerating} />
      <ApiKeyModal isOpen={isModalOpen} onClose={() => { setIsModalOpen(false); setNewKey(''); }} apiKey={newKey} />
      
      <main className="max-w-7xl mx-auto px-6 py-8">

        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
            <StatsCard title="Total Certificates" value={stats.total_certs} icon={ShieldCheck} color="blue" onClick={() => handleQuickFilter('ALL')} isActive={filters.status === 'ALL' && !filters.startDate} />
            <StatsCard title="Expiring Soon" subtitle="(Next 30 Days)" value={stats.expiring_soon} icon={AlertTriangle} color="yellow" onClick={() => handleQuickFilter('EXPIRING')} isActive={filters.status === 'EXPIRING'} />
            <StatsCard title="Expired" value={stats.expired} icon={AlertCircle} color="red" onClick={() => handleQuickFilter('EXPIRED')} isActive={filters.status === 'EXPIRED'} />
            <StatsCard title="Total Agents" value={stats.total_agents} icon={Server} color="gray">
              <div className="flex gap-4">
                <div className="flex items-center gap-1.5 text-green-600 font-medium"><SignalHigh className="w-3 h-3" />{stats.online_agents} Online</div>
                <div className="flex items-center gap-1.5 text-slate-400"><SignalLow className="w-3 h-3" />{stats.offline_agents} Offline</div>
              </div>
            </StatsCard>
          </div>
        )}

        <div className="bg-white p-4 rounded-xl border border-slate-200 mb-6 space-y-4 shadow-sm">
            <div className="flex flex-col md:flex-row gap-4 justify-between">
                <div className="flex flex-col md:flex-row gap-4 flex-1">
                    <div className="flex gap-2 flex-1 max-w-md">
                        <div className="relative flex-1">
                            <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                            <input type="text" placeholder="Search certificates..." className="w-full pl-9 pr-4 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" value={localSearch} onChange={(e) => setLocalSearch(e.target.value)} onKeyDown={handleSearchKeyDown} />
                        </div>
                        <button onClick={triggerSearch} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">Search</button>
                    </div>
                    
                    <div className="flex gap-2 flex-wrap">
                        <div className="relative">
                            <Filter className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
                            <select className="pl-9 pr-8 py-2 border border-slate-300 rounded-lg text-sm bg-white cursor-pointer focus:ring-2 focus:ring-blue-500 outline-none" value={filters.agentId} onChange={(e) => handleFilterChange('agentId', e.target.value)}>
                                <option value="">All Agents</option>
                                {agents.map(a => <option key={a.id} value={a.id}>{a.hostname}</option>)}
                            </select>
                        </div>
                        <select className="px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white cursor-pointer focus:ring-2 focus:ring-blue-500 outline-none" value={filters.trust} onChange={(e) => handleFilterChange('trust', e.target.value)}>
                            <option value="ALL">All Trust</option>
                            <option value="TRUSTED">Trusted Only</option>
                            <option value="UNTRUSTED">Untrusted Only</option>
                        </select>
                        <select className="px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white cursor-pointer focus:ring-2 focus:ring-blue-500 outline-none" value={filters.presence} onChange={(e) => handleFilterChange('presence', e.target.value)}>
                            <option value="ALL">All Visibility</option>
                            <option value="ACTIVE">Active</option>
                            <option value="MISSING">Missing (Ghosts)</option>
                        </select>
                    </div>
                </div>
                <div className="flex items-center gap-2 border-t md:border-t-0 pt-4 md:pt-0 border-slate-100">
                    {filters.presence === 'MISSING' && (
                        <button onClick={handlePruneMissing} className="flex items-center gap-2 px-3 py-2 bg-red-50 text-red-600 hover:bg-red-100 border border-red-200 rounded-lg text-sm font-medium transition-colors" title="Delete all certificates marked as Missing">
                            <Trash2 className="w-4 h-4" /> <span className="hidden lg:inline">Delete Missing</span>
                        </button>
                    )}

                    {/* --- NEW: Persistent Connect Agent Button (Only visible if no key exists) --- */}
                    {!user?.has_api_key && (
                        <button 
                            onClick={handleGenerateKey}
                            disabled={isGenerating}
                            className="flex items-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors shadow-sm animate-in fade-in"
                            title="Generate API Key to connect physical agents"
                        >
                            {isGenerating ? <RefreshCw className="w-4 h-4 animate-spin"/> : <Key className="w-4 h-4" />}
                            <span className="hidden lg:inline">Connect Agent</span>
                        </button>
                    )}
                    {/* -------------------------------------------------------------------------- */}

                    <div className="w-px h-8 bg-slate-200 mx-2 hidden md:block"></div>
                    {hasActiveFilters && (
                        <button onClick={handleResetFilters} className="flex items-center gap-2 px-3 py-2 bg-slate-50 border border-slate-300 rounded-lg hover:bg-slate-100 hover:text-slate-700 transition-colors text-slate-500 text-sm font-medium" title="Clear All Filters"><X className="w-4 h-4" /> Reset</button>
                    )}
                    <button onClick={handleManualRefresh} className="p-2 bg-slate-50 border border-slate-300 rounded-lg hover:bg-white hover:text-blue-600 transition-colors text-slate-500" title="Refresh Data"><RefreshCw className={`w-4 h-4 ${loading || isRefreshing ? 'animate-spin' : ''}`} /></button>
                </div>
            </div>
            {(filters.startDate || filters.endDate || filters.status === 'ALL') && (
                <div className="flex items-center gap-2 pt-2 border-t border-slate-100">
                    <span className="text-xs font-semibold text-slate-500 uppercase">Expiry Range:</span>
                    <input type="date" className="px-2 py-1 border border-slate-300 rounded text-sm text-slate-600 outline-none focus:border-blue-500" value={filters.startDate} onChange={(e) => handleDateChange('startDate', e.target.value)} />
                    <span className="text-slate-300">-</span>
                    <input type="date" className="px-2 py-1 border border-slate-300 rounded text-sm text-slate-600 outline-none focus:border-blue-500" value={filters.endDate} onChange={(e) => handleDateChange('endDate', e.target.value)} />
                </div>
            )}
        </div>

        {error && <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">{error}</div>}

        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          {loading ? (
             <div className="p-12 text-center text-slate-400">Loading...</div> 
          ) : (
            <>
            {/* Empty State Logic */}
            {certs.length === 0 && !hasActiveFilters ? (
                 <div className="p-12 text-center animate-in fade-in duration-300">
                    <div className="mx-auto bg-slate-50 w-20 h-20 rounded-full flex items-center justify-center mb-6">
                        <ShieldCheck className="w-10 h-10 text-slate-400" />
                    </div>
                    <h3 className="text-xl font-bold text-slate-900 mb-2">No Certificates Monitored Yet</h3>
                    <p className="text-slate-500 max-w-md mx-auto mb-8">
                        Get started by adding a public website or connecting a physical server agent.
                    </p>
                    
                    <div className="flex flex-col sm:flex-row gap-4 justify-center">
                        {/* Option 1: Cloud Monitor */}
                        <Link to="/cloud" className="flex items-center gap-3 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-xl shadow hover:shadow-lg transition-all text-left">
                            <div className="p-2 bg-blue-500/20 rounded-lg"><Cloud className="w-5 h-5 text-white" /></div>
                            <div>
                                <div className="font-bold text-sm">Monitor Public Website</div>
                                <div className="text-xs text-blue-100">Scan google.com, etc.</div>
                            </div>
                        </Link>

                        {/* Option 2: Physical Agent */}
                        <button onClick={handleGenerateKey} className="flex items-center gap-3 px-6 py-3 bg-white border border-slate-200 hover:border-blue-300 hover:bg-slate-50 text-slate-700 rounded-xl shadow-sm hover:shadow transition-all text-left">
                            <div className="p-2 bg-slate-100 rounded-lg"><Server className="w-5 h-5 text-slate-600" /></div>
                            <div>
                                <div className="font-bold text-sm">Connect Local Agent</div>
                                <div className="text-xs text-slate-400">Generate API Key</div>
                            </div>
                        </button>
                    </div>
                 </div>
            ) : (
                <>
                  <div className="overflow-x-auto">
                    <table className="w-full text-left text-sm whitespace-nowrap">
                        <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th className="px-4 py-4 w-8"></th>
                            <th className="px-6 py-4 font-semibold text-slate-700 w-10 text-center">Trust</th>
                            <th className="px-6 py-4 font-semibold text-slate-700">Status</th>
                            <th className="px-6 py-4 font-semibold text-slate-700">Agent / Hostname</th>
                            <th className="px-6 py-4 font-semibold text-slate-700">Source Identifier</th>
                            <th className="px-6 py-4 font-semibold text-slate-700">Expires</th>
                            <th className="px-4 py-4 w-10"></th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                        {certs.length === 0 ? (
                            <tr><td colSpan="7" className="px-6 py-12 text-center text-slate-500">No certificates found matching your filters.</td></tr>
                        ) : (
                            certs.map((cert) => (
                            <React.Fragment key={cert.id}>
                                <tr onClick={() => toggleRow(cert.id)} className={`hover:bg-slate-50 transition-colors cursor-pointer ${expandedRowId === cert.id ? 'bg-slate-50' : ''} ${cert.current_status === 'MISSING' ? 'opacity-70 grayscale-[0.5]' : ''}`}>
                                    <td className="px-4 py-4 text-slate-400">{expandedRowId === cert.id ? <ChevronUp className="w-4 h-4"/> : <ChevronDown className="w-4 h-4"/>}</td>
                                    <td className="px-6 py-4 text-center">{cert.is_trusted ? <ShieldCheck className="w-5 h-5 text-green-500 mx-auto" /> : <ShieldAlert className="w-5 h-5 text-red-500 mx-auto" />}</td>
                                    <td className="px-6 py-4"><div className="flex flex-col gap-1 items-start"><StatusBadge status={cert.status} />{cert.current_status === 'MISSING' && <StatusBadge status="MISSING" />}</div></td>
                                    <td className="px-6 py-4 font-medium text-slate-900"><div className="flex items-center gap-2"><div title={`Source Type: ${cert.source_type}`}>{renderSourceIcon(cert.source_type)}</div>{cert.agent_hostname}{cert.current_status === 'MISSING' && (<div className="p-1 rounded bg-slate-100 border border-slate-200 text-slate-400" title="Missing from agent (Ghost)"><Ghost className="w-3.5 h-3.5" /></div>)}</div></td>
                                    <td className="px-6 py-4 max-w-[200px] truncate font-mono text-xs text-slate-600" title={cert.source_uid}>{getDisplaySource(cert)}</td>
                                    <td className={`px-6 py-4 font-mono ${getDateClass(cert.status)}`}>{new Date(cert.valid_until).toLocaleDateString()}</td>
                                    <td className="px-4 py-4 text-right">{cert.current_status === 'MISSING' && (<button onClick={(e) => handleDeleteInstance(e, cert.id, cert.subject.cn)} className="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors" title="Delete Missing Certificate"><Trash2 className="w-4 h-4" /></button>)}</td>
                                </tr>
                                {expandedRowId === cert.id && (
                                <tr className="bg-slate-50 border-b border-slate-100 relative"><td colSpan="7" className="px-4 py-4 sm:px-12 pb-6 relative"><div className="space-y-4 animate-in slide-in-from-top-1 duration-200">
                                        {!cert.is_trusted && (<div className="bg-red-50 border border-red-200 rounded-lg p-3 flex items-start gap-3"><ShieldAlert className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" /><div><h4 className="text-sm font-semibold text-red-800">Certificate Trust Error</h4><p className="text-sm text-red-700 mt-1">{cert.trust_error}</p></div></div>)}
                                        <div><div className="flex justify-between items-end mb-1"><label className="text-xs font-semibold text-slate-500 uppercase tracking-wide">Source Identifier</label><span className="text-xs font-medium text-slate-400 bg-slate-200 px-2 py-0.5 rounded">Type: {cert.source_type || 'FILE'}</span></div><div className="flex items-center gap-2"><div className="flex-1 flex items-center gap-3 bg-white border border-slate-200 p-3 rounded-md text-sm font-mono text-slate-700 break-all">{renderSourceIcon(cert.source_type)}{cert.source_uid}</div><button onClick={(e) => handleCopy(e, cert.source_uid, cert.id)} className="p-2 bg-white border border-slate-200 rounded-md hover:bg-slate-100 text-slate-500 transition-colors" title="Copy Source">{copiedId === cert.id ? <Check className="w-4 h-4 text-green-600" /> : <Copy className="w-4 h-4" />}</button></div></div>
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4"><div><label className="text-xs font-semibold text-slate-500 uppercase tracking-wide">Issuer Details</label><div className="mt-1 text-sm text-slate-700 bg-white border border-slate-200 p-3 rounded-md"><div className="grid grid-cols-[60px_1fr] gap-2"><span className="text-slate-400">CN:</span> <span className="font-medium">{cert.issuer.cn}</span><span className="text-slate-400">Org:</span> <span>{cert.issuer.org || 'N/A'}</span><span className="text-slate-400">OU:</span> <span>{cert.issuer.ou || 'N/A'}</span></div></div></div><div><label className="text-xs font-semibold text-slate-500 uppercase tracking-wide">Subject Details</label><div className="mt-1 text-sm text-slate-700 bg-white border border-slate-200 p-3 rounded-md"><div className="grid grid-cols-[60px_1fr] gap-2"><span className="text-slate-400">CN:</span> <span className="font-medium">{cert.subject.cn}</span><span className="text-slate-400">Org:</span> <span>{cert.subject.org || 'N/A'}</span><span className="text-slate-400">OU:</span> <span>{cert.subject.ou || 'N/A'}</span></div></div></div></div>
                                        <div className="flex justify-center pt-2"><button onClick={(e) => { e.stopPropagation(); toggleRow(cert.id); }} className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-500 hover:text-slate-700 hover:bg-slate-200/50 rounded-full transition-colors"><ChevronUp className="w-4 h-4" />Collapse Details</button></div></div></td></tr>
                                )}
                            </React.Fragment>
                            ))
                        )}
                        </tbody>
                    </table>
                  </div>
                  <div className="bg-slate-50 px-6 py-4 border-t border-slate-200 flex items-center justify-between"><span className="text-sm text-slate-500">Page {filters.page} of {Math.ceil(total / filters.limit) || 1}</span><div className="flex gap-2"><button disabled={filters.page === 1} onClick={() => setFilters(p => ({...p, page: p.page - 1}))} className="px-3 py-1 border rounded bg-white disabled:opacity-50 hover:bg-slate-50 text-slate-600"><ChevronLeft className="w-4 h-4"/></button><button disabled={filters.page * filters.limit >= total} onClick={() => setFilters(p => ({...p, page: p.page + 1}))} className="px-3 py-1 border rounded bg-white disabled:opacity-50 hover:bg-slate-50 text-slate-600"><ChevronRight className="w-4 h-4"/></button></div></div>
                </>
            )}
            </>
          )}
        </div>
      </main>
    </div>
  );
}