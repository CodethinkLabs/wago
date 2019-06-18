# WalletGo

Wallet go is a distributed consensus client for trading arbitrary
currencies between various users via the raft consensus algorithm.
It is not vetted or secure (at all) from bad actors and should not be
used in any trust-less system.

### Usage:

```bash
bazel build //:gazelle
bazel run //cmd/server:main
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
./wago --id 2 --cluster http://127.0.0.1:19200,http://127.0.0.1:19201 --join
```