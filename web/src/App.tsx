import { useState, useEffect, useMemo } from 'react';
import {
  fetchOverview,
  fetchAttributedCosts,
  fetchLicenseGovernance,
  fetchUserSummary,
  fetchSetupStatus,
  OverviewMetrics,
  AllocatedUserCost,
  LicenseGovernance,
  UserSummary,
  DiagnosticReport,
} from './api';
import { Navbar } from './components/Navbar';
import { KPICards } from './components/KPICards';
import { CostChart } from './components/CostChart';
import { CostTable } from './components/CostTable';
import { LicenseTable } from './components/LicenseTable';
import { SetupHealthView } from './components/SetupHealthView';
import { UserModal } from './components/UserModal';
import { UserDashboardView } from './components/UserDashboardView';
import { RefreshCw, AlertTriangle } from 'lucide-react';

export function App() {
  const [activeTab, setActiveTab] = useState<
    'overview' | 'costs' | 'licenses' | 'setup' | 'my-consumption'
  >(() => {
    const params = new URLSearchParams(window.location.search);
    const tabParam = params.get('tab');
    if (
      tabParam === 'my-consumption' ||
      tabParam === 'costs' ||
      tabParam === 'licenses' ||
      tabParam === 'setup'
    ) {
      return tabParam;
    }
    return 'overview';
  });
  const [dashboardUserId, setDashboardUserId] = useState<string>(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get('user') || '';
  });
  const [days, setDays] = useState<number>(30);

  const [overview, setOverview] = useState<OverviewMetrics | null>(null);
  const [costs, setCosts] = useState<AllocatedUserCost[]>([]);
  const [governance, setGovernance] = useState<LicenseGovernance | null>(null);
  const [setupReport, setSetupReport] = useState<DiagnosticReport | null>(null);
  const [setupLoading, setSetupLoading] = useState<boolean>(false);

  const [selectedUser, setSelectedUser] = useState<UserSummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [ov, c, gov] = await Promise.all([
        fetchOverview(days),
        fetchAttributedCosts(days),
        fetchLicenseGovernance(days),
      ]);
      setOverview(ov);
      setCosts(c);
      setGovernance(gov);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred while communicating with the backend.');
      }
    } finally {
      setLoading(false);
    }
  };

  const loadSetup = async () => {
    setSetupLoading(true);
    try {
      const rep = await fetchSetupStatus();
      setSetupReport(rep);
    } catch (err: unknown) {
      console.error('Failed to load setup diagnostics:', err);
    } finally {
      setSetupLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    loadSetup();
  }, [days]);

  useEffect(() => {
    if (activeTab === 'setup' && !setupReport) {
      loadSetup();
    }
  }, [activeTab]);

  // Synchronize URL query params (?tab=...&user=...)
  useEffect(() => {
    try {
      const params = new URLSearchParams(window.location.search);
      if (activeTab === 'overview') {
        params.delete('tab');
      } else {
        params.set('tab', activeTab);
      }
      if (activeTab === 'my-consumption' && dashboardUserId) {
        params.set('user', dashboardUserId);
      } else {
        params.delete('user');
      }
      const newQuery = params.toString();
      const newUrl = window.location.pathname + (newQuery ? '?' + newQuery : '');
      window.history.replaceState(null, '', newUrl);
    } catch {
      // Ignore URL sync errors in environments without history API
    }
  }, [activeTab, dashboardUserId]);

  const availableUsers = useMemo(() => {
    const set = new Set<string>();
    costs.forEach((c) => {
      if (c.user_id) set.add(c.user_id);
    });
    governance?.dormant_users?.forEach((u) => {
      if (u.user_id) set.add(u.user_id);
    });
    return Array.from(set).sort();
  }, [costs, governance]);

  const handleSelectUser = async (userId: string) => {
    try {
      const summary = await fetchUserSummary(userId, days);
      setSelectedUser(summary);
    } catch (err) {
      console.error('Failed to load user summary:', err);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-[#0f1115] text-google-gray-200">
      {/* Navigation Header */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        days={days}
        setDays={setDays}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Error Alert */}
        {error && (
          <div className="p-4 bg-google-red/10 border border-google-red/30 rounded-xl flex items-center space-x-3 text-google-red">
            <AlertTriangle className="h-5 w-5 flex-shrink-0" />
            <div className="text-xs sm:text-sm font-medium">{error}</div>
            <button
              onClick={loadData}
              className="ml-auto underline text-xs font-semibold hover:text-white"
            >
              Retry
            </button>
          </div>
        )}

        {/* Loading Spinner for attribution data */}
        {activeTab !== 'setup' && activeTab !== 'my-consumption' && loading && !overview && (
          <div className="py-24 flex flex-col items-center justify-center space-y-3">
            <RefreshCw className="h-8 w-8 text-google-blue animate-spin" />
            <span className="text-sm text-google-gray-400 font-medium">Reconciling BigQuery Telemetry & Billing...</span>
          </div>
        )}

        {/* Tab: My Personal Consumption & Driving */}
        {activeTab === 'my-consumption' && (
          <UserDashboardView
            days={days}
            availableUsers={availableUsers}
            initialUserId={dashboardUserId}
            onUserChange={setDashboardUserId}
          />
        )}

        {/* Setup & Health Tab */}
        {activeTab === 'setup' && (
          <SetupHealthView
            report={setupReport}
            loading={setupLoading}
            onRefresh={loadSetup}
          />
        )}

        {/* Dashboard Content (Overview, Costs, Licenses) */}
        {activeTab !== 'setup' && activeTab !== 'my-consumption' && overview && (
          <>
            {/* Top KPIs */}
            <KPICards metrics={overview} />

            {/* Tab: Overview */}
            {activeTab === 'overview' && (
              <div className="space-y-6">
                <CostChart
                  trends={overview.daily_trends}
                  modelBreakdown={overview.model_breakdown}
                  currency={overview.currency}
                />
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <h2 className="text-sm font-bold text-white uppercase tracking-wider">
                      Recent Proportional Cost Allocations
                    </h2>
                    <button
                      onClick={() => setActiveTab('costs')}
                      className="text-xs font-semibold text-google-blue hover:underline"
                    >
                      View all costs →
                    </button>
                  </div>
                  <CostTable costs={costs} onSelectUser={handleSelectUser} />
                </div>
              </div>
            )}

            {/* Tab: Costs */}
            {activeTab === 'costs' && (
              <div className="space-y-4">
                <div>
                  <h2 className="text-lg font-bold text-white tracking-tight">Per-User Cost Attribution</h2>
                  <p className="text-xs text-google-gray-400">
                    Proportional cost breakdown derived from BigQuery token counts reconciled with Google Cloud Billing exports.
                  </p>
                </div>
                <CostTable costs={costs} onSelectUser={handleSelectUser} />
              </div>
            )}

            {/* Tab: Licenses */}
            {activeTab === 'licenses' && governance && (
              <div className="space-y-4">
                <div>
                  <h2 className="text-lg font-bold text-white tracking-tight">Gemini Enterprise License Governance</h2>
                  <p className="text-xs text-google-gray-400">
                    Track seat quota utilization, active developer engagement, and reclaim inactive seats to reduce license expenses.
                  </p>
                </div>
                <LicenseTable governance={governance} onSelectUser={handleSelectUser} />
              </div>
            )}
          </>
        )}
      </main>

      {/* Developer Detail Drilldown Modal */}
      <UserModal
        user={selectedUser}
        onClose={() => setSelectedUser(null)}
        onOpenDashboard={(userId) => {
          setSelectedUser(null);
          setDashboardUserId(userId);
          setActiveTab('my-consumption');
        }}
      />

      {/* Footer */}
      <footer className="border-t border-[#1e2330] py-6 text-center text-xs text-google-gray-500">
        <p>
          With &lt;3 by{' '}
          <a
            href="https://github.com/julienbreux"
            target="_blank"
            rel="noopener noreferrer"
            className="text-google-gray-400 hover:text-white transition-colors underline underline-offset-2"
          >
            Julien Breux
          </a>
        </p>
      </footer>
    </div>
  );
}

export default App;
