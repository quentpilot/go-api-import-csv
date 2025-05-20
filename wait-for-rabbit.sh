#!/bin/sh
set -e

host="$1"
port="$2"

echo "⏳ Attente que RabbitMQ soit prêt sur $host:$port..."

while ! nc -z "$host" "$port"; do
  sleep 1
done

echo "✅ RabbitMQ est prêt sur $host:$port"

shift 2
exec "$@"