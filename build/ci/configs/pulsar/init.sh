#!/bin/bash

# List tenants
pulsar-admin tenants list

# Create tenant
pulsar-admin tenants create --allowed-clusters standalone vigie

# Create namespace
pulsar-admin namespaces create vigie/worker

# Create partitioned topics
pulsar-admin topics create-partitioned-topic vigie/worker/test --partitions 1
pulsar-admin topics create-partitioned-topic vigie/worker/v0 --partitions 1
