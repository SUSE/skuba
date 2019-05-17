# End to End tests for caaspctl

# Env. Variable:

## Host related
CONTROLPLANE = IP of host which will be the controlplane
MASTER00 = IP of 1st master

## Binary Localtion (optional)

Use `CAASPCTL-BIN` for specify the full path of a caaspctl binary. ( e.g if you use an RPM)

By default this variable point to GOPATH.



# Run e2e-tests

`CONTROLPLANE=10.86.2.23 MASTER00=10.86.1.103 WORKER00=10.86.2.130 make test-e2e` 

have a look at `ci/tasks/e2e-tests.py` for supported var

### Libvirt
`make test-e2e`
