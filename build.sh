#!/bin/sh

echo "[*] Building plugins"
python3 init_plugins.py --force

echo "[*] Building progoxy"
go build

if [ "$1" = "--compress" ] && [ -x `which upx` ]
then
    echo "[*] Compressing binary (this can take awhile)"
    upx --ultra-brute ./progoxy
fi