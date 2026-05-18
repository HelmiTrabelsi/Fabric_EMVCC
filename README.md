# Fabric_EMVCC

**Early MVCC conflict detection for Hyperledger Fabric.**

This repository is the research prototype that accompanies our paper *Early Detection for Multiversion Concurrency Control Conflicts in Hyperledger Fabric* (Trabelsi and Zhang, [arXiv:2301.06181](https://arxiv.org/abs/2301.06181), 2023). It is a fork of [Hyperledger Fabric v2.x](https://github.com/hyperledger/fabric) that moves MVCC conflict detection out of the commit phase and into the endorsement phase, so transactions that are bound to fail validation get rejected before they ever reach the ordering service.

The upstream Fabric codebase is preserved as-is. The EMVCC changes are small and localized; see [Modified packages](#modified-packages) if you want to read the diff.

## Table of contents

* [Why this exists](#why-this-exists)
* [How EMVCC works](#how-emvcc-works)
* [Repository layout](#repository-layout)
* [Modified packages](#modified-packages)
* [Building](#building)
* [Trying it out with EMVCC_Network](#trying-it-out-with-emvcc_network)
* [Reproducing the conflict scenario](#reproducing-the-conflict-scenario)
* [What this prototype does not do](#what-this-prototype-does-not-do)
* [Paper and thesis](#paper-and-thesis)
* [License](#license)

## Why this exists

Hyperledger Fabric runs transactions in three phases: execute, order, validate.

1. **Execute (endorsement).** A client sends a proposal to one or more endorsing peers. Each peer simulates the chaincode against its current world state and returns a signed read-write set: the keys read (with their versions) and the writes the transaction would apply.
2. **Order.** The client gathers enough endorsements and submits the transaction to the ordering service, which bundles transactions into blocks.
3. **Validate (commit).** Every committing peer rechecks each transaction in the block. The MVCC check compares the versions in the read set against the world state at commit time. If any key changed since simulation, the transaction is marked `MVCC_READ_CONFLICT` and its writes are dropped. The transaction still occupies a slot in the block.

That design is what gives Fabric its parallel execution and its tolerance for non-deterministic chaincode, but it has a known cost: conflicts are only caught at the very end of the pipeline.

In practice that means:

* A doomed transaction still gets simulated, endorsed, signed, gossiped, ordered, and written to a block before it gets thrown away.
* Workloads with hot keys (counters, registries, balance accounts) can see large fractions of a block invalidated, especially as throughput goes up.
* Clients only learn the outcome after the block is committed, well after they could have done anything about it.
* A correct-at-the-time endorsement can become invalid before it even reaches the orderer if a concurrent transaction touches the same key first.

## How EMVCC works

EMVCC adds a conflict-detection step inside the endorsing peer, before the endorsement signature is returned to the client. The intuition is simple: a peer already maintains a recent view of the world state and a queue of read-write sets it has just endorsed, so it has enough information to spot transactions that are certain to conflict at commit time.

The modified endorsement path does this:

1. Simulates the chaincode as in stock Fabric and produces the candidate read-write set.
2. Checks every key in the read set against the current committed version in the local state DB, and against pending writes from read-write sets this peer recently endorsed.
3. If a conflict is found, the peer returns an `EMVCC_CONFLICT` response to the client instead of an endorsement, naming the offending key(s). The client can back off, re-read, and retry immediately. The proposal never enters the ordering pipeline.
4. If no conflict is found, the endorsement proceeds normally. The upstream commit-time MVCC check still runs as a final correctness gate.

```
