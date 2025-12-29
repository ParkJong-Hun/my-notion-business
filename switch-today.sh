#!/bin/bash

today=$(date +%Y%m%d)

# 명령 실행
git switch main
git add .
git reset --hard HEAD
git pull origin main
git switch -c "feature/$today"
