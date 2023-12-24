# Flex AWS Topology

> Explore using the AWS topology API

We want to eventually use the [AWS Topology API](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-topology.html) to add metadata to fluence about a cluster topology.
To do this we can use the [Go bindings](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstanceTopology).

## Design

### Overview

The Flux Framework "flux-sched" or fluxion project provides modular bindings in different languages for intelligent,
graph-based scheduling. When we extend fluxion to a tool or project that warrants logic of this type, we call this a flex!
Thus, the project here demonstrates flex-archspec, or using fluxion to match some system request to what we know in the archspec graph. E.g.,

> Do you have an x86 system with this compiler option?

This is a simple use case that doesn't perfectly reflect the OCI container use case, but we need to start somewhere! For this very basic setup we are going to:

1. Load the machines into a JSON Graph (called JGF).
2. Try doing a query against system metadata

There will eventually be a third component - a container image specification, for which we need to include somewhere here. I am starting simple!

### Concepts

From the above, the following definitions might be useful.

 - **[Flux Framework](https://flux-framework.org)**: a modular framework for putting together a workload manager. It is traditionally for HPC, but components have been used in other places (e.g., here, Kubernetes, etc). It is analogous to Kubernetes in that it is modular and used for running batch workloads.
 - **[fluxion](fluxion)**: refers to [flux-framework/flux-sched](https://github.com/flux-framework/flux-sched) and is the scheduler component or module of Flux Framework. There are bindings in several languages, and specifically the Go bindings (server at [flux-framework/flux-k8s](https://github.com/flux-framework/flux-k8s)) assemble into the project "fluence."
 - **flex** is an out of tree tool, plugin, or similar that uses fluxion to flexibly schedule or match some kind of graph-based resources. This project is an example of a flex!

## Usage

### Build

This demonstrates how to build the bindings. You will need to be in the VSCode developer container environment, or produce the same
on your host. Note that we currently are using [this commit](https://github.com/researchapps/flux-sched/commit/86f5bb331342f2883b057920cf58e2c042aef881) that
si a fork of [milroy's work](https://github.com/flux-framework/flux-sched/pull/1120) to ensure the module name matches what is added to go.mod (it won't work otherwise). When this is merged, we will update to flux-framework/flux-sched. Below shows the make command that builds our final binary!

```bash
make
```
```console
# This needs to match the flux-sched install and latest commit, for now we are using a fork of milroy's branch
# that has a go.mod updated to match the org name
# go get -u github.com/researchapps/flux-sched/resource/reapi/bindings/go/src/fluxcli@86f5bb331342f2883b057920cf58e2c042aef881
go mod tidy
mkdir -p ./bin
GOOS=linux CGO_CFLAGS="-I/opt/flux-sched/resource/reapi/bindings/c" CGO_LDFLAGS="-L/usr/lib -L/opt/flux-sched/resource -lfluxion-resource -L/opt/flux-sched/resource/libjobspec -ljobspec_conv -L//opt/flux-sched/resource/reapi/bindings -lreapi_cli -lflux-idset -lstdc++ -lczmq -ljansson -lhwloc -lboost_system -lflux-hostlist -lboost_graph -lyaml-cpp" go build -ldflags '-w' -o bin/aws-topology src/cmd/main.go
```

The output is generated in bin:

```bash
$ ls bin/
aws-topology
```

### Run

#### 1. Paths

Ensure you have your LD library path set to find flux sched (fluxion) libraries.

```bash
export LD_LIBRARY_PATH=/usr/lib:/opt/flux-sched/resource:/opt/flux-sched/resource/reapi/bindings:/opt/flux-sched/resource/libjobspec
```

#### 2. Credentials

Ensure you have AWS credentials in your environment (I figure any cloud scheduler we use is going to have environment ephemeral secrets and not other kinds of credentials, but we can change this).

```bash
export AWS_ACCESS_KEY_ID=xxxxxxxxxxxxx
export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxx
export AWS_SESSION_TOKEN=xxxxxxxxxxxxx
```

#### 4. Nodes

You can either manually create node (be sure to choose an hpc instance family type) or create a mini cluster with the config provided here:

```bash
eksctl create cluster --config-file eksctl-config.yaml 
```

That will create a cluster with 2 nodes in a placement group in us-east-2. It takes a while.

#### 5. Run

Here is how to run with an instance id, and note that the instance is running:

```bash
./bin/aws-topology --instance i-0cd50e305e797dde3
```

And here is how to run with a placement group (per the creation above):

```bash
./bin/aws-topology --group eks-efa-testing --region us-east-2
```

This will generate the JGF to a non-temporary file for you to debug:

```bash
./bin/aws-topology --group eks-efa-testing --region us-east-2 --file ./aws-topology.json
```

This shows simple output and you can also view the [generated topology](aws-topology.json):

<details>

<summary> Output for JGF </summary>

```console
This is the flex aws topology prototype
 Match policy: first
 Load format: JSON Graph Format (JGF)
Created flex resource graph &{%!s(*fluxcli.ReapiCtx=&{})}
Topology Query Parameters:
{
  DryRun: false,
  GroupNames: ["eks-efa-testing"]
}
{
  Instances: [{
      AvailabilityZone: "us-east-2b",
      GroupName: "eks-efa-testing",
      InstanceId: "i-02125af4faf797399",
      InstanceType: "hpc6a.48xlarge",
      NetworkNodes: ["nn-ec17a935b39a06f41","nn-dd9ec3119ca6ea9dc","nn-a59759166e67e7c02"],
      ZoneId: "use2-az2"
    },{
      AvailabilityZone: "us-east-2b",
      GroupName: "eks-efa-testing",
      InstanceId: "i-0fbbd476a798a3f82",
      InstanceType: "hpc6a.48xlarge",
      NetworkNodes: ["nn-ec17a935b39a06f41","nn-dd9ec3119ca6ea9dc","nn-a59759166e67e7c02"],
      ZoneId: "use2-az2"
    }],
  NextToken: "..."
}
i-02125af4faf797399 is not yet seen, adding with uid 1
nn-ec17a935b39a06f41 is not yet seen, adding with uid 2
nn-dd9ec3119ca6ea9dc is not yet seen, adding with uid 3
nn-a59759166e67e7c02 is not yet seen, adding with uid 4
Creating instance node for i-02125af4faf797399
Creating network node for nn-ec17a935b39a06f41
Creating network node for nn-dd9ec3119ca6ea9dc
Creating network node for nn-a59759166e67e7c02
i-0fbbd476a798a3f82 is not yet seen, adding with uid 5
Creating instance node for i-0fbbd476a798a3f82
Creating node 0 cluster
Creating node 1 i-02125af4faf797399
Creating node 2 nn-ec17a935b39a06f41
Creating node 3 nn-dd9ec3119ca6ea9dc
Creating node 4 nn-a59759166e67e7c02
Creating node 5 i-0fbbd476a798a3f82
Creating edge (4 contains->5) (5 in-> 4) 
Creating edge (0 contains->2) (2 in-> 0) 
Creating edge (2 contains->3) (3 in-> 2) 
Creating edge (3 contains->4) (4 in-> 3) 
Creating edge (4 contains->1) (1 in-> 4) 
```

</details>

Note that we could next add some kind of match - I'm guessing we care about distances in the graph more than attributes.
I'll wait to chat with folks more about next steps, because I've accomplished the goal I set out to do.
This was immensely satisfying to work on.

#### 6. Cleanup

Don't forget to cleanup your nodes - they cost money!

```bash
eksctl delete cluster --config-file eksctl-config.yaml 
```

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
