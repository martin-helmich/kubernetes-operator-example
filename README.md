# "HelloWorld" Kubernetes Operator example

This is an example Kubernetes Operator, using the [Operator Framework](https://github.com/operator-framework).

## Building

To build this project, you'll need the [Operator SDK](https://github.com/operator-framework/operator-sdk) installed on your local machine.

With the SDK installed, build the operator as follows:

```console
> operator-sdk build $DOCKER_IMAGE
```

## Running locally

Make sure you have a local Kubernetes cluster running for this (or a remote one suitable for testing purposes).

Then execute:

```console
> operator-sdk up local --namespace=$NAMESPACE  # default ns is "default"
```

## Installing in a cluster

### Manually

1. Start by creating the CRD:

    ```console
    > kubectl apply -f deploy/crds/example_v1alpha1_helloworld_crd.yaml
    ```

2. Then, deploy the operator:

    ```
    > kubectl apply -f deploy
    ```

3. Finish off by creating a `HelloWorld` custom resource:

    ```console
    > kubectl apply -f deploy/crds/example_v1alpha1_helloworld_cr.yaml

### Using the Operator Lifecycle Manager

Coming soon