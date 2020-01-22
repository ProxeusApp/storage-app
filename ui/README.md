# Proxeus UI

### Prerequisites
+ yarn (1.12.3+)
+ node (8.11.3+)
+ vue-cli

## Important
+ Only use yarn 1.12.3. For linking together local dependencies we use Yarn Workspaces:
https://yarnpkg.com/lang/en/docs/workspaces/
+ Use yarn only from the /core/central/ui directory, as the common dependencies will be stored in
/core/central/ui/node_modules instead of the package subfolders (./core, ./dapp etc.).

## Getting Started

Install dependencies and setup yarn workspaces:
```
make init
```

### Development
To start the local dApp development server:
```
make serve-dapp
```

Same for the main hosted app:
```
make serve-main-hosted
```

### Building frontend components
All the build projects are stated in `./Makefile`. To build all execute:
```
make all
```

Each build generates a ``dist`` directory. The dApp build places it directly in the /core/central/dapp dir so
go-bindata can access it directly.
