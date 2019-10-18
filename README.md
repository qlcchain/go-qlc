<div align="center">
    <img src="assets/logo.png" alt="Logo" width='auto' height='auto'/>
</div>

---

[![Actions Status](https://github.com/qlcchain/go-qlc/workflows/Main%20workflow/badge.svg)](https://github.com/qlcchain/go-qlc/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/qlcchain/go-qlc)](https://goreportcard.com/report/github.com/qlcchain/go-qlc)
[![codecov](https://codecov.io/gh/qlcchain/go-qlc/branch/master/graph/badge.svg)](https://codecov.io/gh/qlcchain/go-qlc)

---

QLC Chain is a next generation public blockchain designed for the NaaS. It deploys a multidimensional Block Lattice architecture and uses virtual machines (VM) to manage and support integrated Smart Contract functionality. Additionally, QLC Chain utilizes dual consensus: Delegated Proof of Stake (DPoS), and Shannon Consensus, which is a novel consensus developed by the QLC Chain team. Through the use of this dual consensus protocol, QLC Chain is able to deliver a high number of transactions per second (TPS), massive scalability and an inherently decentralized environment for NaaS related decentralized applications (dApp). The framework of QLC Chain will enable everyone to operate network services and benefit from it.

Network-as-a-Service (NaaS) is sometimes listed as a separate cloud provider along with Infrastructure- as-a-Service (IaaS), Platform-as-a-Service (PaaS), and Software-as-a-Service (SaaS).
This factors out networking, firewalls, related security, etc.

NaaS can include flexible and extended Virtual Private Network (VPN), bandwidth on demand, custom routing, multicast protocols, security firewall, intrusion detection and prevention, Wide Area Network (WAN), content addressing and filtering, and antivirus.

---

## Key Features

* Multidimensional Block Lattice Structure 
* QLC Chain Smart Contract 
* Dual Consensus Protocol 
    > For more information, see [YellowPaper](https://github.com/qlcchain/YellowPaper). 
    
## Build and Run
```shell
make clean build
./gqlc
```

## Docker
You can build the docker image yourself or download it from docker hub
### Build docker images

```bash
cd docker
./build.sh
```

### Download docker images from docker hub

```bash
docker pull qlcchain/go-qlc:latest
```

### Start docker container
You can choose to run a normal node without an account or run an account node.

#### Run a normal node without an account

```bash
docker container run -d --name go-qlc \
    -p 9734:9734 \
    -p 127.0.0.1:9735:9735 \
    -p 127.0.0.1:9736:9736 \
    qlcchain/go-qlc:latest
```

#### Run an account node
You only need to assign a value to the environment variable `seed` to run the account node

```bash
docker container run -d --name go-qlc \
    -p 9734:9734 \
    -p 127.0.0.1:9735:9735 \
    -p 127.0.0.1:9736:9736 \
    qlcchain/go-qlc:latest --seed=B4F6494E3DD8A036EFF547C0293055B2A0644605DE4D9AC91B45343CD0E0E559
```
#### Run node by Docker Compose

- create `docker-compose.yml`

    ```yml
    version: "3.5"

    services:
    qlcchain_node:
        image: qlcchain/go-qlc:${version}
        container_name: qlcchain_node
        command: ["--configParams=rpc.rpcEnabled=true", "--seed=B4F6494E3DD8A036EFF547C0293055B2A0644605DE4D9AC91B45343CD0E0E559", "--nobootnode=true"]
        ports:
        - "9734:9734"
        - "9735:9735"
        - "127.0.0.1:9736:9736"
        networks:
        - qlcchain
        volumes:
        - type: bind
            source: ./data/
            target: /root/.gqlcchain/
        restart: unless-stopped
    
    networks:
    qlcchain:
        name: qlcchain

    ```
- run 
    ```bash
    docker-compose down -v && docker-compose up -d
    ```

## Contributions

We love reaching out to the open-source community and are open to accepting issues and pull-requests.

For all code contributions, please ensure they adhere as close as possible to the [contributing guidelines](CONTRIBUTING.md)

If you...

1. love the work we are doing,
2. want to work full-time with us,
3. or are interested in getting paid for working on open-source projects

... **we're hiring**.

To grab our attention, just make a PR and start contributing.

## Links & Resources
* [Yellow Paper](https://github.com/qlcchain/YellowPaper)
* [QLC Website](https://qlcchain.org)
* [Reddit](https://www.reddit.com/r/QLCChain/)
* [Medium](https://medium.com/qlc-chain)
* [Twitter](https://twitter.com/QLCchain)
* [Telegram](https://t.me/qlinkmobile)
