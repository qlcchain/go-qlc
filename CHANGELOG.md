
<a name="v1.2.3.1"></a>
## [v1.2.3.1](https://github.com/qlcchain/go-qlc/compare/v1.2.3...v1.2.3.1) (2019-08-02)

### Bug Fixes

* rollback block cache ([#497](https://github.com/qlcchain/go-qlc/issues/497))
* verify exist data when generate rewards data
* some rpc interface return incorrect info
* possible NPE
* return all valid pending infos even exception occurs
* getBlockInfo amount param miss
* auto generate receive block may not be correct

### Pull Requests

* Merge pull request [#496](https://github.com/qlcchain/go-qlc/issues/496) from qlcchain/hotfix/verify-rewards-block
* Merge pull request [#494](https://github.com/qlcchain/go-qlc/issues/494) from qlcchain/hotfix/all-pendings
* Merge pull request [#495](https://github.com/qlcchain/go-qlc/issues/495) from qlcchain/hotfix/rpc
* Merge pull request [#493](https://github.com/qlcchain/go-qlc/issues/493) from qlcchain/hotfix/amount-zero
* Merge pull request [#492](https://github.com/qlcchain/go-qlc/issues/492) from qlcchain/feature/autoGenerateReceiveBlock


<a name="v1.2.3"></a>
## [v1.2.3](https://github.com/qlcchain/go-qlc/compare/v1.2.2...v1.2.3) (2019-07-31)

### Bug Fixes

* accountMetaCache may be not the newest ([#491](https://github.com/qlcchain/go-qlc/issues/491))
* optimize the voting process
* add block child error
* confirmReq may not get confirmAck
* The transaction cannot be confirmed when the representative re online
* start node with wallet call ledger init twice
* badger transaction conflict. fix rollback block bug.
* modify pov miner generate block
* merge master to branch feature/pov-consensus-refactor
* modify mine worker num to 1
* modify pov block size to 2MB
* change miner contract address to 21
* modify miner reward cmd & rpc
* add start & end height to miner reward param
* miner reward shoud check duplicate send & recv
* remove MinerApi vmContext field to local var
* remove useless time sleep to reduce cpu usage
* restore miner contract address to 3
* restore miner contract address to 3
* change consensus debug level and cache block check result
* optimize block check
* merge master to branch feature/refactor-miner-contract
* optimize cons
* optimize consensus
* reward contract should check miner pledge vote
* do not check pov sync state if pov service is not enabled
* remove useless time sleep to reduce cpu usage
* nep5txid encoding/hex error
* add PovMinerRewardHeightStart param
* change min enough pov sync peer count
* reduce duplicate block message in pov sync phase
* save unchecked block to ledger
* sync bug
* rollback error may cause nil pointer painc
* use ; split instead of : split for --configParams
* change dpos get cache len
* change dpos cache
* optimize dpos msg
* remove active tx after 30mins
* optimize dpos msg
* using bigger retry time interval to reduce messages
* consensus testcase
* optimze dpos vote msg
* channel blocked in sendMessageToPeer func
* write twice to messagechan
* badger go mod dep
* badger version
* change receive mag cache time
* add rep vote debug
* change debug level
* read chan timeout
*  pov_getBlockByHeight may return too large size of data
* stream.messageChan <- message may blocked all the time fix [#420](https://github.com/qlcchain/go-qlc/issues/420)
* messageService loop can not exit
* `configparams` can not override string array
* pov sync slow peer init time
* check pov sync peer is too slow
* pov sync not working
* check pov sync rx & block request cache
* badger dependence
* pov block td calc is incorrect
* generate_test_ledger cliâ
* bigger pov sync msg channel size
* pull more state blocks in pov sync
* reduce duplicate message in pov syncing phase
* send cli bug
* pov txtool not pack mintage genesis contract send blocks
* update pov block num per day
* generate_test_ledger cli no need designated account. peldge and mintage cli should use random nep5Txid for test
* cmd pledge should using random nep5 txid just for testing
* `StringToBalance` memory leak
* consensus check pov height
* optimize pov txpool recoverUnconfirmedTxs
* pov txpool SelectPendingTxs
* move hash in AccountState to TokenState
* update debug print info
* representative node send tx bug
* adjust order to close the modules
* change pov miner RewardPerBlock
* change PovMinerPledgeAmountMin from 50w to 100w
* remove comment
* fork process bug
* process fork bug and reduce msg cache time
* rollback unchecked block when rollback stateBlock
* rpc interface NPE ([#405](https://github.com/qlcchain/go-qlc/issues/405))
* reduce cache size
* mod proccess msg sleep interval
* reduce cache size and set gc percent
* configParams bug when use default config name
* updateConfig testcase
* configParms bug
* execute contract more than one time bug
* compile error
* consensus sync bug and merge master
* websocket CORS ([#386](https://github.com/qlcchain/go-qlc/issues/386))
* cpu high loading when query token info
* verify rewards sign
* random block bug
* possible NPE
* test cases
* encode trie
* balance check when process block bug. fix generate block rpc interface bug.
* generate send balance
* process fork for cotractreward block fix totalsupply bug
* pledge DoReceive bug
* QGAS rewards balance
* generate reward block pending
* change the method of judging open block
* find online representation bug
* fix ci error
* process blocks synchronized from other nodes
* fix process msg loop bug
* pov tx pending bug
* target should be current block, not for next block; modify int64 to uint64;

### Features

* pov ledger miner stats ([#479](https://github.com/qlcchain/go-qlc/issues/479))
* modify test case
* complete the functions of the v1.2.3 plan ([#474](https://github.com/qlcchain/go-qlc/issues/474))
* add pov fake consesus
* refactor miner contract ([#455](https://github.com/qlcchain/go-qlc/issues/455))
* add miner history cmd
* add pov pow consensus
* merge feature/refactor-miner-contract
* pov_getMinerStats rpc add online count
* pov rpcs should check cfg enabled or not
* modify pov target timespan params
* change pov genesis block and split pov params
* using PovHeader not PovBlock to reduce cache memory usaging
* move pov files to sub folder
* update libp2p package
* add rewardBlocks to MinerReward contract param
* add pov rpc getLatestAccountState
* don't save contract data when block check
* refactoring pov miner contract
* disable pov service by default in config v4
* check msg in p2p cache before send
* add ctrl channel for reponse message
* send PovBulkPullReq to block source peer
* rollback contract block  ([#428](https://github.com/qlcchain/go-qlc/issues/428))
* move pov block parameter from config file to common const
* optimize pov sync using batch request blocks & txs
* add rpc interface OnlineRepsInfo
* need to have three million QLC to vote
* add change representative cmd
* add pov_batchGetHeadersByHeight rpc
* rpc add pov_getFittestHeader
* add ledger GetLatestPovHeader
* add yamux option for p2p network
* add peer connect status judge when syncing
* print config details
* remove print
* modify RawToBalance interface
* add sync status rpc interface ([#377](https://github.com/qlcchain/go-qlc/issues/377))
* add newaccounts rpc interface
* modify ledger iterator error log
* add ledger_process test case
* modify get all uncheckblock
* judge pledge amount
* implement query rewards details RPC
* add log to ledger
* add log to ledger
* modify account balance return struct
* Implement POV chain - consensus
* add pov ledger functions
* add pov ledger functions

### Refactoring

* use context exit goroutine ([#487](https://github.com/qlcchain/go-qlc/issues/487))
* remove qlcclassic api
* split json RPC lib as a standalone project
* sync RPC lib upstream
* if the configuration file is modified incorrectly, keep the original content and prompt the error
* make rewards exception log clearer
* refactor consensus module([#275](https://github.com/qlcchain/go-qlc/issues/275) [#276](https://github.com/qlcchain/go-qlc/issues/276))

### Pull Requests

* Merge pull request [#490](https://github.com/qlcchain/go-qlc/issues/490) from qlcchain/hotfix/vote-optimize
* Merge pull request [#488](https://github.com/qlcchain/go-qlc/issues/488) from qlcchain/hotfix/childerror
* Merge pull request [#486](https://github.com/qlcchain/go-qlc/issues/486) from qlcchain/hotfix/confirmReq
* Merge pull request [#485](https://github.com/qlcchain/go-qlc/issues/485) from qlcchain/hotfix/trxConfirmed
* Merge pull request [#483](https://github.com/qlcchain/go-qlc/issues/483) from qlcchain/hotfix/ledgerinit
* Merge pull request [#476](https://github.com/qlcchain/go-qlc/issues/476) from qlcchain/hotfix/v1.2.3
* Merge pull request [#475](https://github.com/qlcchain/go-qlc/issues/475) from qlcchain/feature/testcase
* Merge pull request [#467](https://github.com/qlcchain/go-qlc/issues/467) from qlcchain/feature/pov-consensus-refactor
* Merge pull request [#465](https://github.com/qlcchain/go-qlc/issues/465) from qlcchain/feature/remove-time-sleep
* Merge pull request [#463](https://github.com/qlcchain/go-qlc/issues/463) from qlcchain/feature/cons-optimize
* Merge pull request [#454](https://github.com/qlcchain/go-qlc/issues/454) from qlcchain/feature/libp2pUpdate
* Merge pull request [#450](https://github.com/qlcchain/go-qlc/issues/450) from qlcchain/feature/save-contractdata
* Merge pull request [#441](https://github.com/qlcchain/go-qlc/issues/441) from qlcchain/hotfix/cons-check-pov
* Merge pull request [#440](https://github.com/qlcchain/go-qlc/issues/440) from qlcchain/hotfix/randomnep5id
* Merge pull request [#439](https://github.com/qlcchain/go-qlc/issues/439) from qlcchain/feature/pov-optimize
* Merge pull request [#438](https://github.com/qlcchain/go-qlc/issues/438) from qlcchain/feature/libp2pDep
* Merge pull request [#430](https://github.com/qlcchain/go-qlc/issues/430) from qlcchain/hotfix/consensus_testcase
* Merge pull request [#429](https://github.com/qlcchain/go-qlc/issues/429) from qlcchain/hotfix/sendmessagetopeer
* Merge pull request [#427](https://github.com/qlcchain/go-qlc/issues/427) from qlcchain/hotfix/messagechan
* Merge pull request [#426](https://github.com/qlcchain/go-qlc/issues/426) from qlcchain/hotfix/badger
* Merge pull request [#425](https://github.com/qlcchain/go-qlc/issues/425) from qlcchain/feature/pov-integrating
* Merge pull request [#411](https://github.com/qlcchain/go-qlc/issues/411) from qlcchain/feature/conspov
* Merge pull request [#406](https://github.com/qlcchain/go-qlc/issues/406) from qlcchain/feature/sync
* Merge pull request [#404](https://github.com/qlcchain/go-qlc/issues/404) from qlcchain/hotfix/configbug
* Merge pull request [#397](https://github.com/qlcchain/go-qlc/issues/397) from qlcchain/hotfix/config
* Merge pull request [#378](https://github.com/qlcchain/go-qlc/issues/378) from qlcchain/feature/accountbalance
* Merge pull request [#395](https://github.com/qlcchain/go-qlc/issues/395) from qlcchain/feature/rawTobalance


<a name="v1.2.2"></a>
## [v1.2.2](https://github.com/qlcchain/go-qlc/compare/v1.2.1...v1.2.2) (2019-06-03)

### Bug Fixes

* delete unchecked block bug
* websocket CORS ([#386](https://github.com/qlcchain/go-qlc/issues/386))
* cpu high loading when query token info
* verify rewards sign
* random block bug
* possible NPE
* test cases
* encode trie
* balance check when process block bug. fix generate block rpc interface bug.
* generate send balance

### Features

* change ledger verfier checkblock map to local variable
* delete unchecked block depend on forked block
* add sync status rpc interface ([#377](https://github.com/qlcchain/go-qlc/issues/377))
* add newaccounts rpc interface
* modify ledger iterator error log
* add ledger_process test case
* modify get all uncheckblock
* judge pledge amount

### Refactoring

* make rewards exception log clearer

### Pull Requests

* Merge pull request [#394](https://github.com/qlcchain/go-qlc/issues/394) from qlcchain/feature/verifiermap
* Merge pull request [#390](https://github.com/qlcchain/go-qlc/issues/390) from qlcchain/hotfix/unchecked
* Merge pull request [#389](https://github.com/qlcchain/go-qlc/issues/389) from qlcchain/hotfix/unchecked
* Merge pull request [#388](https://github.com/qlcchain/go-qlc/issues/388) from qlcchain/hotfix/optmize-error-log
* Merge pull request [#385](https://github.com/qlcchain/go-qlc/issues/385) from qlcchain/hotfix/high-cpu-loading
* Merge pull request [#384](https://github.com/qlcchain/go-qlc/issues/384) from qlcchain/hotfix/verify-rewards-sign
* Merge pull request [#381](https://github.com/qlcchain/go-qlc/issues/381) from qlcchain/hotfix/onlinerep
* Merge pull request [#380](https://github.com/qlcchain/go-qlc/issues/380) from qlcchain/hotfix/generate-reward-hash
* Merge pull request [#379](https://github.com/qlcchain/go-qlc/issues/379) from qlcchain/test/ccontractblocktestcase
* Merge pull request [#375](https://github.com/qlcchain/go-qlc/issues/375) from qlcchain/hotfix/v1.2-issues
* Merge pull request [#374](https://github.com/qlcchain/go-qlc/issues/374) from qlcchain/feature/parse-rewards-testcase


<a name="v1.2.1"></a>
## [v1.2.1](https://github.com/qlcchain/go-qlc/compare/v1.2.0...v1.2.1) (2019-05-26)

### Bug Fixes

* process fork for cotractreward block fix totalsupply bug
* pledge DoReceive bug
* QGAS rewards balance
* generate reward block pending

### Features

* implement query rewards details RPC
* add log to ledger
* add log to ledger

### Pull Requests

* Merge pull request [#373](https://github.com/qlcchain/go-qlc/issues/373) from qlcchain/hotfix/fork
* Merge pull request [#372](https://github.com/qlcchain/go-qlc/issues/372) from qlcchain/feature/reward-details-rpc
* Merge pull request [#371](https://github.com/qlcchain/go-qlc/issues/371) from qlcchain/hotfix/doreceive
* Merge pull request [#370](https://github.com/qlcchain/go-qlc/issues/370) from qlcchain/feature/log
* Merge pull request [#369](https://github.com/qlcchain/go-qlc/issues/369) from qlcchain/hotfix/rewards-balance
* Merge pull request [#368](https://github.com/qlcchain/go-qlc/issues/368) from qlcchain/hotfix/generate-rewards


<a name="v1.2.0"></a>
## [v1.2.0](https://github.com/qlcchain/go-qlc/compare/v1.1.2...v1.2.0) (2019-05-25)

### Bug Fixes

* integrate test case error
* contract reward block check error
* find online representative bug
* openblock judgment logic need to send a confirmation notice if block from sync
* parseInt on 32bit OS
* sync bug
* --nobootnode bug

### Features

* fix ledger testcase
* eventbus support sync subscribe
* modify stateblock parent
* remove work/sign check for Rewards contractblock
* add chan to sqlite
* set generate block interface privatekey optional
* update pending for contract block
* event bus support buff
* if the node is synchronizing, the synchronization will not be repeated
* add vote cache
* implement airdrop/confidant rewards RPC interface
* remove print
* modify sqlx version
* implement GetTotalPledgeAmount
* implement confidant/airdrop rewards
* add rewards contract
* remove vm field of pledge rpc module
* add block batch
* use ntp time when generate block
* modify Representatives interface return
* support build linux/arm-7
* implement spinlock
* implement ecies

### Refactoring

* merge generate confidant/airdrop receive block into one
* implement handler list
* rewards rpc interface
* mintage and pledge withdraw time

### Pull Requests

* Merge pull request [#367](https://github.com/qlcchain/go-qlc/issues/367) from qlcchain/hotfix/integrate
* Merge pull request [#366](https://github.com/qlcchain/go-qlc/issues/366) from qlcchain/feature/parent
* Merge pull request [#365](https://github.com/qlcchain/go-qlc/issues/365) from qlcchain/feature/eb-sync-sub
* Merge pull request [#364](https://github.com/qlcchain/go-qlc/issues/364) from qlcchain/hotfix/rewardblock
* Merge pull request [#363](https://github.com/qlcchain/go-qlc/issues/363) from qlcchain/hotfix/findonlinerep
* Merge pull request [#362](https://github.com/qlcchain/go-qlc/issues/362) from qlcchain/feature/rewards-testcase
* Merge pull request [#361](https://github.com/qlcchain/go-qlc/issues/361) from qlcchain/feature/rewardblock
* Merge pull request [#360](https://github.com/qlcchain/go-qlc/issues/360) from qlcchain/feature/sqlite
* Merge pull request [#313](https://github.com/qlcchain/go-qlc/issues/313) from qlcchain/feature/eb-buffer
* Merge pull request [#357](https://github.com/qlcchain/go-qlc/issues/357) from qlcchain/intergrate-test
* Merge pull request [#359](https://github.com/qlcchain/go-qlc/issues/359) from qlcchain/feature/generateblocks
* Merge pull request [#358](https://github.com/qlcchain/go-qlc/issues/358) from qlcchain/feature/pending
* Merge pull request [#356](https://github.com/qlcchain/go-qlc/issues/356) from qlcchain/hotfix/unmarshal_balance
* Merge pull request [#355](https://github.com/qlcchain/go-qlc/issues/355) from qlcchain/feature/rewards
* Merge pull request [#354](https://github.com/qlcchain/go-qlc/issues/354) from qlcchain/feature/sqlitebatch
* Merge pull request [#351](https://github.com/qlcchain/go-qlc/issues/351) from qlcchain/feature/pledgemodule
* Merge pull request [#350](https://github.com/qlcchain/go-qlc/issues/350) from qlcchain/feature/blocktime
* Merge pull request [#349](https://github.com/qlcchain/go-qlc/issues/349) from qlcchain/feature/refactor-pledgetime
* Merge pull request [#347](https://github.com/qlcchain/go-qlc/issues/347) from qlcchain/feature/arm-7
* Merge pull request [#348](https://github.com/qlcchain/go-qlc/issues/348) from qlcchain/feature/representatives
* Merge pull request [#321](https://github.com/qlcchain/go-qlc/issues/321) from qlcchain/feature/ecies
* Merge pull request [#335](https://github.com/qlcchain/go-qlc/issues/335) from qlcchain/feature/nobootnode
* Merge pull request [#334](https://github.com/qlcchain/go-qlc/issues/334) from qlcchain/feature/spinlock


<a name="v1.1.2"></a>
## [v1.1.2](https://github.com/qlcchain/go-qlc/compare/v1.1.1...v1.1.2) (2019-05-15)

### Bug Fixes

* configuration parameters invalid

### Features

* search pledge info via nep5Id
* generate test ledger support designated account

### Refactoring

* config structure

### Pull Requests

* Merge pull request [#332](https://github.com/qlcchain/go-qlc/issues/332) from qlcchain/feature/pledgeinfosearch
* Merge pull request [#331](https://github.com/qlcchain/go-qlc/issues/331) from qlcchain/feature/testledger
* Merge pull request [#330](https://github.com/qlcchain/go-qlc/issues/330) from qlcchain/hotfix/config
* Merge pull request [#326](https://github.com/qlcchain/go-qlc/issues/326) from qlcchain/feature/reconfig


<a name="v1.1.1"></a>
## [v1.1.1](https://github.com/qlcchain/go-qlc/compare/v1.1.0...v1.1.1) (2019-05-13)

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

