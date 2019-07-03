# wago

Wallet go is a distributed consensus server for trading arbitrary
currencies between various users via the [raft consensus algorithm](https://raft.github.io/raft.pdf).

### Why?

This is, in short, an example project to explore a number of technologies:

##### Bazel, rules_docker, rules_k8s

Part of this project was to become more familiar with build systems. The goal
was to be able to distill an otherwise complex deployment process into a
something as digestible as possible. Bazel with `rules_docker` and `rules_k8s`
is an effective set of tools. The first "layer" is having bazel build the 
project hermetically.

```bash
bazel build cmd/server:binary
```

Next up from that, `rules_docker` allows trivial creation of docker images from
bazel projects. Building bazel projects in docker the "traditional" way (with
a Dockerfile) has, in its default config, docker build the project inside the
container. This means all the benefits of build caching is wasted because it
is erased every time. Additionally, the image gets bloated. Java, bazel, and
the cache all must be available in the container.

Migrating to rules_docker brought the bloated ubuntu-based image from 1.6GB
down to a lean 14MB as well as dramatically improving incremental build times
(15 minutes down to a few seconds). It also allowed much simpler pushes to 
docker hub.

```bash
# build docker image
bazel run cmd/server:image

# alternatively upload directly
bazel run cmd/server:push
```

> Note: these commands do not have to be run sequentially as above. 
> Push will build the image if changes are made.

The last stage leverages `rules_k8s` to orchestrate entire deployments. It relies
on both of the previous stages and will push new docker images as required when
various commands are run.

```bash
bazel run kubernetes:deploy.create
bazel run kubernetes:deploy.delete
```

##### Remote Execution (tba)

##### Raft Consensus Algorithm

Raft allows many nodes to maintain consensus over state. This is essentially the
reasoning for the project; the actual outcome (consensus about money) is just an 
interesting way to measure the problem with a simple success criterion. 
If all the submitted transactions are applied in a predictable order, and users 
are unable to double spend, the system is correct.

Using raft in a kubernetes cluster in the way we have configured it allows for
impressive availability with minimal effort. Any minority of nodes can go down
at once, and the cluster will still correctly recover and propagate updates to
the nodes that went down.

This property is desirable for any single point of failure. Raft does not 
distribute work across nodes; at any given point there is only one node that has
permission to edit the state machine. For this reason it is only able to boost
availability and is great in cases where a single point of failure is likely.
Due to our findings, we are going to apply this distributed consensus to the
scheduler in build barn to make it crash-resistant. 

##### Protobuf + GRPC

Bazel uses protocol buffers to define its remote execution API. It makes sense,
seeing as this is a technology test for bazel, that this project integrate them 
somewhere as well. To facilitate communication between the client and server
nodes, a proto file is used to define a number of well-known operations which
can be used to generate client and server code.

The server implementation is in `/cmd/server/http.go`.

##### Go

The last, and least important, reason is to learn go and provide an idiomatic
project using bazel.

### Getting Started

There are a few ways to run the program. The first is to run a server directly.
This allows you to build ad-hoc clusters of "server clients" which both host and
interact with the system.

```bash
# build with bazel
bazel run cmd/server:binary --help

# pull the docker image
docker run -it arlyon/wago:latest --help
```

The second is by running a cluster of server nodes that host the system and 
provide access to the system via grpc calls. This is where the client comes in.

```bash
# deploy kubernetes cluster for the server
bazel run kubernetes:deploy.apply

# build with bazel
bazel run cmd/client:binary --cluster=${KUBERNETES_URL}
```

In this configuration, the server cluster maintains the state of the system,
and the clients are able to submit transactions to be applied on their behalf.
The command line usage on the client and server machines are (mostly) the same.

### Usage

At this point you will be greeted by a menu. Type `new` to create a new
wallet. The next useful command is `create`. This will generate currency
for you and is of the format:

```bash
# on the server
create ${PUBKEY} 200 usd

# on the client
create ${PUBKEY} 200 usd [PASSWORD]
```

This will send a command to the cluster to create currency. The `bank` command 
will display your addresses and how much you have on hand. Most of the UI displays 
only a shortened identifier instead of the full key (like a git commit).
You can send currency, but it may be worth creating another wallet first.

```bash
$ new
Generating new key: da6cef7a9b71
$ bank full
Public key 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea:
  - usd: 200.0
Public key da6cef7a9b71ec27f0f75a5d7914fa7c282218c8488c44030355690e67bd5fa9: no currency
$ send 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea da6cef7a9b71ec27f0f75a5d7914fa7c282218c8488c44030355690e67bd5fa9 50 usd
You got money! Transfer of {50 0} usd successfully received from 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea
$ bank
Public key 3916b30c33c0:
  - usd: 150.0
Public key da6cef7a9b71:
  - usd: 50.0
```

You can also delete keys if you no longer want to use them:

```bash
$ delete d
Removed key dda6cef7a9b71
```

You'll notice there is autofill going on here. In cases where wago
can determine precisely what target you are sending currency to
it will auto-complete it.

### Authentication

A password is needed for encrypting your wallet. Set it (per session) by using the
authenticate command. If you use no password when creating the wallet, encryption
will be disabled.

### Multiple clients

The system is obviously not much use when only a single user
is active. To add more nodes to the cluster, we need some additional
arguments. 

| Flag | Description |
| ---- | ----------- |
| `--cluster` | A comma separated list of peer IP addresses
| `--id`      |      This node's index in the list of peers
| `--join`    |    Whether this node is joining an existing cluster

The cluster string should be identical between all nodes.
Because the ID has to be unique between nodes, we can use that
to assign the addresses. ID 1 takes the first IP address and so on.

An example config with 2 nodes:

```bash
./wago --id 1 --cluster http://127.0.0.1:19200,http://127.0.0.1:19201
./wago --id 2 --cluster http://127.0.0.1:19200,http://127.0.0.1:19201
```

There is also another feature. In cases where the id is omitted, wago
will attempt to infer the ID from hostname of the server.

### Cryptography

This project uses cryptography to secure the system (on a surface level).
Currency is organized into wallets which is a ED25519 key pair. ED25519
is a digital signature scheme which is what proves to the system that a
transaction truly comes from the source key. When a transaction is sent,
the sender (source address) signs the concatenation of the 
`SRC_PUB^DST_PUB^AMOUNT^CURR`. The leader of the cluster, when updating
the state machine, verifies this signature before authorizing the 
transaction.

The server is vulnerable to Byzantine faults meaning it has no protection
against malicious or malfunctioning nodes. Raft, the underlying consensus
algorithm, only guarantees correctness when all the nodes in the cluster
are non-Byzantine. This is, for example, observed in the log replication.
In Raft, the leader has complete responsibility for ensuring logs are
properly replicated. A malicious node could, for example, intercept
proposals and while, due to the cryptographic signature, client nodes
are able to reject invalid transactions and elect a new leader, there
is no protection against starving the follower nodes of updates.

If you control the servers, then clients can connect and use the system
without any security issues (or guarantees!).