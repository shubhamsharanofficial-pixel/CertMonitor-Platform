import React, { useState, useEffect, useCallback } from 'react';
import { Globe, Plus, Trash2, Search, RefreshCw, AlertCircle, CheckCircle, XCircle, Clock, Edit, Save, X } from 'lucide-react';
import api from '../services/api';
import { useAuth } from '../context';
import Navbar from '../components/Navbar';
import StatsCard from '../components/StatsCard';

export default function CloudMonitor() {
  const { user, logout } = useAuth();
  
  // Data State
  const [targets, setTargets] = useState([]);
  const [loading, setLoading] = useState(true);
  
  // UI State
  const [showAddForm, setShowAddForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [search, setSearch] = useState('');
  
  // Edit State
  const [editingTarget, setEditingTarget] = useState(null);
  const [editFrequency, setEditFrequency] = useState(12);

  // Add Form Inputs
  const [newUrl, setNewUrl] = useState('');
  const [frequency, setFrequency] = useState(12);

  // 1. Fetch Data
  const fetchTargets = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.getCloudTargets();
      setTargets(res.data || []);
      setError(null);
    } catch (err) {
      console.error(err);
      setError("Failed to load targets");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTargets();
  }, [fetchTargets]);

  // 2. Handlers
  const handleAdd = async (e) => {
    e.preventDefault();
    if (!newUrl) return;
    
    setSubmitting(true);
    setError(null);
    try {
      const cleanUrl = newUrl.replace(/^https?:\/\//, '');
      await api.addCloudTarget(cleanUrl, frequency);
      
      setNewUrl('');
      setShowAddForm(false);
      
      // Refresh Data (Source of Truth)
      fetchTargets();
    } catch (err) {
      setError(err.response?.data || "Failed to add target");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id, url) => {
    if (!window.confirm(`Stop monitoring ${url}?`)) return;
    try {
      await api.deleteCloudTarget(id);
      
      // Refresh Data (Source of Truth)
      fetchTargets();
    } catch (err) {
      alert("Failed to delete target");
    }
  };

  const openEditModal = (target) => {
    setEditingTarget(target);
    setEditFrequency(target.frequency_hours);
    setError(null);
  };

  const handleUpdate = async () => {
    setSubmitting(true);
    try {
      await api.updateCloudTarget(editingTarget.id, editFrequency);
      
      // Refresh Data (Source of Truth)
      fetchTargets(); 
      
      setEditingTarget(null); // Close Modal
    } catch (err) {
      alert("Failed to update frequency");
    } finally {
      setSubmitting(false);
    }
  };

  // 3. Filtering
  const filteredTargets = targets.filter(t => 
    t.target_url.toLowerCase().includes(search.toLowerCase())
  );

  // Stats
  const total = targets.length;
  const healthy = targets.filter(t => t.last_status === 'SUCCESS').length;
  const failing = targets.filter(t => t.last_status === 'FAILED').length;

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <Navbar user={user} logout={logout} />

      <main className="max-w-7xl mx-auto px-6 py-8">
        
        {/* Header & Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <StatsCard title="Monitored Sites" value={total} icon={Globe} color="blue" />
          <StatsCard title="Healthy" value={healthy} icon={CheckCircle} color="green" />
          <StatsCard title="Unreachable" value={failing} icon={XCircle} color="red" />
        </div>

        {/* Controls */}
        <div className="flex flex-col md:flex-row justify-between items-center mb-6 gap-4">
          <div className="relative w-full md:w-96">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
            <input 
              type="text"
              placeholder="Search domains..."
              className="w-full pl-9 pr-4 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          
          <button 
            onClick={() => setShowAddForm(true)} 
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg flex items-center gap-2 shadow-sm transition-colors"
          >
            <Plus className="w-4 h-4" /> Add Website
          </button>
        </div>

        {/* Add Form */}
        {showAddForm && (
          <div className="mb-6 bg-white p-6 rounded-xl border border-blue-100 shadow-sm animate-in slide-in-from-top-2">
            <h3 className="font-bold text-slate-800 mb-4">Monitor New Website</h3>
            <form onSubmit={handleAdd} className="flex flex-col md:flex-row gap-4 items-end">
              <div className="flex-1 w-full">
                <label className="block text-xs font-semibold text-slate-500 uppercase mb-1">Domain / IP</label>
                <input 
                  type="text" 
                  placeholder="e.g. google.com or 192.168.1.5:8443" 
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono text-sm"
                  value={newUrl}
                  onChange={(e) => setNewUrl(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="w-full md:w-48">
                <label className="block text-xs font-semibold text-slate-500 uppercase mb-1">Frequency</label>
                <select 
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm bg-white"
                  value={frequency}
                  onChange={(e) => setFrequency(e.target.value)}
                >
                  <option value="1">Every Hour</option>
                  <option value="6">Every 6 Hours</option>
                  <option value="12">Every 12 Hours</option>
                  <option value="24">Daily (24h)</option>
                </select>
              </div>
              <div className="flex gap-2 w-full md:w-auto">
                <button 
                  type="button" 
                  onClick={() => setShowAddForm(false)}
                  className="px-4 py-2 border border-slate-300 rounded-lg text-slate-600 hover:bg-slate-50 text-sm font-medium"
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  disabled={submitting}
                  className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium flex items-center justify-center gap-2 min-w-[100px]"
                >
                  {submitting ? <RefreshCw className="w-4 h-4 animate-spin"/> : "Start Scan"}
                </button>
              </div>
            </form>
            {error && <p className="mt-3 text-sm text-red-600 flex items-center gap-2"><AlertCircle className="w-4 h-4"/> {error}</p>}
          </div>
        )}

        {/* Table */}
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          {loading ? (
            <div className="p-12 text-center text-slate-400">Loading monitors...</div>
          ) : filteredTargets.length === 0 ? (
            <div className="p-12 text-center text-slate-500 flex flex-col items-center">
              <Globe className="w-12 h-12 text-slate-300 mb-4" />
              <p className="text-lg font-medium">No Websites Monitored</p>
              <p className="text-sm text-slate-400">Add a domain to start tracking its SSL certificates.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4 font-semibold text-slate-700">Target URL</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Scanner Status</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Frequency</th>
                    <th className="px-6 py-4 font-semibold text-slate-700">Last Scan</th>
                    <th className="px-6 py-4 font-semibold text-slate-700 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {filteredTargets.map((t) => (
                    <tr key={t.id} className="hover:bg-slate-50">
                      <td className="px-6 py-4 font-medium text-slate-900 font-mono">{t.target_url}</td>
                      <td className="px-6 py-4">
                        <StatusPill status={t.last_status} error={t.last_error} />
                      </td>
                      <td className="px-6 py-4 text-slate-500">
                        Every {t.frequency_hours}h
                      </td>
                      <td className="px-6 py-4 text-slate-500">
                         <div className="flex items-center gap-2">
                            <Clock className="w-3.5 h-3.5 text-slate-400" />
                            {/* Check for zero date from Go backend */}
                            {t.last_scanned_at && t.last_scanned_at !== "0001-01-01T00:00:00Z" 
                              ? new Date(t.last_scanned_at).toLocaleString() 
                              : 'Pending...'}
                         </div>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button 
                            onClick={() => openEditModal(t)}
                            className="p-2 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
                            title="Edit Frequency"
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          
                          <button 
                            onClick={() => handleDelete(t.id, t.target_url)}
                            className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-md transition-colors"
                            title="Stop Monitoring"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>

      {/* Edit Modal */}
      {editingTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4 animate-in fade-in duration-200">
            <div className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="bg-slate-50 px-6 py-4 border-b border-slate-100 flex justify-between items-center">
                    <h3 className="font-bold text-slate-800">Edit Frequency</h3>
                    <button onClick={() => setEditingTarget(null)} className="text-slate-400 hover:text-slate-600">
                        <X className="w-5 h-5" />
                    </button>
                </div>
                
                <div className="p-6">
                    <div className="mb-4">
                        <label className="block text-xs font-semibold text-slate-500 uppercase mb-1">Target URL</label>
                        <div className="font-mono text-sm text-slate-900 bg-slate-50 p-2 rounded border border-slate-200">
                            {editingTarget.target_url}
                        </div>
                    </div>
                    
                    <div className="mb-6">
                        <label className="block text-xs font-semibold text-slate-500 uppercase mb-1">Scan Frequency</label>
                        <select 
                          className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-sm bg-white"
                          value={editFrequency}
                          onChange={(e) => setEditFrequency(e.target.value)}
                        >
                          <option value="1">Every Hour</option>
                          <option value="6">Every 6 Hours</option>
                          <option value="12">Every 12 Hours</option>
                          <option value="24">Daily (24h)</option>
                        </select>
                    </div>

                    <div className="flex justify-end gap-3">
                        <button 
                            onClick={() => setEditingTarget(null)}
                            className="px-4 py-2 border border-slate-300 rounded-lg text-slate-600 hover:bg-slate-50 text-sm font-medium"
                        >
                            Cancel
                        </button>
                        <button 
                            onClick={handleUpdate}
                            disabled={submitting}
                            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium flex items-center gap-2"
                        >
                            {submitting ? <RefreshCw className="w-4 h-4 animate-spin"/> : <Save className="w-4 h-4" />}
                            Save Changes
                        </button>
                    </div>
                </div>
            </div>
        </div>
      )}
    </div>
  );
}

function StatusPill({ status, error }) {
  if (status === 'SUCCESS') {
    return (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
        <CheckCircle className="w-3 h-3" /> Healthy
      </span>
    );
  }
  if (status === 'FAILED') {
    return (
      <div className="group relative inline-block">
        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 border border-red-200 cursor-help">
          <XCircle className="w-3 h-3" /> Unreachable
        </span>
        {error && (
          <div className="absolute left-0 bottom-full mb-2 w-64 p-2 bg-slate-900 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10 shadow-lg">
            {error}
          </div>
        )}
      </div>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-600 border border-blue-100 animate-pulse">
      <RefreshCw className="w-3 h-3 animate-spin" /> Scanning...
    </span>
  );
}