#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel96-bootc-source
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    run_tests host1 suites/backup/auto-recovery.robot suites/backup/auto-recovery-extra.robot
}
