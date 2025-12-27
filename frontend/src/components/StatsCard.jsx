import React from 'react';
import { clsx } from 'clsx';

export default function StatsCard({ 
  title, 
  value, 
  icon: Icon, 
  color, 
  onClick, 
  isActive, 
  subtitle, 
  children 
}) {
  
  // Color Variants
  const colors = {
    blue: 'bg-blue-50 text-blue-600 border-blue-200',
    red: 'bg-red-50 text-red-600 border-red-200',
    yellow: 'bg-yellow-50 text-yellow-600 border-yellow-200',
    green: 'bg-green-50 text-green-600 border-green-200',
    gray: 'bg-slate-50 text-slate-600 border-slate-200',
  };

  const activeClass = isActive ? "ring-2 ring-offset-1 ring-blue-500" : "";
  const cursorClass = onClick ? "cursor-pointer hover:shadow-md transition-shadow" : "";

  return (
    <div 
      onClick={onClick}
      className={clsx(
        "bg-white p-6 rounded-xl border border-slate-200 flex flex-col justify-between h-full",
        cursorClass,
        activeClass
      )}
    >
      <div className="flex items-start justify-between mb-2">
        <div>
          <div className="flex items-baseline gap-2">
            <p className="text-sm font-medium text-slate-500">{title}</p>
            {subtitle && <span className="text-xs text-slate-400 font-normal">{subtitle}</span>}
          </div>
          <h3 className="text-3xl font-bold text-slate-900 mt-1">{value}</h3>
        </div>
        <div className={clsx("p-3 rounded-lg border", colors[color] || colors.gray)}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
      
      {/* Custom Footer Content (e.g. Agent breakdown) */}
      {children && (
        <div className="mt-2 pt-3 border-t border-slate-100 text-xs text-slate-500">
          {children}
        </div>
      )}
    </div>
  );
}