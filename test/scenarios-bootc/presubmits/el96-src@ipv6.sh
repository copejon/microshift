#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
# Using `hostname` here instead of a raw ip because skopeo only allows either
# ipv4 or fqdn's, but not ipv6. Since the registry is hosted on the ipv6
# network gateway in the host, we need to use a combination of the hostname
# plus /etc/hosts resolution (which is taken care of by kickstart).
# shellcheck disable=SC2034  # used elsewhere
MIRROR_REGISTRY_URL="$(hostname):${MIRROR_REGISTRY_PORT}/microshift"

scenario_create_vms() {
    # Enable IPv6 single stack in kickstart
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source false true
    launch_vm --boot_blueprint rhel96-bootc --network "${VM_IPV6_NETWORK}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/ipv6/singlestack.robot
}
