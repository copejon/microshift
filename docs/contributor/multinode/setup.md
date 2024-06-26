# Multinode Testing Environment for MicroShift
This document describes an opinionated, non-production setup of a two-node MicroShift
setup to facilitate running tests.

MicroShift does **not** support a multinode mode of operation, while some of
the tests require it to run successfully. The instructions in this document are
the recommended way to configure a non-production two-node setup to be used
**only** for running tests.

## Prerequisites
A physical `hypervisor host` with the [libvirt](https://libvirt.org/) virtualization
platform to be used for bootstrapping MicroShift primary and secondary hosts for
running tests.

Depending on the desired MicroShift installation method, follow the instructions
to set up one of the configurations for the primary and secondary hosts:
* [RHEL with MicroShift RPMs](./config_rpm.md)
* [MicroShift on RHEL for Edge](./config_r4e.md)

Log into the `hypervisor host` and set the environment variables that denote
primary and secondary host names and IP addresses to be used in the subsequent
commands.
```
PRI_HOST=microshift-pri
PRI_ADDR=192.168.122.118

SEC_HOST=microshift-sec
SEC_ADDR=192.168.122.70
```

## Configure Primary and Secondary Hosts
Run the following commands to copy the configuration script to the primary host
and run it remotely.
```
git clone https://github.com/openshift/microshift.git ~/microshift
cd ~/microshift/
scp -o StrictHostKeyChecking=no ./scripts/multinode/configure-pri.sh redhat@${PRI_ADDR}:
ssh redhat@${PRI_ADDR} ./configure-pri.sh "${PRI_HOST}" "${PRI_ADDR}" "${SEC_HOST}" "${SEC_ADDR}"
```

If the configuration script runs successfully, it prints the list of the
`kubelet` configuration files that need to be copied to the secondary host.

Copy the `kubelet` configuration files from the primary to the secondary host.
```
scp -3 -o StrictHostKeyChecking=no \
    redhat@${PRI_ADDR}:/home/redhat/kubelet-${SEC_HOST}.{key,crt} \
    redhat@${PRI_ADDR}:/home/redhat/kubeconfig-${PRI_HOST} \
    redhat@${SEC_ADDR}:
```

Run the following commands to copy the configuration script to the secondary host
and run it remotely.
```
cd ~/microshift/
scp -o StrictHostKeyChecking=no ./scripts/multinode/configure-sec.sh redhat@${SEC_ADDR}:
ssh redhat@${SEC_ADDR} ./configure-sec.sh "${PRI_HOST}" "${PRI_ADDR}" "${SEC_HOST}" "${SEC_ADDR}"
```

## Run Tests
Before running tests, make sure that the `microshift-pri` host name is resolved
and accessible from the `hypervisor host`.

Set the `KUBECONFIG` variable using the configuration file from the primary host.
```
export KUBECONFIG=$(mktemp /tmp/microshift-kubeconfig.XXXXXXXXXX)
scp redhat@${PRI_ADDR}:/home/redhat/kubeconfig-${PRI_HOST} ${KUBECONFIG}
```

Verify that the cluster has **two** nodes in the `Ready` status and wait until
all the pods are in the `Running` status.
```
oc get nodes
watch oc get pods -A
```

Run your test suite of choice using the primary and secondary MicroShift instances.

If you want to run [openshift-tests](https://github.com/openshift/origin) suite, follow [this](../openshift_ci.md#running-tests-manually) document.