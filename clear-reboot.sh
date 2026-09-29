#!/bin/sh
sudo docker compose -f docker-compose.prod.yml down --volumes
sudo docker compose -f docker-compose.prod.yml up -d --force-recreate