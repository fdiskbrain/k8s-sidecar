#!/usr/bin/env bash
wget https://gitlab.com/gitlab-org/cli/-/releases/v1.95.0/downloads/glab_1.95.0_linux_$( go env GOARCH).deb -O /tmp/glab.deb 
sudo dpkg -i /tmp/glab.deb