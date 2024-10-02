*** Settings ***
Library    OperatingSystem

Resource   ../resources/kubeconfig.resource
Resource   ../../common.sh

Setup Suite    Setup
Teardown Suite    Teardown

Test Tags    # ...


*** Variables ***
#${VAR1}    VALUE


*** Test Cases ***
MicroShift ...
    [Documentation]    # ...


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Login MicroShift Host1
    Initialize Global Variables
    Setup Suite With Namespace
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Initialize Global Variables
    [Documentation]    Initializes global variables.
    # ...