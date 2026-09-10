FROM amazon/dynamodb-local:latest
LABEL MAINTAINER="ryota.kashio"

# 一時的に root ユーザーに切り替えてディレクトリ作成と権限変更を行う
USER root

# ディレクトリを作成し、所有者を dynamodblocal (UID: 1000) に変更
RUN mkdir -p /home/dynamodblocal/data && \
    chown -R 1000:1000 /home/dynamodblocal/data

USER dynamodblocal

EXPOSE 8000
