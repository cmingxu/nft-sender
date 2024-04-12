package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/xssnick/tonutils-go/address"
)

var (
	ownerAddrStr      = flag.String("owner", "UQAMLht_mIo-0OMkDLni6hr_fu01-BNSEd77f4ucVSHJllOA", "owner address")
	collectionAddrStr = flag.String("collection", "EQDmkj65Ab_m0aZaW8IpKw4kYqIgITw_HRstYEkVQ6NIYCyW", "collection address")
)

func main() {
	flag.Parse()

	ownerAddr := address.MustParseAddr(*ownerAddrStr)
	collectionAddr := address.MustParseAddr(*collectionAddrStr)

	url := fmt.Sprintf("https://tonapi.io/v2/accounts/%s/nfts?collection=%s&limit=1000&offset=0&indirect_ownership=false", ownerAddr, collectionAddr)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	content := string(body)

	value := gjson.Get(content, "nft_items.#.address")

	res := strings.Replace(value.String(), "[", "", -1)
	res = strings.Replace(res, "]", "", -1)
	res = strings.Replace(res, "\"", "", -1)

	for _, x := range strings.SplitN(res, ",", -1) {
		fmt.Println(x)
	}

}
