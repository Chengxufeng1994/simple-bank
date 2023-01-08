#!/bin/sh

set -e

echo "run db migration"
/usr/app/migrate -path=/migrations -database "postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST:$DB_PORT/$POSTGRES_DATABASE?sslmode=disable" -verbose up

echo "start the app"
exec "$@"