#!/bin/bash

set -e

dnf upgrade --refresh -y

dnf install -y epel-release
dnf install -y wget ca-certificates tmux vim tcpdump chrony sos htop
systemctl enable --now chronyd


dnf clean -y all
dnf install -y https://yum.theforeman.org/releases/3.11/el9/x86_64/foreman-release.rpm
dnf install -y https://yum.theforeman.org/katello/4.13/katello/el9/x86_64/katello-repos-latest.rpm
dnf install -y https://yum.puppet.com/puppet7-release-el-9.noarch.rpm

# verification
dnf repolist enabled

# Installation
dnf upgrade -y
dnf install -y foreman-installer-katello


# NOTICE INFO
foreman-installer \
    --scenario katello \
    -l NOTICE \
    --foreman-initial-organization "E2E-Company" \
    --foreman-initial-location "E2E-Location" \
    --foreman-initial-admin-username "admin" \
    --foreman-initial-admin-password "admine2e"
