#!/bin/bash
mv  go-ethereum/vendor ../
for i in  `find ./go-ethereum -name "ethereum"  -type d  |  awk '{print $NF}'  | awk -F "/ethereum" '{print $1}' ` ;do  mv  $i/ethereum $i/matrix  ;done
find ./go-ethereum -name "go-ethereum" -type d  | xargs -i mv {} ./go-ethereum/build/_workspace/pkg/linux_amd64/github.com/matrix/go-matrix
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i "s/ethereum\/go-matrix/matrix\/go-matrix/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i  "s/Ethereum/Matrix/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i  "s/ethereum/matrix/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i  "s/ETHEREUM/MATRIX/g"
find ./go-ethereum/ -name "*.go" -type f  | xargs sed -i "/\bgeth[ |.]/s/geth/gman/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i   "s/\bethash/manash/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i   "s/\bether/man/g"
find ./go-ethereum -name "*.go"  -type f  | xargs sed -i   "s/\beth/man/g"
find ./go-ethereum -name "eth"  -type  d  | xargs  -i  mv {}  ./go-ethereum/man
find ./go-ethereum -name "ethclient"  -type d  | xargs  -i mv {}  ./go-ethereum/manclient
find ./go-ethereum -name "ethdb"  -type d  | xargs  -i mv {}  ./go-ethereum/mandb
find ./go-ethereum -name "ethstats"  -type d  | xargs  -i mv {}  ./go-ethereum/manstats
find ./go-ethereum -name "ethash"  -type d |  xargs -i mv {} ./go-ethereum/consensus/manash
find ./go-ethereum -name "ethapi"  -type d  | xargs -i mv {} ./go-ethereum/internal/manapi
find ./go-ethereum -name "geth"  -type d | xargs -i mv {} ./go-ethereum/cmd/gman
find ./go-ethereum/ -name "*.go" -type f  | xargs  sed -i '/\bgeth/s/geth/gman/g'
find ./go-ethereum/ -name "*.go" -type f  | xargs  sed -i '/testgeth/s/geth/gman/g'
find ./go-ethereum/ -name "*.go" -type f  | xargs sed -i 
find ./go-ethereum/   -type f  | xargs sed -i  "/cmd\/geth/s/geth/gman/g"
sed -i 's/geth/gman/g' ./go-ethereum/Makefile
sed -i 's/ethereum/matrix/g' ./go-ethereum/build/env.sh
find  ./go-ethereum/ -name "*.go" -type f |  xargs  sed -i    "/\"ETH\"/s/ETH/MAN/g"
find  ./go-ethereum/ -name "*.go" -type f |  xargs  sed -i  "/Geth/s/Geth/Gman/g"
find ./go-ethereum -type f | xargs sed -i  "/\"geth\"/s/geth/gman/g"
sed -i '/Geth/s/Geth/Gman/g'  ./go-ethereum/node/config.go
sed -i "/\"Geth\"/s/Geth/Gman/g"  ./go-ethereum/build/ci.go
find ./go-ethereum/  -name "backend.go" | xargs sed -i '/Namespace/s/man/eth/g'
find ./go-ethereum -name "*.go"  -type f | xargs sed -i  "s/man_/eth_/g"
mv ../vendor ./go-ethereum/
find ./go-ethereum/vendor/ -name "matrix" | xargs -i mv {} ./go-ethereum/vendor/github.com/ethereum
for i in `find ./go-ethereum -path "./vendor" -a -prune -o -name "*.go" -print `;do sed -i  "s#eth_#man_#g" $i ;sed -i  "s#ethConf#manConf#g"  $i; sed -i  "s#ethApi#manApi#g"  $i; sed -i  "s#MANBackend#manBackend#g"  $i; sed -i  "s#ethapi#manapi#g"  $i;sed -i  "s#ethdb#mandb#g"  $i;sed -i  "s#ethcrypto#mancrypto#g"  $i; sed -i  "s#ethServ#manServ#g"  $i;done
for i in `find ./ -path "./vendor" -a -prune -o -name "*.go" -print `;do 
	a=`grep "Copyright (c) 2018 The MATRIX Authors " $i |wc -l`
	if [ $a -eq 0 ];then 
	sed -i "1i // Copyright (c) 2008 The MATRIX Authors " $i
	sed -i "2i // Distributed under the MIT software license, see the accompanying" $i
	sed -i "3i // file COPYING or or http://www.opensource.org/licenses/mit-license.php" $i
	fi
done
