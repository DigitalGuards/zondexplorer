# QRL Proof-Of-Stake Explorer

Quantum Resistant Ledger Proof-of-Stake explorer. It is blazing fast with a modern frontend using NextJS and Golang for the backend, stable and compatible with the Ethereum network. Very easy to setup. Synchronising the blockchain node to MongoDB takes only 3 to 5 seconds, depending on your hardware, network speed, and size of the blockchain. Which makes it incredibly easy debug the system, as you can easily delete the blockchain data from MongoDB and simply restart to sync it again. Saves a lot of time.

There are four components you will need:

quanta-explorer-go containing two components, frontend and server (handler), 
QRLtoMongoDB-PoS - which is a blockchain synchroniser to MongoDB and
Zond node (It can be either external or local). 

Note: These instructions are only for the explorer related components. Are you trying to get your Zond up and running? Visit https://test-zond.theqrl.org/linux.html

Make sure to clone quanta-explorer-go and QRLtoMongoDB-PoS first:

```
git clone https://github.com/moscowchill/qrl-explorer-pos.git
```

#### Requirements
- Install golang, mongodb/mongosh, pm2 packages - check their documentations
- Ubuntu 20.04 LTS system

### Frontend
Now cd into quanta-explorer-go

```
cd quanta-explorer-go/frontend
```

Create the .env files if you don't have them
```
touch .env && touch .env.local
```

#### .env fields

| VARIABLE | VALUE |
| ------ | ------ |
| DATABASE_URL | mongodb://localhost:27017/qrldata?readPreference=primary |
| NEXTAUTH_URL | 127.0.0.1 |
| NEXT_PUBLIC_DOMAIN_NAME | http://localhost:3000 (dev) OR http://your_domain_name.io (prod) |
| NEXT_PUBLIC_HANDLER_URL | http://localhost:8080 (dev) OR http://your_domain_name.io:8443 (prod) |

#### .env.local fields 

| VARIABLE | VALUE |
| ------ | ------ |
| DATABASE_URL | mongodb://localhost:27017/qrldata?readPreference=primary |
| NEXTAUTH_SECRET | YOUR_SECRET |
| ADMIN_PUBLIC_ADDRESS | YOUR_SECRET |
| DOMAIN_NAME | http://localhost:3000 (dev) OR http://your_domain_name.io (prod) |
| HANDLER_URL | http://localhost:8080 (dev) OR http://your_domain_name.io:8443 (prod) |

Next, run the following command
```
npm run build
```

```
pm2 start npm --name "frontend" -- start
```
End of instructions for Frontend

### Server (Handler)

Cd into the server folder and create two .env files if you don't have them: 
```
touch .env.development && touch .env.production
```

#### .env.development fields 

| VARIABLE | VALUE |
| ------ | ------ |
| GIN_MODE | release |
| MONGOURI | mongodb://localhost:27017/qrldata?readPreference=primary |
| HTTP_PORT | :8080 |
| NODE_URL | http://localhost:8545 |

#### .env.production fields

| VARIABLE | VALUE |
| ------ | ------ |
| GIN_MODE | release |
| MONGOURI | mongodb://localhost:27017/qrldata?readPreference=primary |
| CERT_PATH | PATH_TO_CERT |
| KEY_PATH | PATH_TO_KEY |
| HTTPS_PORT | :8443 |
| NODE_URL | http://localhost:8545 |

Now we need to build the main.go file
```
go build main.go
```
Now we have a main executable. We can add it to pm2.

```
pm2 start ./main --name "handler" -- start
```
End of Server (Handler) instructions

### QRLtoMongoDB-PoS (Blockchain to MongoDB Synchroniser) 

Cd into the QRLtoMongoDB-PoS directory
```
touch .env
```

#### .env fields
| VARIABLE | VALUE |
| ------ | ------ |
| MONGOURI | mongodb://localhost:27017 |
| NODE_URL | http://localhost:8545 |

```
go build main.go
```

Now you have a main executable. We can add it to pm2 

```
pm2 start ./main --name "synchroniser" -- start
```

##### Optional
To save the pm2 processes, run the following command
```
pm2 save
```
Great! That is all. The explorer should now be live. (Don't forget to use tmux or pm2 for Zond too!)

## System Architecture and Data Flow

### Backend Architecture

The backend consists of three main components that work together to provide blockchain data to the frontend:

1. **QRL Zond Node**
   - The source of blockchain data
   - Runs the QRL blockchain protocol
   - Exposes RPC endpoints for data access
   - Handles consensus and network communication

2. **QRLtoMongoDB-PoS (Synchronizer)**
   - Connects to Zond node via RPC
   - Continuously syncs blockchain data to MongoDB
   - Components:
     * Block synchronization (blocks, transactions)
     * Validator data collection
     * Market data aggregation (CoinGecko integration)
     * Wallet statistics
   - Uses producer/consumer pattern for efficient data processing
   - Maintains data consistency with rollback capability

3. **Server (Handler)**
   - REST API server built with Gin framework
   - Provides endpoints for frontend data access
   - Aggregates data from MongoDB collections
   - Handles real-time data requests
   - Manages API rate limiting and caching

### Data Flow

1. **Blockchain to MongoDB**:
   ```
   Zond Node -> RPC Calls -> Synchronizer -> MongoDB
   ```
   - Synchronizer polls Zond node for new blocks
   - Processes block data into structured collections
   - Updates related collections (transactions, addresses, etc.)
   - Maintains indexes for efficient querying

2. **MongoDB to Frontend**:
   ```
   MongoDB -> Handler -> REST API -> Frontend
   ```
   - Handler provides RESTful endpoints
   - Frontend makes HTTP requests to handler
   - Data is cached where appropriate
   - Real-time updates through periodic polling

### Key Collections in MongoDB

1. **Blocks Collection**
   - Stores block headers and metadata
   - Tracks chain progression
   - Links to transactions

2. **Transactions Collection**
   - Stores all transaction data
   - Includes internal transactions
   - Maintains address relationships

3. **Addresses Collection**
   - Tracks account balances
   - Stores contract metadata
   - Maintains transaction history

4. **Validators Collection**
   - Stores validator information
   - Tracks staking data
   - Records validator performance

### Frontend Integration

The frontend communicates with the backend through several key endpoints:

1. **Overview Data**
   - `/overview`: General blockchain statistics
   - Real-time market data and network status

2. **Block Explorer**
   - `/blocks`: Block listing and details
   - `/tx`: Transaction details
   - `/address`: Address information and history

3. **Validator Information**
   - `/validators`: Active validator list
   - Staking statistics and performance metrics

4. **Search Functionality**
   - Unified search across blocks, transactions, and addresses
   - Auto-suggestion and quick navigation

### Error Handling and Recovery

The system includes several reliability features:

1. **Synchronizer Recovery**
   - Automatic recovery from node disconnections
   - Block chain reorganization handling
   - Data consistency checks

2. **Handler Resilience**
   - Graceful error handling
   - Default values for missing data
   - Request timeout management

3. **Frontend Fallbacks**
   - Loading states for async data
   - Error boundaries for component failures
   - Retry mechanisms for failed requests

## License

MIT
