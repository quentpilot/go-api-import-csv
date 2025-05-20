#!/bin/sh
set -e

host="$1"
port="$2"

echo "⏳ Waiting for RabbitMQ starts on $host:$port..."

while ! nc -z "$host" "$port"; do
  sleep 1
done

echo "✅ RabbitMQ is ready on $host:$port"

shift 2
exec "$@"