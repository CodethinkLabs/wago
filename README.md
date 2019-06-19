# WalletGo

Wallet go is a distributed consensus client for trading arbitrary
currencies between various users via the [raft consensus algorithm](https://raft.github.io/raft.pdf).
It is not secure (at all) against malicious actors and should 
not be used by anyone.

### Usage:

There are a few ways to run the program:

```bash
# using go directly
go run cmd/server/main.go

# using bazel
bazel run //cmd/server:main

# using docker
docker run -it arlyon/wago:latest
```

At this point you will be greeted by a menu. Type `new` to create a new
wallet. The next useful command is `create`. This will generate currency
for you and is of the format:

```bash
create ${PUBKEY} 200 usd
```

This will send a command to the cluster to create currency. The `bank` command 
will display your addresses and how much you have on hand.
You can send currency, but it may be worth creating another wallet first.
You'll notice the autofill only suggests addresses that have money in them,
so go ahead and type `bank full` to get the public keys in your wallet.
The `bank` command (and most of the UI) usually only displays a shortened 
identifier.

```bash
> new
    Generating new key: da6cef7a9b71
> bank full
    Public key 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea:
      - usd: 200.0
    Public key da6cef7a9b71ec27f0f75a5d7914fa7c282218c8488c44030355690e67bd5fa9: no currency
> send 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea da6cef7a9b71ec27f0f75a5d7914fa7c282218c8488c44030355690e67bd5fa9 50 usd
    You got money! Transfer of {50 0} usd successfully received from 3916b30c33c07d766f3deab1ff1cf0d9949925a1dfc277bb6a066be0c0c867ea
> bank
    Public key 3916b30c33c0:
      - usd: 150.0
    Public key da6cef7a9b71:
      - usd: 50.0
```

You can also delete keys if you no longer want to use them:

```bash
> delete d
Removed key dda6cef7a9b71
```

You'll notice there is autofill going on here. In cases where wago
can determine precisely which wallet you are trying to transfer to,
it will autocomplete the addresses.

### Authentication

A password is needed for encrypting you wallet. Set it (per session) by using the
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

### Cryptography

This project uses cryptography to secure the system (on a surface level).
Currency is organized into wallets which is a ED25519 key pair. ED25519
is a digital signature scheme which is what proves to the system that a
transaction truly comes from the source key. When a transaction is sent,
the sender (source address) signs the concatenation of the 
`SRC_PUB^DST_PUB^AMOUNT^CURR`. The leader of the cluster, when updating
the state machine, verifies this signature before authorizing the 
transaction.

The system is vulnerable to Byzantine faults meaning it has no protection
against malicious or malfunctioning nodes. Raft, the underlying consensus
algorithm, only guarantees correctness when all the nodes in the cluster
are non-Byzantine. This is, for example, observed in the log replication.
In Raft, the leader has complete responsibility for ensuring logs are
properly replicated. A malicious node could, for example, intercept
proposals and while, due to the cryptographic signature, client nodes
are able to reject invalid transactions and elect a new leader, there
is no protection against starving the follower nodes of updates.