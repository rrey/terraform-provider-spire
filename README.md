# Terraform Provider SPIRE

This provider allows to interact with [SPIRE](https://spiffe.io/docs/latest/deploying/registering/) with Terraform.

## Run acceptance tests

**Warning**: A SPIRE Server must be started locally to run the tests since the unix socket must be reachable.
You can download the binaries and start the server with the following command:
```bash
./bin/spire-server run
```

To run the acceptance tests:

```bash
make testacc
```

## Auto generate documentation

```bash
make gendoc
```

All changes made on files after this command execution shall be committed.

## Limitations

* At the moment, the provider communicate with SPIRE through the local unix socket. It won't work with a real remote SPIRE server.
