import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// --- HELPER METHODS ---

// Certificates
api.getCertificates = (params) => api.get(`/certs?${params.toString()}`);
api.deleteCertificate = (id) => api.delete(`/certs/${id}`);
api.pruneMissingCertificates = () => api.delete(`/certs/missing`);

// Agents
api.getAgents = () => api.get('/agents');
api.deleteAgent = (id) => api.delete(`/agents/${id}`);

// Cloud Monitor (Agentless)
api.getCloudTargets = () => api.get('/cloud/targets');

// Note: Backend expects 'frequency_hours' as an integer
api.addCloudTarget = (url, frequency) => api.post('/cloud/targets', { 
    url, 
    frequency_hours: parseInt(frequency) 
});

api.updateCloudTarget = (id, frequency) => api.put(`/cloud/targets/${id}`, { 
    frequency_hours: parseInt(frequency) 
});

api.deleteCloudTarget = (id) => api.delete(`/cloud/targets/${id}`);

// Auth - Core
api.login = (email, password) => api.post('/login', { email, password });
api.signup = (email, password, orgName) => api.post('/signup', { email, password, orgName });
api.regenerateKey = () => api.post('/key/regenerate');

// Auth - Account Recovery & Verification (NEW)
api.verifyEmail = (token) => api.post('/auth/verify', { token });
api.forgotPassword = (email) => api.post('/auth/forgot-password', { email });
api.resetPassword = (token, newPassword) => api.post('/auth/reset-password', { token, new_password: newPassword });

// Profile
api.getProfile = () => api.get('/profile');
api.updateProfile = (data) => api.put('/profile', data);

// Dashboard Stats
api.getStats = () => api.get('/stats');

export default api;