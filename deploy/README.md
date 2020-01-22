# Proxeus Deployments
This repository contains scripts for automated deployement agents (like bamboo).

## Structure
1. Each service has it's own directory
2. Each service has it's own docker-compose so they can be deployed on remote machines in any combination.

## Procedure
1. Agent downloads and copies desired subset of artifacts directory  into remote machine
2. Agent copies deployment scripts into remote machine
3. Agent causes make script execution (e.g. through ssh)