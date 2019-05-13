
<a name="v.1.1.1"></a>
## [v.1.1.1](https://github.com/qlcchain/go-qlc/compare/v1.1.0...v.1.1.1) (2019-05-13)

### Bug Fixes

* contract empty pointer ([#325](https://github.com/qlcchain/go-qlc/issues/325))
* handle p2p error ([#323](https://github.com/qlcchain/go-qlc/issues/323))
* modify balance unmarsh
* modify migrate accountmeta
* migrate accountmeta
* total balance calculation ([#316](https://github.com/qlcchain/go-qlc/issues/316))

### Features

* add peersInfo rpc interface
* batch insert sqlite data
* cli --config supports non-existent file creation ([#320](https://github.com/qlcchain/go-qlc/issues/320))
* docker image support account/override params
* modify default message count
* add limit to sms api
* add rpc interface for pledge contract. fix: some bug

### Refactoring

* withdraw pledge need NEP5TxId

### Pull Requests

* Merge pull request [#324](https://github.com/qlcchain/go-qlc/issues/324) from qlcchain/feature/peersInfo-rpc
* Merge pull request [#319](https://github.com/qlcchain/go-qlc/issues/319) from qlcchain/feature/docker
* Merge pull request [#322](https://github.com/qlcchain/go-qlc/issues/322) from qlcchain/feature/sqlite1
* Merge pull request [#315](https://github.com/qlcchain/go-qlc/issues/315) from qlcchain/feature/withdrawpledge-refactor
* Merge pull request [#318](https://github.com/qlcchain/go-qlc/issues/318) from qlcchain/hotfix/accountmeta
* Merge pull request [#314](https://github.com/qlcchain/go-qlc/issues/314) from qlcchain/feature/sms
* Merge pull request [#305](https://github.com/qlcchain/go-qlc/issues/305) from qlcchain/feature/1.0.5


<a name="v1.1.0"></a>
## [v1.1.0](https://github.com/qlcchain/go-qlc/compare/v1.0.4...v1.1.0) (2019-04-29)

### Bug Fixes

* mintage config
* contract time
* override config  value
* rpc publicModules add pledge
* account mode NPE
* default build tags
* dependence and format code
* genesis token info bug
* pledge time
* legder test case
* ledger process
* withdraw nep5 pledge
* pledge contract
* pledge contract
* tmp disable replace go mod
* ci error
* TestGetTxn test case
* Query token info
* wallet test case

### Features

* add --nobootnode to disable all bootstrap nodes
* modify representative weight ([#300](https://github.com/qlcchain/go-qlc/issues/300))
* get ntp time
* add nep5 chain contracts for testnet
* config parameter supports command line configuration
* implement config migration from v2 to v3
* add config for sqlite
* add search pledge info rpc interface
* save contract data and hash
* implement vm log and cache
* storage nep5 txid
* add benefit balance to rpc account interface
* add withdraw mintage cli
* add withdraw pledge cli
* modify generate stateblock
* modify BlockProcess
* recalculate voting weight
* pledge cli
* xgo cross compile
* use build/test tags split testnet/mainnet
* add eventbus to sqlite store  ([#271](https://github.com/qlcchain/go-qlc/issues/271))
* implement timing functions
* improve p2p performance
* implement hashmap and trie
* implement internal event bus
* implement nep5 pledge contract RPC
* implement withdraw mintage pledge RPC
* implement NEP5 pledge chain  contract

### Refactoring

* pledge history
* default config ([#297](https://github.com/qlcchain/go-qlc/issues/297))
* optimize db config
* manage eb instances
* change the sorting method
* pledge time and pledge count
* refine mintage/pledge contract


<a name="v1.0.4"></a>
## [v1.0.4](https://github.com/qlcchain/go-qlc/compare/v1.0.3...v1.0.4) (2019-04-03)

### Features

*  add generate test ledger cli([#245](https://github.com/qlcchain/go-qlc/issues/245))
* modify rpc switch ([#231](https://github.com/qlcchain/go-qlc/issues/231))
* add QGAS genesis block info ([#230](https://github.com/qlcchain/go-qlc/issues/230))


<a name="v1.0.3"></a>
## [v1.0.3](https://github.com/qlcchain/go-qlc/compare/v1.0.2...v1.0.3) (2019-03-29)


<a name="v1.0.2"></a>
## [v1.0.2](https://github.com/qlcchain/go-qlc/compare/v1.0.1...v1.0.2) (2019-03-28)

### Bug Fixes

* p2p bug

### Refactoring

* --config is followed by the configuration file


<a name="v1.0.1"></a>
## [v1.0.1](https://github.com/qlcchain/go-qlc/compare/v1.0.0...v1.0.1) (2019-03-22)

### Bug Fixes

* migration rpc/p2p config to v2
* default rpc public modules is nil
* mintage contract pledger amount
* testnet genesis block data field error

### Features

* check rpc enable
* add timestamp to AccountsPending api
* modify balance unit conversion
* multiple block can contain the same message hash
* implement migration config
* configure rpc module enable
* add version for mainnet
* add endpoint for testnet
* enhance server cli ([#209](https://github.com/qlcchain/go-qlc/issues/209))

### Refactoring

* move consensus test case to test folder


<a name="v1.0.0"></a>
## [v1.0.0](https://github.com/qlcchain/go-qlc/compare/v0.1.0...v1.0.0) (2019-03-15)

### Bug Fixes

* online Representatives bug

### Features

* add version command for cli server
* move tokeninfo interface from mintage to ledger
* add switch that automatically generates receive block ([#198](https://github.com/qlcchain/go-qlc/issues/198))
* add ledger testcase

### Refactoring

* accountsPending rpc interface


<a name="v0.1.0"></a>
## [v0.1.0](https://github.com/qlcchain/go-qlc/compare/v0.0.9...v0.1.0) (2019-03-14)

### Bug Fixes

* find online representatives bug


<a name="v0.0.9"></a>
## [v0.0.9](https://github.com/qlcchain/go-qlc/compare/v0.0.5...v0.0.9) (2019-03-14)

### Bug Fixes

* ledger
* consensus bug.
* ledger
* sync error
* local representatives bug
* generate receive block bug
* block check
* ledger and consensus integrate testcase
* ledger and verify token symbol and token name
* rpc test
* init genesis block
* update frontier in rollback
* genesis block and add genesis testcase
* Invalid change to InvalidData
* [#154](https://github.com/qlcchain/go-qlc/issues/154)
* build error
* message cache bug in p2p module
* peer ID conversion bug.
* ledger unpack TokenInfo
* consensus modules init account bug.
* mdns bug and node discovery structure adjustment. - fix [#162](https://github.com/qlcchain/go-qlc/issues/162).
* wallet/vm test case ([#159](https://github.com/qlcchain/go-qlc/issues/159))
* execute blockprocess in a transaction
* wallet testcase
* libp2p dependence
* bug for get uncheckedblock from db([#122](https://github.com/qlcchain/go-qlc/issues/122))

### Features

* add two bootNodes
* enhance auto generate receive block
* implement search pending by address
* add docker
* verify token symbol and token name
* api sms
* rpc ledger
* ledger interface
* increase the processing of transactions related to smart contracts
* ledgerstore define
* integrate chain contract with consensus,fix [#42](https://github.com/qlcchain/go-qlc/issues/42) and fix [#104](https://github.com/qlcchain/go-qlc/issues/104)
* add p2p testcase
* add consensus testcase
* add test case in ledger
* add interface to get block by phone number
* add p2p parameters to the configuration file
* remove duplicate print log
* add tokeninfo when add block
* implement ledger interface to store contract data
* merge performance test code and add configuration switches for performance testing
* implement mintage contract
* add chain contract interface
* add message cache. - implement consensus message cache to ensure that previously received transactions are no longer broadcast. - implement p2p message cache to ensure that the message can be resent after the message fails to be sent. - fix rollback bug. - fix bug for handling fork transactions. - fix [#123](https://github.com/qlcchain/go-qlc/issues/123) - fix [#124](https://github.com/qlcchain/go-qlc/issues/124). - some other optimizations.
* blockprocess error return
* implement abi pack/unpack ([#156](https://github.com/qlcchain/go-qlc/issues/156))
* optimize common lib ([#155](https://github.com/qlcchain/go-qlc/issues/155))
* add version for ledger
* add table for blocks relationship
* set balance unit to raw in rpc parameter and return
* add interactive and non-interactive mode for cli ([#138](https://github.com/qlcchain/go-qlc/issues/138))
* change return unit for balance in rpc interface
* add config for rpc network type
* implement interactive cli by ishell library
* add meta data for uncheckedblock table from db
* optimization of the external interface of p2p messages([#125](https://github.com/qlcchain/go-qlc/issues/125))
* use online representation of 50% of total weight as a consensus threshold ([#120](https://github.com/qlcchain/go-qlc/issues/120))
* add api to return all blocks and addresses ([#121](https://github.com/qlcchain/go-qlc/issues/121))
* implement VM and base functions ([#93](https://github.com/qlcchain/go-qlc/issues/93))

### Refactoring

* change sender and receiver of StateBlock
* use seed to run node with account
* optimize mintage contract
* refine service
* optimize block interface  ([#165](https://github.com/qlcchain/go-qlc/issues/165))
* add version.go and set params
* qlc CLI
* modify rpc test case
* modify the way to get all data in a badger table


<a name="v0.0.5"></a>
## [v0.0.5](https://github.com/qlcchain/go-qlc/compare/v0.0.4...v0.0.5) (2019-01-30)

### Bug Fixes

* change map to sync.map for roots

### Features

* add wallet,util and net interface to json-rpc
* use std json
* implement save/load performance test time ([#96](https://github.com/qlcchain/go-qlc/issues/96))
* add create account and batch send token commands


<a name="v0.0.4"></a>
## [v0.0.4](https://github.com/qlcchain/go-qlc/compare/v0.0.3...v0.0.4) (2019-01-23)

### Bug Fixes

* cpu too high. - optimize code that processes large amounts of data.
* solve synchronization problems caused by frontier incorrect sorting

### Features

* print debug info to rpc request ([#89](https://github.com/qlcchain/go-qlc/issues/89))


<a name="v0.0.3"></a>
## [v0.0.3](https://github.com/qlcchain/go-qlc/compare/v0.0.2...v0.0.3) (2019-01-15)

### Bug Fixes

* log configuration ([#87](https://github.com/qlcchain/go-qlc/issues/87))
* windows rpc endpoint


<a name="v0.0.2"></a>
## [v0.0.2](https://github.com/qlcchain/go-qlc/compare/v0.0.1...v0.0.2) (2019-01-15)

### Bug Fixes

* add find Online Representatives timer task
* consensus bug
* support choose config path
* windows ipc testcase
* let node know online representatives
* AccountHistoryTopn() returns one more
* amountbalance NPE
* mock stateblock
* rpcapi-Process()
* balance sub
* GetOnlineRepresentatives() return
* solve CPU usage is too high

### Refactoring

* rpcapi-AccountHistoryTopn()
* Process return block hash
* move logger to QlcApi
* rpc interface
*  set token symbol to uppercase
* replace jsoniter with std json
* use cobra refine cli.


<a name="v0.0.1"></a>
## v0.0.1 (2018-12-28)

### Bug Fixes

* cli auto get root_token type
* build windows 386 error
* convert balance to smallest unit ([#65](https://github.com/qlcchain/go-qlc/issues/65))

### Features

* Add log, wallet, ledger internal service

### Refactoring

* remove json files
* integrate test ([#62](https://github.com/qlcchain/go-qlc/issues/62))
* integration test ([#58](https://github.com/qlcchain/go-qlc/issues/58))

### Pull Requests

* Merge pull request [#11](https://github.com/qlcchain/go-qlc/issues/11) from qlcchain/feature/task

