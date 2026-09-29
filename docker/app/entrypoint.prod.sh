#!/bin/sh
set -eu

APP_USER=solvi
APP_GROUP=solvi

for dir in /app/uploads /app/logs; do
    mkdir -p "$dir"
    chown -R "${APP_USER}:${APP_GROUP}" "$dir"
done

exec su-exec "${APP_USER}:${APP_GROUP}" "$@"
