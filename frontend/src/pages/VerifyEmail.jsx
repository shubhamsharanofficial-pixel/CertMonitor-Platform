import React, { useEffect, useState, useRef } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { CheckCircle, XCircle, Loader2, Shield } from 'lucide-react';
import { useAuth } from '../context';

export default function VerifyEmail() {
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token');
    const { verifyEmail } = useAuth();
    
    const [status, setStatus] = useState('verifying'); // verifying | success | error
    const [errorMsg, setErrorMsg] = useState('');
    
    // Fix for React Strict Mode (Prevent double-firing)
    const verificationAttempted = useRef(false);

    useEffect(() => {
        if (!token) {
            setStatus('error');
            setErrorMsg('Invalid verification link.');
            return;
        }

        // If we already tried verifying this session, stop.
        if (verificationAttempted.current) return;
        verificationAttempted.current = true;

        const runVerification = async () => {
            try {
                await verifyEmail(token);
                setStatus('success');
            } catch (err) {
                setStatus('error');
                setErrorMsg(err.response?.data || 'Verification failed. The link may have expired.');
            }
        };

        runVerification();
    }, [token, verifyEmail]);

    return (
        <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
            <div className="max-w-md w-full bg-white rounded-xl shadow-lg border border-slate-100 overflow-hidden p-8 text-center">
                
                {status === 'verifying' && (
                    <div className="flex flex-col items-center">
                        <Loader2 className="w-12 h-12 text-blue-600 animate-spin mb-4" />
                        <h2 className="text-xl font-bold text-slate-900">Verifying Email...</h2>
                        <p className="text-slate-500 mt-2">Please wait while we activate your account.</p>
                    </div>
                )}

                {status === 'success' && (
                    <div className="flex flex-col items-center animate-in zoom-in duration-300">
                        <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
                            <CheckCircle className="w-8 h-8 text-green-600" />
                        </div>
                        <h2 className="text-2xl font-bold text-slate-900">Email Verified!</h2>
                        <p className="text-slate-500 mt-2 mb-6">Your account has been successfully activated.</p>
                        
                        <Link to="/login" className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 rounded-lg transition-colors block">
                            Continue to Login
                        </Link>
                    </div>
                )}

                {status === 'error' && (
                    <div className="flex flex-col items-center animate-in zoom-in duration-300">
                        <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
                            <XCircle className="w-8 h-8 text-red-600" />
                        </div>
                        <h2 className="text-xl font-bold text-slate-900">Verification Failed</h2>
                        <p className="text-slate-500 mt-2 mb-6">{errorMsg}</p>
                        
                        <Link to="/login" className="text-blue-600 font-semibold hover:text-blue-700">
                            Back to Login
                        </Link>
                    </div>
                )}
            </div>
        </div>
    );
}