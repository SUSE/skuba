# Skuba Continous Integration , Deployment and Infrastructure.

## Infrastructure deployments

[infra](infra/README.md)

## Jenkins
[jenkins jobs](jenkins/README.md)

## Makefile

The Makefile has targets for our CI to use for the different steps in the pipelines.

Right now it consists of Stages which are intended run the other targets in order and should not change often. 
The other targets are individual Steps which can be called but are more likely to be changed or removed.
