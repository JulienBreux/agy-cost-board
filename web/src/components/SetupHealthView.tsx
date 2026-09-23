import React, { useState } from 'react';
import {
  DiagnosticReport,
  DiagnosticCheck,
} from '../api';
import {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  RefreshCw,
  Copy,
  Check,
  Terminal,
  Clock,
  Shield,
  FileCheck,
} from 'lucide-react';

interface SetupHealthViewProps {
  report: DiagnosticReport | null;
  loading: boolean;
  onRefresh: () => void;
}

export const SetupHealthView: React.FC<SetupHealthViewProps> = ({
  report,
  loading,
  onRefresh,
}) => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const handleCopy = (cmd: string, index: number) => {
    navigator.clipboard.writeText(cmd);
    setCopiedIndex(index);
    setTimeout(() => {
      setCopiedIndex(null);
    }, 2000);
  };

  const formatDuration = (nanos: number) => {
    if (nanos === 0) return '0ms';
    const ms = nanos / 1_000_000;
    if (ms < 1) return '<1ms';
    return `${ms.toFixed(1)}ms`;
  };

  const getStatusBadge = (status: DiagnosticCheck['status']) => {
    switch (status) {
      case 'OK':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-google-green/15 text-google-green border border-google-green/30">
            <CheckCircle2 className="h-3.5 w-3.5 mr-1" />
            PASSED
          </span>
        );
      case 'WARNING':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-google-yellow/15 text-google-yellow border border-google-yellow/30">
            <AlertTriangle className="h-3.5 w-3.5 mr-1" />
            WARNING
          </span>
        );
      case 'ERROR':
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-google-red/15 text-google-red border border-google-red/30">
            <XCircle className="h-3.5 w-3.5 mr-1" />
            ERROR
          </span>
        );
    }
  };

  if (!report && loading) {
    return (
      <div className="py-24 flex flex-col items-center justify-center space-y-3">
        <RefreshCw className="h-8 w-8 text-google-blue animate-spin" />
        <span className="text-sm text-google-gray-400 font-medium">Running GCP Telemetry & Diagnostics Probe...</span>
      </div>
    );
  }

  if (!report) {
    return (
      <div className="p-8 text-center text-google-gray-400">
        No diagnostic report available. Click "Re-run Verification" to perform health checks.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Top Banner & Refresh */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-[#14171f] p-5 rounded-xl border border-[#202530]">
        <div>
          <div className="flex items-center space-x-3">
            <h2 className="text-lg font-bold text-white tracking-tight">GCP Prerequisites & Telemetry Health</h2>
            <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-[#1e2330] text-google-gray-300 border border-[#2d3548]">
              {report.project_id || 'Active GCP Project'}
            </span>
          </div>
          <p className="text-xs text-google-gray-400 mt-1">
            Automated verification of ADC credentials, IAM permissions, BigQuery logging sink, and Cloud Billing export table.
          </p>
        </div>
        <button
          onClick={onRefresh}
          disabled={loading}
          className="inline-flex items-center justify-center space-x-2 px-4 py-2 rounded-lg text-xs font-semibold bg-google-blue hover:bg-google-blue/90 text-white transition-colors disabled:opacity-50"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>{loading ? 'Verifying...' : 'Re-run Verification'}</span>
        </button>
      </div>

      {/* Summary Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Overall Status */}
        <div className="bg-[#14171f] p-5 rounded-xl border border-[#202530] flex items-center justify-between">
          <div>
            <div className="text-xs font-medium text-google-gray-400 uppercase tracking-wider">Overall Status</div>
            <div className="text-xl font-bold mt-1 text-white">
              {report.overall_status === 'OK' && <span className="text-google-green">Healthy</span>}
              {report.overall_status === 'WARNING' && <span className="text-google-yellow">Attention Needed</span>}
              {report.overall_status === 'ERROR' && <span className="text-google-red">Setup Incomplete</span>}
            </div>
          </div>
          <div className="p-3 rounded-lg bg-[#1a1e28]">
            <Shield className="h-6 w-6 text-google-blue" />
          </div>
        </div>

        {/* Passed Checks */}
        <div className="bg-[#14171f] p-5 rounded-xl border border-[#202530] flex items-center justify-between">
          <div>
            <div className="text-xs font-medium text-google-gray-400 uppercase tracking-wider">Passed Checks</div>
            <div className="text-xl font-bold text-google-green mt-1">{report.passed_count} / {report.checks.length}</div>
          </div>
          <div className="p-3 rounded-lg bg-google-green/10">
            <CheckCircle2 className="h-6 w-6 text-google-green" />
          </div>
        </div>

        {/* Warnings */}
        <div className="bg-[#14171f] p-5 rounded-xl border border-[#202530] flex items-center justify-between">
          <div>
            <div className="text-xs font-medium text-google-gray-400 uppercase tracking-wider">Warnings</div>
            <div className="text-xl font-bold text-google-yellow mt-1">{report.warning_count}</div>
          </div>
          <div className="p-3 rounded-lg bg-google-yellow/10">
            <AlertTriangle className="h-6 w-6 text-google-yellow" />
          </div>
        </div>

        {/* Errors */}
        <div className="bg-[#14171f] p-5 rounded-xl border border-[#202530] flex items-center justify-between">
          <div>
            <div className="text-xs font-medium text-google-gray-400 uppercase tracking-wider">Errors</div>
            <div className="text-xl font-bold text-google-red mt-1">{report.error_count}</div>
          </div>
          <div className="p-3 rounded-lg bg-google-red/10">
            <XCircle className="h-6 w-6 text-google-red" />
          </div>
        </div>
      </div>

      {/* Diagnostics Check List */}
      <div className="bg-[#14171f] rounded-xl border border-[#202530] overflow-hidden">
        <div className="p-4 border-b border-[#202530] flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <FileCheck className="h-4 w-4 text-google-blue" />
            <h3 className="text-sm font-bold text-white uppercase tracking-wider">Verification Pipeline Checks</h3>
          </div>
          <span className="text-xs text-google-gray-400">
            {new Date(report.timestamp).toLocaleTimeString()}
          </span>
        </div>

        <div className="divide-y divide-[#1e2330]">
          {report.checks.map((check, idx) => (
            <div key={check.id} className="p-5 hover:bg-[#161a23] transition-colors space-y-3">
              {/* Check Header */}
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                <div className="flex items-center space-x-3">
                  <div className="text-white font-semibold text-sm sm:text-base flex items-center space-x-2">
                    <span>{check.name}</span>
                  </div>
                  <span className="text-[11px] font-mono text-google-gray-400 px-2 py-0.5 rounded bg-[#1e2330] border border-[#282f40]">
                    {check.id}
                  </span>
                </div>
                <div className="flex items-center space-x-3">
                  <span className="inline-flex items-center text-xs text-google-gray-400 space-x-1">
                    <Clock className="h-3.5 w-3.5 mr-1" />
                    {formatDuration(check.duration)}
                  </span>
                  {getStatusBadge(check.status)}
                </div>
              </div>

              {/* Message */}
              <p className="text-xs sm:text-sm text-google-gray-300">
                {check.message}
              </p>

              {/* Details if present */}
              {check.details && Object.keys(check.details).length > 0 && (
                <div className="bg-[#0f1115] p-3 rounded-lg border border-[#202530] text-xs font-mono text-google-gray-400 space-y-1">
                  {Object.entries(check.details).map(([k, v]) => (
                    <div key={k} className="flex items-start space-x-2">
                      <span className="text-google-blue font-semibold">{k}:</span>
                      <span className="text-google-gray-200">{Array.isArray(v) ? v.join(', ') : String(v)}</span>
                    </div>
                  ))}
                </div>
              )}

              {/* Remediation Command snippet if present */}
              {check.remediation_command && (
                <div className="mt-2 bg-[#0c0e12] rounded-lg border border-[#282f40] p-3">
                  <div className="flex items-center justify-between text-[11px] text-google-gray-400 mb-1">
                    <span className="flex items-center space-x-1 font-semibold text-google-yellow">
                      <Terminal className="h-3.5 w-3.5 mr-1" />
                      Remediation Action:
                    </span>
                    <button
                      onClick={() => handleCopy(check.remediation_command!, idx)}
                      className="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-[#1a1e28] hover:bg-[#202530] text-google-gray-300 hover:text-white transition-colors"
                      title="Copy command to clipboard"
                    >
                      {copiedIndex === idx ? (
                        <>
                          <Check className="h-3 w-3 text-google-green" />
                          <span className="text-[10px] text-google-green font-semibold">Copied!</span>
                        </>
                      ) : (
                        <>
                          <Copy className="h-3 w-3" />
                          <span className="text-[10px]">Copy</span>
                        </>
                      )}
                    </button>
                  </div>
                  <pre className="text-xs font-mono text-google-gray-200 overflow-x-auto whitespace-pre-wrap select-all py-1">
                    {check.remediation_command}
                  </pre>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
