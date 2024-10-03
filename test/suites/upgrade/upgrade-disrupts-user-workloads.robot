*** Settings ***
Resource    ../../resources/kubeconfig.resource
Resource    ../../resources/common.resource


Suite Setup    Setup Suite
Suite Teardown    Teardown Suite


*** Variables ***
${TEST_IMAGE}=    ${EMPTY}


*** Test Cases ***
MicroShift ...
    [Documentation]    # ...
    #

*** Keywords ***
Setup
    [Documentation]    do a setup

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Initialize Global Variables
    [Documentation]    Initializes global variables.
    # ...