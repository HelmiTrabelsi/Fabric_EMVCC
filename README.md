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
       Stock Fabric                              EMVCC
       -----------                               -----
   Client → Endorse                          Client → Endorse
            ↓                                          ↓
        (simulate)                                (simulate)
            ↓                                          ↓
         endorsed                          ┌── EMVCC pre-check ──┐
            ↓                              │                     │
          Order                            ▼                     ▼
            ↓                          endorsed          EMVCC_CONFLICT
        Validate                           ↓             (early reject)
            ↓                            Order
   MVCC_READ_CONFLICT                       ↓
       at commit                        Validate
                                            ↓
                                  (rarely fires, already filtered)
```

The commit-time check is kept intact. EMVCC is an extra gate, not a replacement, so correctness is unchanged: the only transactions EMVCC can reject are transactions the existing validator would have rejected anyway. What changes is when, and how cheaply, that rejection happens.

## Repository layout

The top of the tree is a Hyperledger Fabric v2.x source tree. The two folders most relevant to this fork are:

| Path | What's in it |
|------|--------------|
| `EMVCC_Network/` | A self-contained test network (modeled on `fabric-samples/test-network`) for bringing up EMVCC-enabled peers and reproducing the conflict scenario. |
| `core/` | The Fabric peer core, including the endorsement and ledger code paths modified by EMVCC. |
| `docs/` | Upstream Fabric documentation, kept as-is. |
| `integration/`, `internal/`, `protoutil/`, `vendor/` | Unmodified upstream Fabric. |

Everything else (`bccsp/`, `common/`, `gossip/`, `msp/`, `orderer/`, `discovery/`, `idemix/`, and friends) is upstream Fabric, kept in the tree so the fork builds as a single project.

## Modified packages

If you want to read the diff against upstream Fabric, start here:

* `core/endorser/` is the endorsement entry point. EMVCC adds the pre-check hook that runs after simulation and before signing the proposal response.
* `core/ledger/kvledger/txmgmt/validator/` holds the shared validation utilities that the pre-check reuses. It is the same logic the commit path runs, refactored so it can be called mid-endorsement.
* `core/ledger/kvledger/txmgmt/statedb/` exposes the read paths the pre-check uses to compare read-set versions against the current committed state.
* `protoutil/` introduces an `EMVCC_CONFLICT` proposal-response status so clients can distinguish an early conflict from an endorsement-policy or chaincode error.

The paper (see [Paper and thesis](#paper-and-thesis)) walks through each touched file and explains the rationale.

## Building

EMVCC builds with the same toolchain as upstream Fabric v2.x.

**Prerequisites**

* Go 1.17 or newer
* Docker 20.10+ and Docker Compose v2
* GNU Make
* Git

**Build the peer and CLI binaries**

```bash
git clone https://github.com/HelmiTrabelsi/Fabric_EMVCC.git
cd Fabric_EMVCC

# Native binaries (peer, orderer, configtxgen, cryptogen, ...)
make native

# Or build the peer/orderer Docker images locally
make docker
```

The `make` targets follow the usual Fabric conventions. Run `make help` for the full list.

## Trying it out with EMVCC_Network

The `EMVCC_Network/` directory ships a small two-org test network that uses the EMVCC peer images.

```bash
cd EMVCC_Network

# Bring up orderers, peers, and a channel
./network.sh up createChannel

# Deploy the sample chaincode used by the conflict scenario
./network.sh deployCC -ccn basic -ccp ../path/to/chaincode -ccl go

# Tear down
./network.sh down
```

Exact flags will match what's in `EMVCC_Network/network.sh`. Check that script for any project-specific options (channel name, chaincode path, image tags).

## Reproducing the conflict scenario

The included workload is designed to expose the stock-Fabric MVCC failure mode and contrast it with EMVCC behavior. The general recipe:

1. Deploy a chaincode that keeps a small set of hot keys, for example a shared counter.
2. Drive concurrent updates to those keys from multiple clients at a rate high enough that block batching produces overlapping read-write sets.
3. Measure:
   * Aggregate abort rate. In stock Fabric this shows up as `MVCC_READ_CONFLICT` at commit; in this fork it shows up as `EMVCC_CONFLICT` at endorsement.
   * End-to-end latency for successful transactions.
   * Wasted work, meaning transactions that reached the commit stage only to be invalidated.

You should expect EMVCC to shift rejections from commit time to endorsement time, with the same or lower overall abort rate and noticeably less wasted work, because the rejected transactions never enter the ordering pipeline.

When you have numbers, drop them into a `results/` section here.

## What this prototype does not do

* It is a same-peer view. The pre-check only sees the conflicts visible to the endorsing peer's local state and its own pending endorsements. It is a best-effort early filter, not a global serializer, and the upstream commit-time check still has the final word.
* It does not touch endorsement policies, MSPs, or the ordering service. EMVCC is a pure peer-side optimization.
* Clients that do not recognize `EMVCC_CONFLICT` will simply see it as a generic proposal failure and can retry the way they would for any transient error.
* It was validated on the included `EMVCC_Network`. Production-scale topologies and adversarial workloads are out of scope for this thesis prototype.

## Paper and thesis

This code is the prototype behind:

> **Early Detection for Multiversion Concurrency Control Conflicts in Hyperledger Fabric.**
> Helmi Trabelsi and Kaiwen Zhang. arXiv:2301.06181, 2023.
> [arXiv](https://arxiv.org/abs/2301.06181) · [ResearchGate](https://www.researchgate.net/publication/367217332_Early_Detection_for_Multiversion_Concurrency_Control_Conflicts_in_Hyperledger_Fabric)

The work was done as part of a Master's thesis at École de technologie supérieure (ÉTS), Montréal:

> **Optimisation des performances de Hyperledger Fabric.**
> Helmi Trabelsi. M.Sc. thesis, École de technologie supérieure, 2022.
> [Full text (espace.etsmtl.ca)](http://espace.etsmtl.ca/2857/1/TRABELSI_Helmi.pdf)

If you build on this work, citations are appreciated:

```bibtex
@article{trabelsi2023emvcc,
  author  = {Trabelsi, Helmi and Zhang, Kaiwen},
  title   = {Early Detection for Multiversion Concurrency Control Conflicts
             in Hyperledger Fabric},
  journal = {arXiv preprint arXiv:2301.06181},
  year    = {2023}
}

@mastersthesis{trabelsi2022thesis,
  author = {Helmi Trabelsi},
  title  = {Optimisation des performances de Hyperledger Fabric},
  school = {{\'E}cole de technologie sup{\'e}rieure},
  year   = {2022}
}
```

## License

Like upstream Hyperledger Fabric, this project is released under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file. Documentation under `docs/` keeps its Creative Commons Attribution 4.0 International License (CC-BY-4.0).
