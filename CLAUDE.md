# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QRL Proof-of-Stake Explorer - A blockchain explorer for the Quantum Resistant Ledger (QRL) Zond network. Three main components sync blockchain data, serve it via REST API, and display it in a web UI.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌───────────────┐     ┌──────────────┐
│  QRL Zond Node  │────▶│  Zond2mongoDB    │────▶│   MongoDB     │◀────│  backendAPI  │
│  (RPC :8545)    │     │  (synchronizer)  │     │  (qrldata-z)  │     │ (REST :8081) │
└─────────────────┘     └──────────────────┘     └───────────────┘     └──────────────┘
                                                                              │
                                                                              ▼
                                                                    ┌──────────────────┐
                                                                    │ ExplorerFrontend │
                                                                    │  (Next.js :3000) │
                                                                    └──────────────────┘
```

Note: Backend uses port 8081 locally since beacon chain occupies 8080. Production typically uses 8080 or proxied through nginx.

## Build & Run Commands

### Frontend (ExplorerFrontend/)
```bash
npm install           # Install dependencies
npm run dev           # Development server (port 3000)
npm run build         # Production build
npm start             # Production server
npm test              # Run tests
npm run lint          # Run linting
```

### Backend API (backendAPI/)
```bash
go mod download                    # Install dependencies
go build -o backendAPI main.go     # Build executable
./backendAPI                       # Run server (port 8080)
go test ./...                      # Run tests
```

### Synchronizer (Zond2mongoDB/)
```bash
go mod download                     # Install dependencies
go build -o synchroniser main.go    # Build executable
./synchroniser                      # Run synchronizer
```

### Full Stack Deployment (with PM2)
```bash
./deploy.sh                         # Full setup (Linux/macOS)
./deploy-windowsgitbash.sh          # Full setup (Windows Git Bash)
```

### Update Scripts
```bash
./update-backend.sh                 # Rebuild and restart backend + syncer
./update-frontend.sh                # Rebuild and restart frontend
```

## Key Environment Variables

**Frontend (.env):**
- `DATABASE_URL` - MongoDB connection string
- `NEXT_PUBLIC_HANDLER_URL` - Backend API URL (e.g., http://localhost:8080)
- `NEXT_PUBLIC_DOMAIN_NAME` - Frontend domain

**Backend (.env):**
- `MONGOURI` - MongoDB connection string
- `NODE_URL` - Zond RPC endpoint (e.g., http://localhost:8545)
- `HTTP_PORT` - API port (default :8080)

**Synchronizer (.env):**
- `MONGOURI` - MongoDB connection string (without database name)
- `NODE_URL` - Zond RPC endpoint for block sync
- `MEMPOOL_NODE_URL` - (Optional) Separate RPC endpoint for mempool/pending tx detection. Falls back to NODE_URL if not set. Useful when using a public RPC for blocks but local node for mempool.
- `BEACONCHAIN_API` - Beacon chain HTTP API endpoint

## MongoDB Collections (database: qrldata-z)

Core data:
- `blocks` - Block headers and transactions
- `transactionByAddress` - Indexed transactions by address
- `addresses` - Wallet balances and metadata
- `pending_transactions` - Mempool transactions (status: pending/mined)
- `validators` - Single document containing all validators per epoch
- `contractCode` - Smart contract deployments

Analytics:
- `coingecko` - Market price data
- `walletCount` - Total wallet metrics
- `dailyTransactionsVolume` - Volume tracking
- `totalCirculatingQuanta` - Supply tracking

## Code Organization

### Frontend (Next.js 15 App Router)
- `app/` - Pages and API routes
- `app/components/` - Shared components (SearchBar, Sidebar, AreaChart)
- `app/lib/helpers.ts` - Formatting and conversion utilities
- Pattern: Server Components (`page.tsx`) fetch data; Client Components (`*-client.tsx`) handle interactivity

### Backend API (Go + Gin)
- `configs/` - Environment and MongoDB setup
- `db/` - Database query functions (one file per entity type)
- `handler/` - Request processing and middleware
- `models/` - Data structures
- `routes/routes.go` - All REST API endpoints

### Synchronizer (Go)
- `synchroniser/sync.go` - Main sync loop, batch processing, block insertion
- `synchroniser/pending_sync.go` - Mempool transaction sync (every 5s)
- `db/` - Database operations for syncing
- `rpc/` - Zond node RPC client
- `services/validator_service.go` - Validator data processing

## Sync Constants (Zond2mongoDB)

```go
DefaultBatchSize       = 64    // Normal batch size
LargeBatchSize         = 128   // When >1000 blocks behind
BatchSyncThreshold     = 64    // Triggers batch mode
MaxProducerConcurrency = 8     // Parallel block fetchers
MEMPOOL_SYNC_INTERVAL  = 5s    // Pending tx polling
MAX_PENDING_AGE        = 24h   // Pending tx cleanup threshold
```

## Recent Fixes

1. **Pending Transaction Lifecycle (Fixed 2026-01-03)**: Fixed bug where mined transactions still showed as pending. The issue was in `backendAPI/routes/routes.go` - the `/pending-transaction/:hash` endpoint was returning mined transactions instead of returning 404. Also updated `ExplorerFrontend/app/tx/[query]/page.tsx` to check `status === 'pending'` before showing pending view.

## Known Issues & Areas Needing Work

1. **Missing Blocks**: Some blocks that exist on the node may not be properly synced. Compare node data (via RPC) with MongoDB to identify gaps.

## Testing Against Local Node

A local QRL Zond node is available for testing. Compare:
- Node RPC (localhost:8545) - source of truth for blockchain data
- MongoDB logs - what the syncer has stored
- Syncer logs - operational status of sync process

## Git Workflow

- **Default branch**: Always work on the `dev` branch for new features and fixes
- **Never commit directly to `main`** - all changes go through PRs
- **PR process**:
  1. Commit and push changes to `dev`
  2. Create PR from `dev` to `main`
  3. Wait for Gemini's automated review (3-10 minutes depending on PR size)
  4. Address all review comments before merging
  5. Only merge after review comments are resolved

## Commit Convention

Use conventional commits: `feat:`, `fix:`, `perf:`, `docs:`, `chore:`, `test:`

## Data Format Notes

- Numeric values stored as hex strings with "0x" prefix in blocks/transactions
- Addresses and hashes stored in hex format (WITH "0x" prefix in most places)
- Timestamps are Unix timestamps in decimal or hex format
- Validator data stores epochs/balances as decimal strings
