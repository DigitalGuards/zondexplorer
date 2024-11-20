# Quanta Explorer Server

The backend server component of the Quanta Explorer application, written in Go. This server provides API endpoints for the frontend to fetch blockchain data and interact with the QRL Zond network.

## Project Structure

```
server/
├── configs/           # Configuration management
│   ├── const.go      # Constants and configuration values
│   ├── env.go        # Environment variable handling
│   └── setup.go      # Application setup and initialization
├── db/               # Database operations
│   ├── db.go         # Main database operations
│   └── db_test.go    # Database tests
├── handler/          # Request handlers
│   └── handler.go    # HTTP request handlers
├── models/           # Data models
│   ├── address.go    # Address-related structures
│   ├── blocksize.go  # Block size tracking
│   ├── coinbase.go   # Coinbase transaction models
│   ├── coingecko.go  # CoinGecko price data models
│   ├── contract.go   # Smart contract models
│   ├── jsonrpc.go    # JSON-RPC request/response structures
│   ├── trace.go      # Transaction trace models
│   ├── transactionbyaddress.go  # Address transaction models
│   ├── transfer.go   # Token transfer models
│   ├── validator.go  # Validator data models
│   ├── wallet.go     # Wallet-related structures
│   └── zond.go       # Zond-specific models
├── resources/        # Static resources
│   └── favicon.ico
├── routes/           # API route definitions
│   └── routes.go     # Route handlers and middleware
├── main.go          # Application entry point
├── go.mod           # Go module definition
└── go.sum           # Go module checksums
```

## Key Components

### Configs
Handles all configuration-related functionality including:
- Environment variable management
- Constants and configuration values
- Application setup and initialization

### Database (db/)
Manages database operations and interactions:
- Connection management
- CRUD operations
- Database testing

### Handlers (handler/)
Contains HTTP request handlers that process incoming requests and return appropriate responses.

### Models (models/)
Defines data structures for:
- Blockchain addresses
- Block information
- Transactions
- Smart contracts
- Validators
- Wallet data
- Price data from CoinGecko
- JSON-RPC communication

### Routes (routes/)
Defines API endpoints and their corresponding handlers:
- Route registration
- Middleware configuration
- Request routing

## Environment Configuration

The server uses two environment files:
- `.env.development` for development environment
- `.env.production` for production environment

## Getting Started

1. Ensure Go is installed on your system
2. Clone the repository
3. Navigate to the server directory
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Set up appropriate environment variables
6. Run the server:
   ```bash
   go run main.go
   ```

## Development

- Follow Go best practices and conventions
- Ensure proper error handling
- Write tests for new functionality
- Update documentation when adding new features

## API Documentation

The server provides various API endpoints for:
- Block information
- Transaction data
- Address details
- Smart contract interaction
- Validator information
- Network statistics

For detailed API documentation, refer to the handler implementations in `handler/handler.go` and route definitions in `routes/routes.go`.