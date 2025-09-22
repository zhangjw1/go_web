package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	store "go-web-starter/contract"
	"log"
)

func main() {
	//deployByAbi()
	//deployByEthClient()
	//loadContract()
	callContractByABI()
}

func loadContract() {

	client, err := ethclient.Dial("https://sepolia.infura.io/v3/efe70fa30ffd4eb08d17e189753d80b9")
	if err != nil {
		log.Fatal(err)
	}
	contract, err := store.NewStore(common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"), client)
	if err != nil {
		log.Fatal(err)
	}
	_ = contract
	log.Println("Contract loaded")
	log.Printf("Contract StoreCaller:%v\n", contract.StoreCaller)
}
