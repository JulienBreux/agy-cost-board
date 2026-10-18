import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TokensEffortClientSection } from './TokensEffortClientSection';
import { TokenCostShare, ModelEffortStats, ClientStats } from '../api';

describe('TokensEffortClientSection Component', () => {
  const mockShares: TokenCostShare[] = [
    {
      category: 'Uncached Input',
      tokens: 400000,
      token_share: 40.0,
      cost: 12.0,
      cost_share: 24.0,
    },
    {
      category: 'Cached Input',
      tokens: 450000,
      token_share: 45.0,
      cost: 2.5,
      cost_share: 5.0,
    },
    {
      category: 'Output',
      tokens: 100000,
      token_share: 10.0,
      cost: 20.5,
      cost_share: 41.0,
    },
    {
      category: 'Thinking',
      tokens: 50000,
      token_share: 5.0,
      cost: 15.0,
      cost_share: 30.0,
    },
  ];

  const mockEffortStats: ModelEffortStats[] = [
    {
      model: 'gemini-2.5-pro',
      effort_level: 'high',
      calls: 42,
      tokens: 650000,
      cached_tokens: 300000,
      cache_hit_rate: 46.2,
      wtu: 35000,
      included_cost: 22.0,
      payg_cost: 8.0,
      total_cost: 30.0,
    },
    {
      model: 'claude-3-7-sonnet',
      effort_level: 'max',
      calls: 15,
      tokens: 350000,
      cached_tokens: 150000,
      cache_hit_rate: 42.9,
      wtu: 20000,
      included_cost: 0.0,
      payg_cost: 20.0,
      total_cost: 20.0,
    },
  ];

  const mockClients: ClientStats[] = [
    {
      client_name: 'vscode',
      calls: 35,
      tokens: 600000,
      wtu: 32000,
      total_cost: 30.0,
      share_of_cost: 60.0,
    },
    {
      client_name: 'agy-cli',
      calls: 22,
      tokens: 400000,
      wtu: 23000,
      total_cost: 20.0,
      share_of_cost: 40.0,
    },
  ];

  it('renders Where Tokens and Money Go, Models & Thinking Effort, and Antigravity Clients sections', () => {
    render(
      <TokensEffortClientSection
        tokenCostShares={mockShares}
        modelEffortStats={mockEffortStats}
        clientBreakdown={mockClients}
        currency="USD"
      />
    );

    // Section 1: Where Tokens and Money Go
    expect(screen.getByText('Where Tokens and Money Go')).toBeInTheDocument();
    expect(screen.getByText('Uncached Input')).toBeInTheDocument();
    expect(screen.getByText('Cached Input')).toBeInTheDocument();
    expect(screen.getByText('Output')).toBeInTheDocument();
    expect(screen.getByText('Thinking')).toBeInTheDocument();
    expect(screen.getByText('41.0%')).toBeInTheDocument();

    // Section 2: Models & Thinking Effort
    expect(screen.getByText('Models & Thinking Effort')).toBeInTheDocument();
    expect(screen.getByText('gemini-2.5-pro')).toBeInTheDocument();
    expect(screen.getByText('high')).toBeInTheDocument();
    expect(screen.getByText('46.2%')).toBeInTheDocument();
    expect(screen.getByText('claude-3-7-sonnet')).toBeInTheDocument();
    expect(screen.getByText('max')).toBeInTheDocument();

    // Section 3: Antigravity Clients
    expect(screen.getByText('Antigravity Clients')).toBeInTheDocument();
    expect(screen.getByText('vscode')).toBeInTheDocument();
    expect(screen.getByText('agy-cli')).toBeInTheDocument();
    expect(screen.getByText('60.0%')).toBeInTheDocument();
  });

  it('returns null when all arrays are empty', () => {
    const { container } = render(
      <TokensEffortClientSection
        tokenCostShares={[]}
        modelEffortStats={[]}
        clientBreakdown={[]}
        currency="USD"
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it('handles raw Go backend JSON field names (total_tokens, allocated_cost, cost_usd, label, token_share) without throwing', () => {
    render(
      <TokensEffortClientSection
        tokenCostShares={[
          {
            category: 'uncached_input',
            label: 'Uncached Input',
            tokens: 850,
            token_share: 55.5,
            cost_usd: 14.25,
            cost_share: 32.1,
          },
        ]}
        modelEffortStats={[
          {
            model: 'gemini-2.5-pro',
            effort_level: 'high',
            calls: 12,
            total_tokens: 950,
            cached_tokens: 400,
            cache_hit_rate: 42.1,
            wtu: 1850,
            included_cost: 10.0,
            payg_cost: 4.25,
            allocated_cost: 14.25,
          },
        ]}
        clientBreakdown={[
          {
            client_name: 'vscode',
            calls: 12,
            total_tokens: 950,
            wtu: 1850,
            allocated_cost: 14.25,
            token_share: 100.0,
          },
        ]}
        currency="USD"
      />
    );

    expect(screen.getByText('Uncached Input')).toBeInTheDocument();
    expect(screen.getByText(/850 tok/)).toBeInTheDocument();
    expect(screen.getByText('950')).toBeInTheDocument();
    expect(screen.getByText(/950 tok/)).toBeInTheDocument();
    expect(screen.getByText('100.0%')).toBeInTheDocument();
  });
});
