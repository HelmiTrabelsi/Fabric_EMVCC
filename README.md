Fabric_EMVCC

Early MVCC conflict detection for Hyperledger Fabric.

Fabric_EMVCC is the research prototype that accompanies the paper "Early Detection for Multiversion Concurrency Control Conflicts in Hyperledger Fabric" (Trabelsi & Zhang, arXiv:2301.06181, 2023). It is a fork of Hyperledger Fabric v2.x that moves Multi-Version Concurrency Control (MVCC) conflict detection from the commit phase to the endorsement phase. Transactions that are guaranteed to fail at validation are caught — and rejected — before they are ever ordered into a block, eliminating wasted work, reducing end-to-end latency, and increasing effective throughput under contention.



Forked from Hyperledger Fabric v2.x. The upstream codebase is preserved; the EMVCC changes are localized and clearly scoped (see Modified packages).



Table of contents





Background: the MVCC problem in Fabric



The EMVCC approach



Repository layout



Modified packages



Building



Trying it out — EMVCC_Network



Reproducing the conflict scenario



Limitations and scope



Academic context



License



Acknowledgements



Background: the MVCC problem in Fabric

Hyperledger Fabric follows an execute → order → validate transaction flow:





Execute (endorsement). A client sends a proposal to one or more endorsing peers. Each peer simulates the chaincode against its current world state and produces a read–write set (RWSet) — the set of keys read (with their versions) and the writes the transaction would apply. The peer signs the RWSet and returns it.



Order. The client collects enough endorsements and sends the signed transaction to the ordering service, which bundles transactions into blocks and disseminates them.



Validate (commit). Every committing peer re-checks each transaction in the block. The MVCC check compares the versions in the RWSet against the world state at commit time. If any key's version has changed since simulation, the transaction is marked MVCC_READ_CONFLICT and its writes are discarded — but it still occupies a slot in the block.

This design gives Fabric its non-deterministic-chaincode tolerance and parallel execution properties, but it has a well-known weakness: conflicts are only detected at the very end of the pipeline.

The practical consequences:





Wasted work. A doomed transaction is simulated, endorsed, signed, gossiped, ordered, written to a block, and validated — all to be thrown away at the last step.



Head-of-line blocking under contention. Workloads with hot keys (counters, registries, balance accounts) can see large fractions of a block invalidated, especially as block size and throughput increase.



Opaque failures for clients. Clients only learn the outcome after the block is committed, well after they could have done anything about it.



Stale endorsements. A correct-at-the-time endorsement becomes invalid before it even reaches the orderer if a concurrent transaction touches the same key first.

The EMVCC approach

EMVCC — Early MVCC — adds a conflict-detection step inside the endorsing peer, before the endorsement signature is returned to the client. The intuition: a peer already maintains a recent view of the world state and a queue of in-flight RWSets it has just endorsed; it can use that information to detect transactions that are certain to conflict at commit time, and reject them at endorsement.

At a high level, the modified endorsement path:





Simulates the chaincode as in stock Fabric and produces the candidate RWSet.



Checks every key in the read set against:





the current committed version in the local state DB, and



the pending writes from RWSets this peer has recently endorsed and not yet seen committed.



If a conflict is detected, the peer returns a typed EMVCC_CONFLICT response to the client instead of an endorsement, with the conflicting key(s) named. The client can immediately back off, re-read, and retry — without the proposal ever entering the ordering pipeline.



If no conflict is detected, the endorsement proceeds normally; the upstream commit-time MVCC check still runs as a final correctness gate.

       Stock Fabric                              EMVCC
       -----------                               ------
  Client → Endorse                          Client → Endorse
            ↓                                          ↓
        (simulate)                                (simulate)
            ↓                                          ↓
        endorsed                            ┌── EMVCC pre-check ──┐
            ↓                               │                     │
         Order                              ▼                     ▼
            ↓                          endorsed              EMVCC_CONFLICT
        Validate                           ↓                  (early reject)
            ↓                            Order
   MVCC_READ_CONFLICT                       ↓
       at commit                        Validate
                                            ↓
                                    (rarely triggers — already filtered)

The commit-time check is kept intact: EMVCC is an additional gate, not a replacement. Correctness is therefore unchanged — EMVCC can only reject transactions that the existing validator would have rejected anyway. What changes is when and how cheaply the rejection happens.

Repository layout

The top of the tree is a Hyperledger Fabric v2.x source tree. The two folders most relevant to this fork are:







Path



What's in it





EMVCC_Network/



A self-contained test network (modeled on fabric-samples/test-network) for bringing up EMVCC-enabled peers and reproducing the conflict scenario.





core/



The Fabric peer core, including the endorsement and ledger code paths modified by EMVCC.





docs/



Upstream documentation tree, kept in sync with Fabric.





integration/, internal/, protoutil/, vendor/



Unmodified upstream Fabric packages.

Everything else (bccsp/, common/, gossip/, msp/, orderer/, discovery/, idemix/, …) is upstream Fabric, included to keep the fork buildable as a single tree.

Modified packages

If you want to read the diff against upstream Fabric, start here:





core/endorser/ — the endorsement entry point. EMVCC adds the pre-check hook that runs after simulation and before signing the proposal response.



core/ledger/kvledger/txmgmt/validator/ — shared validation utilities reused by the pre-check (the same logic that runs at commit time, refactored to be callable mid-endorsement).



core/ledger/kvledger/txmgmt/statedb/ — read paths exposed to the pre-check so it can compare RWSet versions against the current committed state.



protoutil/ — a new EMVCC_CONFLICT proposal-response status so clients can distinguish early conflicts from endorsement-policy or chaincode errors.



The thesis (see Academic context) documents each touched file and the rationale.

Building

EMVCC builds with the same toolchain as upstream Fabric v2.x.

Prerequisites





Go 1.17+



Docker 20.10+ and Docker Compose v2



GNU Make



Git

Build the peer and CLI binaries

git clone https://github.com/HelmiTrabelsi/Fabric_EMVCC.git
cd Fabric_EMVCC

# Native binaries (peer, orderer, configtxgen, cryptogen, …)
make native

# Or build the peer/orderer Docker images locally
make docker

make targets follow the upstream Fabric conventions. Run make help for the full list.

Trying it out — EMVCC_Network

The EMVCC_Network/ directory ships a small two-org test network that uses the EMVCC-modified peer images.

cd EMVCC_Network

# Bring up orderers, peers, and a channel
./network.sh up createChannel

# Deploy the sample chaincode used by the conflict scenario
./network.sh deployCC -ccn basic -ccp ../path/to/chaincode -ccl go

# Tear down
./network.sh down

Exact flags will match what's in EMVCC_Network/network.sh; check that script for any project-specific options (channel name, chaincode path, image tags).

Reproducing the conflict scenario

The repository includes a workload designed to expose the stock-Fabric MVCC failure mode and contrast it with EMVCC behavior. The general shape:





Deploy a chaincode that maintains a small set of hot keys (e.g., a shared counter).



Drive concurrent updates to those keys from multiple clients with a high enough request rate that block-level batching causes overlapping RWSets.



Measure:





Aggregate abort rate (fraction of transactions ending in MVCC_READ_CONFLICT in stock Fabric, or EMVCC_CONFLICT early-reject in this fork).



End-to-end latency for successful transactions.



Wasted work (transactions that reached the commit stage only to be invalidated).

You should expect EMVCC to shift rejections from commit-time MVCC_READ_CONFLICT to endorsement-time EMVCC_CONFLICT, with the same — or lower — overall abort rate and noticeably lower wasted work, because rejected transactions never enter the ordering pipeline.

Plug your measurement numbers into a results/ section here when ready.

Limitations and scope





Same-peer view only. The pre-check sees only the conflicts visible to the endorsing peer's local state and its own pending endorsements. It is a best-effort early filter, not a global serializer — the upstream commit-time check still has the final word and remains correct.



Endorsement policy untouched. EMVCC does not change endorsement policies, MSPs, or the ordering service. It is a pure peer-side optimization.



No protocol change required for unmodified clients. Clients that don't recognize EMVCC_CONFLICT will simply see it as a generic proposal failure and can retry as they would for any transient error.



Tested on a controlled topology. EMVCC has been validated on the included EMVCC_Network; production-scale topologies and adversarial workloads are out of scope for this thesis prototype.

Academic context

This code is the prototype that accompanies the paper:



Early Detection for Multiversion Concurrency Control Conflicts in Hyperledger Fabric.
Helmi Trabelsi and Kaiwen Zhang. arXiv:2301.06181, 2023.
arXiv · ResearchGate

The work was carried out as part of the author's Master's thesis at École de technologie supérieure (ÉTS), Montréal:



Optimisation des performances de Hyperledger Fabric.
Helmi Trabelsi. M.Sc. thesis, École de technologie supérieure, 2022.
Full text (espace.etsmtl.ca)

If you build on this work, a citation along these lines is appreciated:

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

License

Like upstream Hyperledger Fabric, this project is released under the Apache License, Version 2.0. See the LICENSE file. The Fabric documentation under docs/ retains its Creative Commons Attribution 4.0 International License (CC-BY-4.0).

Acknowledgements

Built on top of the work of the Hyperledger Fabric maintainers and contributors. The EMVCC modifications are localized; the broader peer, orderer, gossip, ordering, and identity subsystems are theirs.
