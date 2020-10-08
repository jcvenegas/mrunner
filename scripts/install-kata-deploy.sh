#!/bin/bash

kata_version="1.12.0-alpha1"
sudo docker run -v /opt/kata:/opt/kata -v /var/run/dbus:/var/run/dbus -v /run/systemd:/run/systemd -v /etc/docker:/etc/docker -it katadocker/kata-deploy:"${kata_version}" kata-deploy-docker install
sudo docker info | grep Runtimes
