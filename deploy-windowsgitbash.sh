#!/bin/bash

# Print colored output
print_status() {
    echo -e "\e[1;34m>>> $1\e[0m"
}

print_error() {
    echo -e "\e[1;31m>>> Error: $1\e[0m"
    exit 1
}

print_status "Deploying frontend is currently commented out. Uncomment the function or deploy it manually."

# Clean PM2 logs and processes
clean_pm2() {
    print_status "Cleaning PM2 logs and processes..."
    
    # Delete all PM2 logs
    pm2 flush || print_status "No logs to flush"
    
    # Stop and delete only processes started by this deployment
    for name in handler syncer frontend; do
        pm2 delete $name || print_status "No process named $name to delete"
    done
    
    # Clear PM2 dump file
    pm2 cleardump || print_status "No dump file to clear"
}

# Clean MongoDB database and log files
clean_database_and_logs() {
    print_status "Do you want to clean MongoDB database and log files? (y/n)"
    read -p "Enter choice: " DELETE_DB_LOGS
    
    if [[ $DELETE_DB_LOGS =~ ^[Yy]$ ]]; then
        print_status "Cleaning MongoDB database and log files..."
        
        # Drop the MongoDB database
        mongosh --eval "db.getSiblingDB('qrldata-z').dropDatabase()" || print_status "Failed to drop database or database doesn't exist"
        export BASE_DIR=$(pwd)
        # Delete the log file if it exists
        if [ -f "$BASE_DIR/Zond2mongoDB/logs/zond_sync.log" ]; then
            rm "$BASE_DIR/Zond2mongoDB/logs/zond_sync.log" || print_status "Failed to delete log file"
        else
            print_status "Log file not found, skipping deletion"
        fi
    else
        print_status "Skipping database and log files cleanup"
    fi
}

# Check for required tools
check_dependencies() {
    print_status "Checking dependencies..."

    command -v node >/dev/null 2>&1 || { print_error "Node.js is required but not installed."; }
    command -v npm >/dev/null 2>&1 || { print_error "npm is required but not installed."; }
    command -v go >/dev/null 2>&1 || { print_error "Go is required but not installed."; }
    #command -v mongod >/dev/null 2>&1 || { print_error "MongoDB is required but not installed."; }

    # Install PM2 if not present
    if ! command -v pm2 >/dev/null 2>&1; then
        print_status "Installing PM2..."
        npm install -g pm2 || print_error "Failed to install PM2"
    fi
}

# Check if MongoDB is running
check_mongodb() {
   if ! nc -z localhost 27017; then
        print_error "MongoDB is not running on localhost:27017. Or nc is not installed..."
    fi
}

# Prompt for node selection
select_node() {
    print_status "Select Zond node to use:"
    PS3="Please choose the node (1-4): "
    options=("Local node (127.0.0.1:8545)" "BETANET Remote node (95.170.68.91:8545)" "DG TestnetV1 node (35.158.17.89:32776)" "Foundation testnetv1 (buidl.localbits.org:8545)")
    select opt in "${options[@]}"
    do
        case $opt in
            "Local node (127.0.0.1:8545)")
                NODE_URL="http://127.0.0.1:8545"
                break
                ;;
            "Remote node (95.170.68.91:8545)")
                NODE_URL="http://95.170.68.91:8545"
                break
                ;;
            "DG TestnetV1 node (35.158.17.89:32776)")
                NODE_URL="http://35.158.17.89:32776"
                break
                ;;
            "Foundation testnetv1 (buidl.localbits.org:8545)")
                NODE_URL="http://buidl.localbits.org:8545"
                break
                ;;
            *) echo "Invalid option. Please try again.";;
        esac
    done
    print_status "Selected node: $NODE_URL"
    export NODE_URL
}

# Check if Zond node is accessible
check_zond_node() {
    RESPONSE=$(curl --silent --fail -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","id":1,"method":"net_listening","params":[]}' \
        $NODE_URL)

    if [[ $? -ne 0 || -z "$RESPONSE" ]]; then
        print_error "Zond node is not accessible at $NODE_URL"
    fi
}

# Check if port is available
check_port() {
    PORT=$1
    if netstat -ano | findstr :$PORT > nul; then
        print_error "Port $PORT is already in use."
    fi
}

# Clone the repository
clone_repo() {
    if [ -d ".git" ]; then
        print_status "Repository already exists. Checking git status..."
        git status
        
        read -p "Would you like to pull the latest changes? (y/n): " PULL_CHANGES
        if [[ $PULL_CHANGES =~ ^[Yy]$ ]]; then
            print_status "Pulling latest changes..."
            git pull || print_error "Failed to pull latest changes"
        else
            print_status "Skipping pull, continuing with existing code..."
        fi
    else
        print_status "Cloning QRL Explorer repository..."
        git clone https://github.com/DigitalGuards/zondexplorer.git ../zondexplorer || print_error "Failed to clone repository"
        cd ../backendAPI || print_error "Failed to enter project directory"
    fi
    export BASE_DIR=$(pwd)
}

# Setup server environment
setup_server() {
    print_status "Setting up server..."
    cd "$BASE_DIR/backendAPI" || print_error "Server directory not found"

    # Create .env.development file
    print_status "Creating .env.development file..."
    cat > .env << EOL
GIN_MODE=release
MONGOURI=mongodb://localhost:27017/qrldata-z?readPreference=primary
HTTP_PORT=:8080
NODE_URL=$NODE_URL
EOL

    # Build the server with explicit output name
    print_status "Building server..."
    go build -o backendAPI.exe main.go || print_error "Failed to build server"

    # Start server with PM2, specifying the working directory and APP_ENV
    print_status "Starting server with PM2..."
    pm2 start ./backendAPI.exe --name "handler" --cwd "$BASE_DIR/backendAPI" || print_error "Failed to start server"
}

# Setup frontend environment
setup_frontend() {
    print_status "Setting up frontend..."
    cd "$BASE_DIR/ExplorerFrontend" || print_error "Frontend directory not found"

    # Create .env file
    cat > .env << EOL
DATABASE_URL=mongodb://localhost:27017/qrldata-z?readPreference=primary
DOMAIN_NAME=http://localhost:3000
HANDLER_URL=http://127.0.0.1:8080
EOL

    # Create .env.local file
    cat > .env.local << EOL
DATABASE_URL=mongodb://localhost:27017/qrldata-z?readPreference=primary
DOMAIN_NAME=http://localhost:3000
HANDLER_URL=http://127.0.0.1:8080
EOL

    # Install dependencies
    print_status "Installing frontend dependencies..."
    npm install || print_error "Failed to install frontend dependencies"

    # Update browserslist database
    print_status "Updating browserslist database..."
    npx browserslist@latest --update-db || print_error "Failed to update browserslist"

    # Start frontend in development mode with PM2 using bash
    print_status "Starting frontend in development mode..."
    pm2 start bash --name "frontend" -- -c "npm run dev" || print_error "Failed to start frontend"
}

# Setup blockchain synchronizer
setup_synchronizer() {
    print_status "Setting up blockchain synchronizer..."
    cd "$BASE_DIR/Zond2mongoDB" || print_error "Synchronizer directory not found"

    # Create .env file
    cat > .env << EOL
MONGOURI=mongodb://localhost:27017
NODE_URL=$NODE_URL
BEACONCHAIN_API=http://95.170.68.91:3500
EOL
    # Build synchronizer with explicit output name
    print_status "Building synchronizer..."
    go build -o zsyncer.exe main.go || print_error "Failed to build synchronizer"

    # Start synchronizer with PM2, explicitly setting environment variables
    print_status "Starting synchronizer with PM2..."
    pm2 start ./zsyncer.exe --name "syncer" --cwd "$BASE_DIR/Zond2mongoDB" || print_error "Failed to start synchronizer"
}

# Save PM2 processes
save_pm2() {
    print_status "Saving PM2 processes..."
    pm2 save || print_error "Failed to save PM2 processes"
}

# Main deployment function
main() {
    print_status "Starting QRL Explorer deployment..."

    # Clean PM2 logs and processes before starting
    clean_pm2
    
    # Clean database and log files
    clean_database_and_logs

    # Check for required tools
    check_dependencies

    # Prompt for node selection
    select_node

    # Check if MongoDB and Zond node are running
    #check_mongodb
    #check_zond_node

    # Check if required ports are available
    #check_port 3000
    #check_port 8080

    # Clone and setup
    clone_repo
    
    setup_synchronizer
    #setup_frontend
    setup_server
    #save_pm2

    print_status "Deployment complete! Services are starting up..."
    echo -e "\nAccess points:"
    echo "- Frontend: http://localhost:3000"
    echo "- Server API: http://localhost:8080"
    echo -e "\nMake sure you have:"
    echo "1. MongoDB running on localhost:27017"
    echo "2. Zond node accessible at $NODE_URL"
    echo -e "\nTo monitor services:"
    echo "pm2 status"
    echo -e "\nTo view logs:"
    echo "pm2 logs"
    echo -e "\nTo clear logs:"
    echo "pm2 flush"
    echo -e "\nTo stop all services:"
    echo "pm2 stop all"
}

# Run the deployment
main
