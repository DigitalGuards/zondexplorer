#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting backend update and deployment...${NC}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if required commands exist
if ! command_exists go; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

if ! command_exists pm2; then
    echo -e "${RED}Error: PM2 is not installed${NC}"
    exit 1
fi

if ! command_exists git; then
    echo -e "${RED}Error: Git is not installed${NC}"
    exit 1
fi

# Update from git
echo -e "${YELLOW}Pulling latest changes from git...${NC}"
git pull
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Git pull failed${NC}"
    exit 1
fi
echo -e "${GREEN}Git pull completed successfully${NC}"

# Deploy Zond2mongoDB synchronizer
echo -e "${YELLOW}Building and deploying synchronizer...${NC}"
cd Zond2mongoDB
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Could not find Zond2mongoDB directory${NC}"
    exit 1
fi

go build -o syncer
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to build synchroniser${NC}"
    exit 1
fi

# Stop existing PM2 processes if they exist
pm2 stop synchroniser 2>/dev/null
pm2 delete synchroniser 2>/dev/null

# Start synchroniser with PM2
pm2 start ./syncer --name synchroniser
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to start synchroniser with PM2${NC}"
    exit 1
fi
echo -e "${GREEN}Synchroniser deployed successfully${NC}"

# Deploy BackendAPI server
echo -e "${YELLOW}Building and deploying server...${NC}"
cd /BackendAPI
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Could not find BackendAPI directory${NC}"
    exit 1
fi

go build -o handler
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to build handler${NC}"
    exit 1
fi

# Stop existing PM2 processes if they exist
pm2 stop handler 2>/dev/null
pm2 delete handler 2>/dev/null

# Start handler with PM2
pm2 start ./handler --name handler
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to start handler with PM2${NC}"
    exit 1
fi
echo -e "${GREEN}Handler deployed successfully${NC}"

# Save PM2 configuration
pm2 save

echo -e "${GREEN}Backend update and deployment completed successfully!${NC}"
echo -e "${YELLOW}PM2 process status:${NC}"
pm2 list

# Return to original directory
cd ../
