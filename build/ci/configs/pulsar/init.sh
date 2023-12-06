./pulsar-admin tenants list


./pulsar-admin tenants create vigie
./pulsar-admin namespaces create vigie/worker
./pulsar-admin topics create-partitioned-topic vigie/worker/test -p 1

./pulsar-admin topics create-partitioned-topic vigie/worker/v0 -p 1