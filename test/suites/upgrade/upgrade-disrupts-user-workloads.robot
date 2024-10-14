*** Settings ***
Documentation   foo

Resource    ../../resources/oc.resource
Resource    ../../resources/kubeconfig.resource
Resource    ../../resources/common.resource
Resource    ../../resources/microshift-process.resource
Resource    ../../resources/ostree.resource

Library    SSHLibrary

Suite Setup    Setup Suite With Namespace
Suite Teardown    Teardown Suite With Namespace


*** Variables ***
${TARGET_REF}             ${EMPTY}
${WORKLOAD}     assets/hello-microshift.yaml


*** Test Cases ***
Verify User Workloads With Shared MicroShift Layer
    [Documentation]    ....
    [Setup]    Setup
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}
    Expect Crio Journal To Contain Error About Missing Lower Layer
    [Teardown]  Teardown


*** Keywords ***
Setup
    [Documentation]    do a setup
    Oc Apply    -n ${NAMESPACE} -f ${WORKLOAD}
    Named Pod Should Be Ready    hello-microshift

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Expect Crio Journal To Contain Error About Missing Lower Layer
    [Documentation]     .../
    ${rc}   ${stdout}    Execute Command
    ...    journalctl --boot crio | grep stat lower layer'
    ...    return_stdout=True   return_rc=True
    Should Be Equal As Integers    ${rc}    0
    Log    ${stdout}
