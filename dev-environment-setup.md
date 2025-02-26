# Development Environment Setup Guide

This guide provides detailed instructions for setting up the development environment for the URL Minifier project.

## Prerequisites
Before starting, ensure you have administrator access on your machine and a terminal application installed.

## 1. Go Development Environment

### Installation
1. Visit the official Go downloads page: https://go.dev/dl/
2. Download the latest stable version for your operating system
3. Follow the installation instructions for your platform
4. Verify installation:
   ```bash
   go version
   ```
5. Set up your GOPATH and workspace:
   ```bash
   # Add to your shell profile (.bashrc, .zshrc, etc.)
   export GOPATH=$HOME/go
   export PATH=$PATH:$GOPATH/bin
   ```

### IDE Setup
1. Install Visual Studio Code or GoLand
2. Install Go extension/plugin
3. Configure Go tools:
   ```bash
   go install golang.org/x/tools/gopls@latest
   go install golang.org/x/tools/cmd/goimports@latest
   ```

## 2. Node.js and Package Manager

### Node.js Installation
1. Install Node Version Manager (nvm):
   ```bash
   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
   ```
2. Install LTS version of Node.js:
   ```bash
   nvm install --lts
   nvm use --lts
   ```
3. Verify installation:
   ```bash
   node --version
   ```

### Package Manager Setup
1. npm is included with Node.js installation
2. Verify installation:
   ```bash
   npm --version
   ```

## 3. Docker and Docker Compose

### Docker Installation
1. Install Docker Desktop from https://www.docker.com/products/docker-desktop
2. Start Docker Desktop
3. Verify installation:
   ```bash
   docker --version
   docker compose version
   ```

### Docker Configuration
1. Increase Docker resources (recommended):
   - CPUs: 4
   - Memory: 8GB
   - Swap: 2GB
2. Enable Kubernetes (optional)

## 4. MongoDB

### Local Installation (Optional if using Docker)
1. Follow platform-specific instructions from https://www.mongodb.com/try/download/community
2. Start MongoDB service
3. Verify installation:
   ```bash
   mongosh --version
   ```

### Docker Setup (Recommended)
1. Pull MongoDB image:
   ```bash
   docker pull mongodb/mongodb-community-server
   ```
2. Create data directory:
   ```bash
   mkdir -p ~/mongodb/data
   ```

## 5. Redis

### Local Installation (Optional if using Docker)
1. Follow platform-specific instructions from https://redis.io/download
2. Start Redis service
3. Verify installation:
   ```bash
   redis-cli --version
   ```

### Docker Setup (Recommended)
1. Pull Redis image:
   ```bash
   docker pull redis
   ```
2. Create data directory:
   ```bash
   mkdir -p ~/redis/data
   ```

## 6. Code Quality Setup

### GitHub Actions for Code Quality
Instead of local git hooks, we'll use GitHub Actions to ensure code quality checks are run consistently in the cloud.

1. Create `.github/workflows/backend-quality.yml`:
   ```yaml
   name: Backend Quality

   on:
     push:
       branches: [ main ]
       paths:
         - '**.go'
         - 'go.*'
         - '.golangci.yml'
     pull_request:
       branches: [ main ]
       paths:
         - '**.go'
         - 'go.*'
         - '.golangci.yml'

   jobs:
     quality:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         
         - name: Set up Go
           uses: actions/setup-go@v4
           with:
             go-version: '1.21'
             
         - name: Install golangci-lint
           run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
           
         - name: Run golangci-lint
           run: golangci-lint run ./...
           
         - name: Check formatting
           run: |
             if [ -n "$(gofmt -l .)" ]; then
               echo "The following files are not formatted correctly:"
               gofmt -l .
               exit 1
             fi
             
         - name: Run Go tests
           run: go test -v ./...
   ```

2. Create `.github/workflows/frontend-quality.yml`:
   ```yaml
   name: Frontend Quality

   on:
     push:
       branches: [ main ]
       paths:
         - 'frontend/**'
     pull_request:
       branches: [ main ]
       paths:
         - 'frontend/**'

   jobs:
     quality:
       runs-on: ubuntu-latest
       defaults:
         run:
           working-directory: ./frontend

       steps:
         - uses: actions/checkout@v4

         - name: Setup Node.js
           uses: actions/setup-node@v4
           with:
             node-version: '20'
             cache: 'npm'
             cache-dependency-path: './frontend/package-lock.json'

         - name: Install dependencies
           run: npm ci

         - name: Type checking
           run: npm run type-check

         - name: Lint
           run: npm run lint

         - name: Run tests
           run: npm run test

         - name: Build
           run: npm run build
   ```

3. Create `.golangci.yml` in project root for Go linter configuration:
   ```yaml
   linters:
     enable:
       - gofmt
       - golint
       - govet
       - errcheck
       - staticcheck
       - gosimple
       - ineffassign

   run:
     deadline: 5m
   ```

4. Create frontend configuration files:

   `.eslintrc.json`:
   ```json
   {
     "extends": [
       "eslint:recommended",
       "plugin:@typescript-eslint/recommended",
       "plugin:react-hooks/recommended",
       "plugin:react/recommended"
     ],
     "parser": "@typescript-eslint/parser",
     "plugins": ["@typescript-eslint", "react", "react-hooks"],
     "root": true,
     "settings": {
       "react": {
         "version": "detect"
       }
     }
   }
   ```

   `tsconfig.json`:
   ```json
   {
     "compilerOptions": {
       "target": "ES2020",
       "lib": ["DOM", "DOM.Iterable", "ESNext"],
       "module": "ESNext",
       "skipLibCheck": true,
       "moduleResolution": "bundler",
       "allowImportingTsExtensions": true,
       "resolveJsonModule": true,
       "isolatedModules": true,
       "noEmit": true,
       "jsx": "react-jsx",
       "strict": true,
       "noUnusedLocals": true,
       "noUnusedParameters": true,
       "noFallthroughCasesInSwitch": true
     },
     "include": ["src"],
     "references": [{ "path": "./tsconfig.node.json" }]
   }
   ```

This setup will automatically run:
Backend (Go):
- Code linting with golangci-lint
- Code formatting checks
- Unit tests

Frontend (TypeScript/React):
- TypeScript type checking
- ESLint for code quality
- Unit tests
- Build verification

The checks will run on:
- Every push to main branch
- Every pull request to main branch
- Only when relevant files are changed (Go files trigger backend checks, frontend files trigger frontend checks)

### Local Development Tools
While CI checks happen in the cloud, you may still want these tools locally for development:

1. Backend (Go):
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   # Format code
   gofmt -w .
   ```

2. Frontend (TypeScript):
   ```bash
   # In frontend directory
   npm install
   # Type checking
   npm run type-check
   # Lint
   npm run lint
   # Format code
   npm run format
   ```

These local tools are optional but can help catch issues before pushing code.

## Verification

After completing all installations, verify the development environment:

1. Check all version numbers:
   ```bash
   go version
   node --version
   docker --version
   docker compose version
   mongosh --version
   redis-cli --version
   pre-commit --version
   ```

2. Test Docker Compose:
   ```bash
   docker compose up -d
   docker compose ps
   docker compose down
   ```

## Troubleshooting

### Common Issues and Solutions

1. Docker permission issues:
   ```bash
   sudo groupadd docker
   sudo usermod -aG docker $USER
   ```

2. MongoDB connection issues:
   - Check if service is running
   - Verify port 27017 is not in use
   - Check firewall settings

3. Redis connection issues:
   - Check if service is running
   - Verify port 6379 is not in use
   - Check firewall settings

### Support Resources
- Go Documentation: https://golang.org/doc/
- Node.js Documentation: https://nodejs.org/docs/
- Docker Documentation: https://docs.docker.com/
- MongoDB Documentation: https://docs.mongodb.com/
- Redis Documentation: https://redis.io/documentation
