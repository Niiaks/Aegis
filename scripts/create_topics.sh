#!/bin/bash
# Create Kafka topics
docker exec broker kafka-topics.sh --create --topic aegis.payment.created --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.ledger.entries --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.balance.update --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.webhook.pending --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.payout.pending --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.payout.status.update --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.reconciliation.job --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.discrepancy.detected --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists
docker exec broker kafka-topics.sh --create --topic aegis.dlq --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1 --if-not-exists

echo "Kafka topics created successfully."
