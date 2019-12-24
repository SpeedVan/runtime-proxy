#!/bin/sh

PROXY_EVENTSTORE_ENDPOINT='tcp://admin:changeit@10.10.139.35:1113' \
PROXY_WEBAPP_LISTEN_ADDRESS=':9999' \
PROXY_PORT=8888 PROXY_FORK_COMMAND='/usr/local/bin/python3' \
PROXY_FORK_COMMAND_ARG='/Users/finup/project/gitlab.puhuitech.cn/finup-faas/function-runtime-app/app.py' \
./build/runtime-proxy

