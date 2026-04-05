#!/bin/bash
set -e

DB_USER=${DB_USER:-root}
DB_HOST=${DB_HOST:-127.0.0.1}
DB_PORT=${DB_PORT:-3306}

echo "Setting up database..."
mysql -u "$DB_USER" -p -h "$DB_HOST" -P "$DB_PORT" <<EOF
CREATE DATABASE IF NOT EXISTS ticketing;
USE ticketing;
CREATE TABLE IF NOT EXISTS tickets (
    id        INT     PRIMARY KEY,
    end_range BIGINT  NOT NULL DEFAULT 0
);
INSERT IGNORE INTO tickets (id, end_range) VALUES (1, 0);
EOF

echo "Done. Set your DB_DSN env var:"
echo "  export DB_DSN=\"\$DB_USER:<password>@tcp(\$DB_HOST:\$DB_PORT)/ticketing\""
