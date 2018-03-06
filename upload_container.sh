#!/bin/bash
docker login -u "$DOCKER_USER" -p "$DOCKER_PASSWORD";
docker build -t dockercliqz/cloudwatch-writer:latest .
docker push dockercliqz/cloudwatch-writer:latest

