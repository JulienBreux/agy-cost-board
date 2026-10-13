# Specification: Native Currency Support Across Engine, CLI, TUI, Statusline, and Web

## 1. Overview
Currently, `agy-cost-board` partially assumes USD throughout its presentation and calculation layers:
- The attribution engine queries BigQuery for the billing export's `currency` column, but hardcodes `Currency: "USD"` when assembling `AllocatedUserCost`, `UserCostSummary`, and `OverviewMetrics`.
- Several CLI commands, TUI views, and Web UI components hardcode the `$` symbol (e.g., `KPICards`, `CostTable`, `LicenseTable`, `UserModal`, `statusline.sh.tmpl`).
- While demo data specifies `"EUR"`, the hardcoded overrides force it to display as USD.

This feature enables **full native currency preservation and formatting** without currency conversion (FX). The system detects and propagates whatever ISO 4217 currency exists in the Cloud Billing export (e.g., EUR, USD, GBP, JPY, CAD, AUD, etc.) or demo dataset, formatting all amounts consistently across the Go backend, CLI tables, TUI, statusline script, and React web dashboard.

## 2. Functional Requirements

### 2.1 Attribution Engine & Currency Propagation
- `internal/attribution/engine.go`:
  - Determine the primary billing currency from `FetchBilledCosts(ctx, days)`. If multiple rows exist, adopt the dominant currency; fallback to `"USD"` if no billed cost rows are present or currency is empty.
  - In `GetAttributedCosts`, assign the resolved currency to each `domain.AllocatedUserCost.Currency`.
  - In `GetUserSummary`, assign the resolved currency to `domain.UserCostSummary.Currency`.
  - In `GetOverviewMetrics`, assign the resolved currency to `domain.OverviewMetrics.Currency`.
  - In `GetUserConsumptionDriving`, preserve and propagate the resolved currency.
  - Update `domain.GenerateOptimizationTips` to use dynamic currency symbols/text rather than hardcoding `$45/mo`.

### 2.2 Go Domain, CLI & TUI Formatting
- Implement a reusable currency helper in Go (e.g., `internal/domain/currency.go` or `internal/cli/formatters.go`):
  - `CurrencySymbol(currencyCode string) string`: Maps standard ISO 4217 codes to their standard symbols (`USD` -> `$`, `EUR` -> `€`, `GBP` -> `£`, `JPY` -> `¥`, `CAD` -> `CA$`, `AUD` -> `A$`, `CHF` -> `CHF `, `INR` -> `₹`, `BRL` -> `R$`, etc.), falling back to `${code} ` for unknown currencies.
  - `FormatCost(amount float64, currencyCode string) string`: Formats with two decimal places and the appropriate symbol/prefix.
- Update CLI commands:
  - `cost.go`: Render cost column using the item's currency symbol.
  - `user.go`: Render user summary and per-model breakdown using the user's currency symbol.
  - `license.go`: Render estimated savings using the resolved currency symbol.
- Update TUI (`internal/tui/model.go`):
  - Replace hardcoded `$` in KPI cards (`TOTAL BILLED`, `POTENTIAL SAVINGS`), cost list rows, and user breakdown details with `FormatCost` / `CurrencySymbol`.

### 2.3 Antigravity CLI Statusline Script (`/statusline.sh`)
- `internal/server/statusline.sh.tmpl`:
  - When parsing `/api/v1/users/{id}`, parse both `total_cost` and `currency` (e.g. from `.currency // "USD"`).
  - Cache the currency symbol or store the currency in the cache alongside the cost.
  - Replace hardcoded `\$` in `COST_FMT` with the resolved currency symbol (e.g. `€` for EUR, `$` for USD, `£` for GBP).
  - Fallback gracefully to `$` if currency cannot be determined.

### 2.4 Web Dashboard (React + TypeScript)
- Centralized Currency Formatter (`web/src/utils/currency.ts`):
  - `formatCurrency(amount: number, currency?: string, decimals?: number): string`: Uses `Intl.NumberFormat` with `{ style: 'currency', currency: currency || 'USD' }` or custom options to cleanly format any ISO 4217 currency.
  - `getCurrencySymbol(currency?: string): string`: Extracts the native symbol from `Intl.NumberFormat` parts with fallback to the currency code.
- Component Updates:
  - `KPICards.tsx`: Accept `currency` from `overview.currency`; format `estimated_monthly_savings` with the resolved currency.
  - `CostTable.tsx`: Accept `currency` or read from row items; format `allocated_cost` with the resolved currency.
  - `LicenseTable.tsx`: Accept `currency` or pass through from `overview`; format `estimated_monthly_savings` with the resolved currency.
  - `CostChart.tsx`: Update chart axis, tooltips, summary values, and the `$` toggle button to display the dynamic `currencySymbol`.
  - `PersonalCostChart.tsx`: Update chart axis, tooltips, summary values, and toggle button to display the dynamic `currencySymbol`.
  - `BudgetProgressBar.tsx`: Use `formatCurrency` / `getCurrencySymbol`.
  - `OptimizationCard.tsx`: Use `formatCurrency` / `getCurrencySymbol`.
  - `RecentActivityTable.tsx`: Use `formatCurrency` / `getCurrencySymbol`.
  - `PersonalKPICards.tsx`: Use `formatCurrency` / `getCurrencySymbol`.
  - `StatuslineHelpBox.tsx`: Use dynamic currency symbol.
  - `UserModal.tsx`: Replace hardcoded `$` in breakdown rows with `formatCurrency`.

## 3. Tech Stack
- **Backend:** Go 1.23+ (`bigquery`, `cobra`, `bubbletea`, `lipgloss`)
- **Frontend:** React 18, TypeScript, Vite, Tailwind CSS, Lucide icons
- **Testing:** Go standard testing (`go test -race`), Vitest + React Testing Library

## 4. Commands
- Build Web: `npm --prefix web run build`
- Test Web: `npm --prefix web test -- --run`
- Test Go: `go test -v -race ./...`
- Lint Go: `go tool golangci-lint run ./...`
- Run Demo CLI: `go run ./cmd/agy-cost-board cost --demo`
- Run Demo Server: `go run ./cmd/agy-cost-board serve --demo`

## 5. Project Structure
```
internal/
  attribution/
    engine.go          → Extract and propagate native currency from billed costs
  domain/
    currency.go        → ISO 4217 symbol resolver and formatting functions
    currency_test.go   → Comprehensive unit tests for currency formatting
    models.go          → Model updates (tips with dynamic currency formatting)
  cli/
    cost.go            → Use currency formatting helper
    user.go            → Use currency formatting helper
    license.go         → Use currency formatting helper
  tui/
    model.go           → Use currency formatting helper
  server/
    statusline.sh.tmpl → Dynamic currency resolution in bash script
web/
  src/
    utils/
      currency.ts      → Centralized Intl.NumberFormat helper
      currency.test.ts → Unit tests for ISO 4217 currencies (USD, EUR, GBP, JPY, etc.)
    components/
      KPICards.tsx
      CostTable.tsx
      LicenseTable.tsx
      CostChart.tsx
      PersonalCostChart.tsx
      BudgetProgressBar.tsx
      OptimizationCard.tsx
      RecentActivityTable.tsx
      PersonalKPICards.tsx
      StatuslineHelpBox.tsx
      UserModal.tsx
```

## 6. Code Style & Conventions
- **Go:**
  ```go
  // CurrencySymbol resolves standard symbols for common ISO-4217 currencies.
  func CurrencySymbol(code string) string {
      switch strings.ToUpper(strings.TrimSpace(code)) {
      case "EUR": return "€"
      case "GBP": return "£"
      case "JPY": return "¥"
      case "USD", "": return "$"
      default: return code + " "
      }
  }
  ```
- **TypeScript:**
  ```ts
  export function formatCurrency(
    amount: number,
    currency = 'USD',
    maximumFractionDigits = 2
  ): string {
    try {
      return new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency: currency.toUpperCase(),
        maximumFractionDigits,
        minimumFractionDigits: maximumFractionDigits > 0 ? 2 : 0,
      }).format(amount);
    } catch {
      return `${currency} ${amount.toFixed(maximumFractionDigits)}`;
    }
  }
  ```

## 7. Testing Strategy
- **Go Tests:**
  - Test `CurrencySymbol` and `FormatCost` across standard currencies (`USD`, `EUR`, `GBP`, `JPY`, unknown/custom codes).
  - Test `Engine.GetAttributedCosts`, `Engine.GetUserSummary`, `Engine.GetOverviewMetrics` with simulated billed costs having `"EUR"`, verifying all output structures preserve `"EUR"`.
  - Test fallback to `"USD"` when billed costs slice is empty or has blank currency.
  - Verify CLI table outputs render with expected currency symbol.
- **Web Tests:**
  - Unit tests for `currency.ts` verifying formatting for USD, EUR, GBP, JPY, CAD.
  - Component tests ensuring `KPICards`, `CostTable`, `LicenseTable`, `CostChart`, and `PersonalCostChart` render matching symbols for EUR, USD, and custom ISO codes.

## 8. Boundaries
- **Always:**
  - Preserve exact numerical cost values from BigQuery/billing without any exchange rate manipulation or multiplier.
  - Fallback to `"USD"` if currency is empty or not provided.
  - Run all Go and web tests before committing changes.
- **Ask First:**
  - Any change to BigQuery SQL query schemas.
  - Adding new npm or Go third-party dependencies.
- **Never:**
  - Make external HTTP network requests to exchange rate APIs.
  - Hardcode currency strings (`$`, `USD`) in component render logic.

## 9. Success Criteria
- [ ] Attribution engine accurately extracts and propagates the currency from BigQuery billing exports across all metrics and summary models.
- [ ] Demo mode (`--demo`) uses `"EUR"` from synthetic data and renders everywhere as `€` (or `EUR`).
- [ ] Go CLI (`cost`, `user`, `license`) and TUI properly display the detected currency symbol instead of hardcoded `$`.
- [ ] Statusline script `/statusline.sh` dynamically outputs the user's currency symbol.
- [ ] Web application uses a single centralized `currency.ts` utility backed by `Intl.NumberFormat`.
- [ ] All 11 web dashboard components render dynamic currencies without any hardcoded `$` signs.
- [ ] All Go tests and web tests pass with 100% green status.

## 10. Open Questions
- None. Requirements confirmed: no conversion/FX, retain native billing export currency, support all ISO 4217 codes via `Intl.NumberFormat` on web and standard symbol map in Go/CLI/TUI.
