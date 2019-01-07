# go-qlc    CLI    Introduction

## go-qlc  builds  the  underlying  architecture  of   cli  with `cobra`

### Commands

- ### `bc`

  - feature：returns the total number of blocks in the database

  - flag：this command does not need to have flag

  - example:

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go bc
    total block count is : 15
    ```



- ### `send`

  - feature：send transaction

  - flag：

    - -f：send account
    - -t：receive account
    - -k：token hash for send action
    - -m：send amount
    - -p：password for account

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go send -f qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic -t qlc_1uxmmpagxqoeupatnj9a9qohtx83ezsrcnzs9znycih5un
    eeopdqj9s9ct6s -m 100 -p 123
    ```

- ### `tl`

  - feature：token list

  - flag：this command does not need to have flag

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go tl
    1、TokenId:9bf0dd78eb52f56cf698990d7d3e4f0827de858f6bdabc7713c869482abfd914  TokenName:QLC  TokenSymbol:qlc  TotalSupply:60000000000000000  Decimals:8  Owner:qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic
    2、TokenId:02acb4a0e87c1fdd09b794a317ccaa6eddcb143c5fa0139f6fe145e9d94fcb17  TokenName:QN1  TokenSymbol:qn1  TotalSupply:70000000000000000  Decimals:8  Owner:qlc_1uudbhpenjr1ii3k4ah9kax8f5f1qh3tt9qirjkj9mghew61qkfibw4haxtc
    3、TokenId:3b32f29885c04d9931319a8a564692880f68d1310a4ed527dd65bfd87896e8e4  TokenName:QN2  TokenSymbol:qn2  TotalSupply:50000000000000000  Decimals:8  Owner:qlc_3s1agkbw6osftnodbcu9otawgdhz6q74xzpgsu641qzjgs8qdqfujim3z7ii
    4、TokenId:6b50aa569c41d3a4efdfbc2138590fb39c556db4a39deb589c3f8b55530ce074  TokenName:QN3  TokenSymbol:qn3  TotalSupply:90000000000000000  Decimals:8  Owner:qlc_1dtjuwiibgq7n5zcockqdyfgap7egfbxzqw8ygbtjn5aqffwxccxensu53i8
    5、TokenId:f8a2f79c397875d86cdf9342fc060449a95843cbcf46d2cfcac34f7489660604  TokenName:QN4  TokenSymbol:qn4  TotalSupply:80000000000000000  Decimals:8  Owner:qlc_3uq446hur8j7rgqu1bu1phyxb8gk3kkihqp3aqjghyg1hagsq8sizm4n7596
    6、TokenId:b9cd84de89deb22305779c39d9a94461ce7b31bf91bf9d4f631d09c1881d18b2  TokenName:QN5  TokenSymbol:qn5  TotalSupply:90000000000000000  Decimals:8  Owner:qlc_1yuh366sczsfdekt183pp7b9d8gztj4mg3bshsyam7mxxrmy4nn89638o3me
    ```

- ### `version`

  - feature：show version info

  - flag：this command does not need to have flag

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go version
    current version is : 1
    ```

- ### `wcp`

  - feature：change wallet password

  - flag：

    - -a：account which need change password
    - -p：old password
    - -n：new password

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go wcp -a qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic -p 123 -n 12345
    change password success for account: qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic
    ```

- ### `wc`

  - feature：create a wallet（random）

  - flag：

    - -p：set random wallet password

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go wc -p 12345
    create wallet: address=>qlc_13ysagt8qtbeidcoad4m7qisnkat115dyayt8gufgyhfi91zspcwnsrqyuoi, password=>12345 success
    ```

- ### `wi`

  - feature：import a wallet

  - flag：

    - -s：seed for  wallet which want import
    - -p：set a password for import wallet

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go wi -s DB68096C0E2D2954F59DA5DAAE112B7B6F72BE35FC96327FE0D81FD0CE5794A9 -p 12345
    import seed[DB68096C0E2D2954F59DA5DAAE112B7B6F72BE35FC96327FE0D81FD0CE5794A9] password[12345] => qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1e
    s5aobwed5x4pjojic success
    ```

- ### `wl`

  - feature：wallet list

  - flag：this command does not need to have flag

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go wl
    qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic
    qlc_3npj46zq4g4w4odndbei4m6sxdsqucuqgujzywgb8epjmd87kbjtq56e47hq
    ```

- ### `wr`

  - feature：remove a wallet

  - flag：

    - -a：which account do you want remove
    - -p：the password of removed account

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go wr -a qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic -p 12345
    remove wallet: qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic success
    ```

- ### `help`

  - feature：print help info

  - flag：no need

  - example：

    ```
    65967@DESKTOP-V0CNDGO MINGW64 /e/goproject/go-qlc/cmd (feature/cli)
    $ go run main.go help
    QLC Chain is the next generation public blockchain designed for the NaaS.
    
    Usage:
      QLCChain [flags]
      QLCChain [command]
    
    Available Commands:
      bc          block count
      help        Help about any command
      send        send transaction
      tl          token list
      version     show version info
      wc          create a wallet for QLCChain node
      wcp         change wallet password
      wi          import a wallet
      wl          wallet address list
      wr          remove wallet
    
    Flags:
      -a, --account string    wallet address,if is nil,just run a node
      -c, --config string     config file
      -h, --help              help for QLCChain
      -p, --password string   password for wallet
    
    Use "QLCChain [command] --help" for more information about a command.
    ```