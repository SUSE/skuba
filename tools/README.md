# Caaspctl Tools pkg

Why is this needed?

We need this pkg tools for making working cmd-line vendored pkgs, such `ginkgo`.
In order to make the vendoring work for the cmd, we import those pkgs here, since they are not used in the code anyhow.
