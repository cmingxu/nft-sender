package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

var (
	targetAddrStr = flag.String("target", "UQAMLht_mIo-0OMkDLni6hr_fu01-BNSEd77f4ucVSHJllOA", "target address")
	nftAddrStr    = flag.String("nft", "EQB0FaEHs-hMvKQx7Zj-jMGy3VrvyXdezZrwEtbUHvNXRXnv", "nft address")
)

var api = func() ton.APIClientWrapped {
	client := liteclient.NewConnectionPool()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := client.AddConnectionsFromConfigUrl(ctx, "https://ton.org/global-config.json")
	if err != nil {
		panic(err)
	}

	return ton.NewAPIClient(client).WithRetry()
}()

var _seed = os.Getenv("WALLET_SEED")

func main() {
	flag.Parse()

	nftAddr := address.MustParseAddr(*nftAddrStr)
	newAddr := address.MustParseAddr(*targetAddrStr)

	seed := strings.Split(_seed, " ")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	ctx = api.Client().StickyContext(ctx)

	w, err := wallet.FromSeed(api, seed, wallet.V4R2)
	if err != nil {
		log.Fatal("FromSeed err:", err.Error())
	}

	log.Println("test wallet address:", w.Address())

	block, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		log.Fatal("CurrentMasterchainInfo err:", err.Error())
	}

	balance, err := w.GetBalance(ctx, block)
	if err != nil {
		log.Fatal("GetBalance err:", err.Error())
	}

	fmt.Println("Balance:", balance.String())

	if balance.Nano().Uint64() < 3000000 {
		log.Fatal("not enough balance", w.Address(), balance.String())
	}

	collectionAddr := address.MustParseAddr("EQDmkj65Ab_m0aZaW8IpKw4kYqIgITw_HRstYEkVQ6NIYCyW")
	collection := nft.NewCollectionClient(api, collectionAddr)
	collectionData, err := collection.GetCollectionData(ctx)
	if err != nil {
		panic(err)
	}

	nft := nft.NewItemClient(api, nftAddr)
	transferData, err := nft.BuildTransferPayload(newAddr, tlb.MustFromTON("0.01"), nil)
	if err != nil {
		log.Fatal("BuildMintPayload err:", err.Error())
	}

	fmt.Println("Transferring NFT...")
	transfer := wallet.SimpleMessage(nftAddr, tlb.MustFromTON("0.065"), transferData)

	fmt.Printf("%+v", transfer)
	fmt.Printf("%+v", collectionData)

	_, block, err = w.SendWaitTransaction(context.Background(), transfer)
	if err != nil {
		log.Fatal("Send err:", err.Error())
	}

	// wait next block to be sure everything updated
	block, err = api.WaitForBlock(block.SeqNo + 5).GetMasterchainInfo(ctx)
	if err != nil {
		log.Fatal("Wait master err:", err.Error())
	}

	newData, err := nft.GetNFTDataAtBlock(ctx, block)
	if err != nil {
		log.Fatal("GetNFTData err:", err.Error())
	}

	fullContent, err := collection.GetNFTContentAtBlock(ctx, collectionData.NextItemIndex, newData.Content, block)
	if err != nil {
		log.Fatal("GetNFTData err:", err.Error())
	}

	fmt.Printf("%+v\n", fullContent)

	roy, err := collection.RoyaltyParamsAtBlock(ctx, block)
	if err != nil {
		log.Fatal("RoyaltyParams err:", err.Error())
	}

	fmt.Println("Owner:", newData.OwnerAddress.String())
	// fmt.Println("Full content:", fullContent.(*nft.ContentOffchain).URI)
	fmt.Println("Royalty:", roy.Address.String(), roy.Base, "/", roy.Factor)

	if newData.OwnerAddress.String() != newAddr.String() {
		log.Fatal("nft owner not updated")
	}
}
