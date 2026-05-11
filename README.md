```
project
├─ cmd
│  └─ api
├─ Dockerfile
├─ internal
│  ├─ app
│  ├─ config
│  ├─ entity
│  │  ├─ document
│  │  │  ├─ assembly
│  │  │  ├─ incoming
│  │  │  └─ outgoing
│  │  ├─ event
│  │  └─ report
│  ├─ event
│  │  └─ subscriber
│  ├─ handlers
│  │  ├─ alert
│  │  ├─ assembly
│  │  ├─ assembly_order
│  │  ├─ batch
│  │  ├─ counterparty
│  │  ├─ incoming
│  │  ├─ movement
│  │  ├─ outgoing
│  │  ├─ product
│  │  ├─ report
│  │  ├─ stock
│  │  └─ warehouses
│  ├─ repository
│  │  └─ postgres
│  │     ├─ alert
│  │     ├─ assembly
│  │     ├─ assemblyorder
│  │     ├─ batch
│  │     ├─ batch_movement
│  │     ├─ counterparty
│  │     ├─ incoming
│  │     ├─ movement
│  │     ├─ outgoing
│  │     ├─ product
│  │     ├─ stock
│  │     └─ warehouse
│  ├─ service
│  │  ├─ alert
│  │  ├─ assembly
│  │  ├─ assemblyorder
│  │  ├─ batch
│  │  ├─ counterparty
│  │  ├─ incoming
│  │  ├─ outgoing
│  │  ├─ product
│  │  ├─ report
│  │  │  └─ cogs
│  │  ├─ stock
│  │  ├─ stockmovement
│  │  └─ warehouse
│  └─ storage
│     └─ postgres
├─ Makefile
└─ migrations

```