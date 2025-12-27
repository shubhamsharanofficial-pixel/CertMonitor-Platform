import React, { useState, useEffect, useCallback } from 'react';
import { Server, Trash2, Clock, Activity, AlertCircle, Search, Filter, Signal, SignalLow, SignalHigh, RefreshCw } from 'lucide-react';
import api from '../services/api';
import { useAuth } from '../context';
import Navbar from '../components/Navbar';
import StatsCard from '../components/StatsCard';

// UPDATED: Synchronized with Dashboard to 30 seconds
const POLL_INTERVAL = 30000; 

export default function Agents() {
  const { user, logout } = useAuth();
  
  // Data State
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState(null);

  // Filter State
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL'); // 'ALL' | 'ONLINE' | 'OFFLINE'

  // Fetch Logic - Modified for Polling
  const fetchAgents = useCallback(async (isBackground = false) => {
    if (!isBackground) setLoading(true);
    else setIsRefreshing(true);

    try {
      const response = await api.getAgents();
      setAgents(response.data || []);
      setError(null);
    } catch (err) {
      console.error(err);
      if (!isBackground) setError("Failed to load agents.");
    } finally {
      if (!isBackground) setLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  // Initial Load
  useEffect(() => {
    fetchAgents(false);
  }, [fetchAgents]);

  // Polling Effect
  useEffect(() => {
    const intervalId = setInterval(() => {
        fetchAgents(true);
    }, POLL_INTERVAL);

    return () => clearInterval(intervalId);
  }, [fetchAgents]);

  // Delete Logic
  const handleDelete = async (id, hostname) => {
    if (!window.confirm(`Are you sure you want to delete agent "${hostname}"? \n\nThis will also delete all certificates associated with it.`)) {
      return;
    }

    try {
      await api.deleteAgent(id);
      // Optimistic Update
      setAgents(prev => prev.filter(a => a.id !== id));
    } catch (err) {
      alert("Failed to delete agent: " + (err.response?.data || err.message));
    }
  };

  // --- CLIENT-SIDE FILTERING LOGIC ---
  const filteredAgents = agents.filter(agent => {
    const matchesSearch = agent.hostname.toLowerCase().includes(search.toLowerCase()) || 
                          (agent.ip_address && agent.ip_address.includes(search));
    
    const matchesStatus = statusFilter === 'ALL' || 
                          (statusFilter === 'ONLINE' && agent.status === 'Online') ||
                          (statusFilter === 'OFFLINE' && agent.status === 'Offline');

    return matchesSearch && matchesStatus;
  });

  // Calculate Stats on the fly
  const totalCount = agents.length;
  const onlineCount = agents.filter(a => a.status === 'Online').length;
  const offlineCount = totalCount - onlineCount;

  // Helper for Relative Time
  const timeAgo = (dateString) => {
    const diff = Date.now() - new Date(dateString).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes} mins ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} hours ago`;
    return `${Math.floor(hours / 24)} days ago`;
  };

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <Navbar user={user} logout={logout} />

      <main className="max-w-7xl mx-auto px-6 py-8">
        
        {/* 1. Stats Row */}
        {!loading && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
            <StatsCard 
              title="Total Agents" 
              value={totalCount} 
              icon={Server} 
              color="blue" 
              onClick={() => setStatusFilter('ALL')}
              isActive={statusFilter === 'ALL'}
            />
            <StatsCard 
              title="Online" 
              value={onlineCount} 
              icon={SignalHigh} 
              color="green" 
              onClick={() => setStatusFilter('ONLINE')}
              isActive={statusFilter === 'ONLINE'}
            />
            <StatsCard 
              title="Offline" 
              value={offlineCount} 
              icon={SignalLow} 
              color="gray" 
              onClick={() => setStatusFilter('OFFLINE')}
              isActive={statusFilter === 'OFFLINE'}
            />
          </div>
        )}

        {/* 2. Controls Row */}
        <div className="flex flex-col md:flex-row justify-between items-center mb-6 gap-4">
          <div className="flex items-center gap-4 w-full md:w-auto">
            {/* Search */}
            <div className="relative w-full md:w-64">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
              <input 
                type="text"
                placeholder="Search hostname or IP..."
                className="w-full pl-9 pr-4 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>

            {/* Status Filter */}
            <div className="relative">
              <Filter className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
              <select 
                className="pl-9 pr-8 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 bg-white appearance-none min-w-[140px]"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="ALL">All Status</option>
                <option value="ONLINE">Online Only</option>
                <option value="OFFLINE">Offline Only</option>
              </select>
            </div>
          </div>

          <button onClick={() => fetchAgents(false)} className="px-4 py-2 bg-white border border-slate-300 rounded-lg text-sm font-medium text-slate-700 hover:bg-slate-50 flex items-center gap-2">
            <RefreshCw className={`w-4 h-4 ${loading || isRefreshing ? 'animate-spin' : ''}`} />
            Refresh List
          </button>
        </div>

        {error && (
           <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6 flex items-center gap-2">
             <AlertCircle className="w-5 h-5" />
             {error}
           </div>
        )}

        {/* Table */}
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          {loading ? (
            <div className="p-12 text-center text-slate-400">Loading inventory...</div>
          ) : filteredAgents.length === 0 ? (
            <div className="p-12 text-center text-slate-500 flex flex-col items-center">
              <Server className="w-12 h-12 text-slate-300 mb-4" />
              <p className="text-lg font-medium">No Agents Found</p>
              <p className="text-sm">
                {agents.length === 0 ? "Install the agent on a server to see it here." : "No agents match your filters."}
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4 font-semibold text-slate-700">Status</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Hostname</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">IP Address</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Certificates</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Last Seen</th>
                    <th className="px-6 py-4 font-semibold text-slate-700 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {filteredAgents.map((agent) => (
                    <tr key={agent.id} className="hover:bg-slate-50">
                      <td className="px-6 py-4">
                        <AgentStatusBadge status={agent.status} />
                      </td>
                      <td className="px-6 py-4 font-medium text-slate-900">{agent.hostname}</td>
                      <td className="px-6 py-4 text-slate-500 font-mono text-xs">{agent.ip_address || "N/A"}</td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-1.5">
                            <Activity className="w-4 h-4 text-slate-400" />
                            {agent.cert_count}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-slate-500">
                        <div className="flex items-center gap-1.5" title={new Date(agent.last_seen_at).toLocaleString()}>
                            <Clock className="w-4 h-4 text-slate-400" />
                            {timeAgo(agent.last_seen_at)}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <button 
                          onClick={() => handleDelete(agent.id, agent.hostname)}
                          className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-md transition-colors"
                          title="Delete Agent"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

// Internal Helper Component for Agent Status
function AgentStatusBadge({ status }) {
  const isOnline = status === 'Online';
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
      isOnline 
        ? 'bg-green-100 text-green-800 border border-green-200' 
        : 'bg-gray-100 text-gray-800 border border-gray-200'
    }`}>
      <span className={`w-1.5 h-1.5 mr-1.5 rounded-full ${
        isOnline ? 'bg-green-500' : 'bg-gray-500'
      }`}></span>
      {status}
    </span>
  );
}