import { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught React render error:', error, errorInfo);
  }

  public handleReload = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-[#0f1115] text-white flex items-center justify-center p-6">
          <div className="bg-[#14171f] border border-[#282f40] rounded-2xl p-8 max-w-lg w-full text-center space-y-4 shadow-2xl">
            <div className="w-12 h-12 rounded-xl bg-google-red/10 border border-google-red/20 text-google-red flex items-center justify-center mx-auto">
              <AlertTriangle className="h-6 w-6" />
            </div>
            <h2 className="text-lg font-bold text-white tracking-tight">Dashboard Render Error</h2>
            <p className="text-xs text-google-gray-400">
              An unexpected error occurred while rendering the dashboard UI components:
            </p>
            <div className="bg-[#1a1e28] p-3 rounded-lg text-left text-xs font-mono text-google-red/90 overflow-x-auto border border-[#282f40]">
              {this.state.error?.message || 'Unknown render exception'}
            </div>
            <button
              onClick={this.handleReload}
              className="inline-flex items-center space-x-2 px-4 py-2 bg-google-blue hover:bg-google-blue/90 text-white rounded-lg text-xs font-semibold transition-colors"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Reload Dashboard</span>
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
