#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_ref=rhel-9.4-microshift-source-isolated
target_ref=rhel-9.4-source-isolated-busybox-upgrade

scenario_create_vms() {
    if ! does_commit_exist "${start_ref}"; then
        echo "Commit '${start_ref}' not found in ostree repo - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart.ks.template "${start_ref}"
    launch_vm
}

scenario_remove_vms() {
    if ! does_commit_exist "${start_ref}"; then
        echo "Commit '${start_ref}' not found in ostree repo - skipping test"
        return 0
    fi
    remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${start_ref}"; then
        echo "Commit '${start_ref}' not found in ostree repo - skipping test"
        return 0
    fi
    run_tests host1 \
        --variable "TARGET_REF:${target_ref}" \
        suites/upgrade/upgrade-disrupts-user-workloads.robot
}
