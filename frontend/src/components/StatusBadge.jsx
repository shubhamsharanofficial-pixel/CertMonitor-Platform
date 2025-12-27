import React from 'react';
import { CheckCircle, Clock, AlertTriangle, XCircle, ShieldAlert, Timer, Ghost } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export default function StatusBadge({ status, className }) {
  const styles = {
    'Valid': 'bg-green-100 text-green-800 border-green-200',
    
    // Urgency Levels
    'Expiring Soon': 'bg-yellow-50 text-yellow-700 border-yellow-200',        // Lowest Urgency
    'Expiring This Week': 'bg-yellow-100 text-yellow-800 border-yellow-300 font-semibold', // Medium Urgency
    'Expiring Tomorrow': 'bg-orange-100 text-orange-800 border-orange-200 font-bold',      // High Urgency
    'Expiring Today': 'bg-red-100 text-red-800 border-red-200 font-bold',                  // Critical Urgency
    
    'Expired': 'bg-red-100 text-red-800 border-red-200',
    'Not Yet Valid': 'bg-gray-100 text-gray-800 border-gray-200',
    'Untrusted': 'bg-red-50 text-red-700 border-red-200',
    
    // New Ghost Status (Optional, if passed explicitly)
    'MISSING': 'bg-slate-100 text-slate-500 border-slate-200 grayscale',
  };

  const icons = {
    'Valid': CheckCircle,
    'Expiring Soon': Clock,
    'Expiring This Week': Clock,
    'Expiring Tomorrow': Timer,
    'Expiring Today': AlertTriangle,
    'Expired': XCircle,
    'Not Yet Valid': Clock,
    'Untrusted': ShieldAlert,
    'MISSING': Ghost,
  };

  const Icon = icons[status] || CheckCircle;
  
  const baseClasses = "flex items-center w-fit px-2.5 py-0.5 rounded-full text-xs font-medium border";
  
  return (
    <span className={twMerge(clsx(baseClasses, styles[status] || 'bg-gray-100 text-gray-800', className))}>
      <Icon className="w-3 h-3 mr-1" />
      {status || "Unknown"}
    </span>
  );
}