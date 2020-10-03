package algorand

import (
	"fmt"
	"testing"

	"github.com/eventers/velvet/pkg/utl/log"
	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	eventers := Account{
		AccountAddress:     "7CENFGP7MKSGYHGXXXVH4LRDHKJPJ7CXWOCG4ULHTNI3QEBMG63HWVTEYE",
		SecurityPassphrase: "empty soul grass suspect license bachelor upset sing rice mother path scatter box zebra moon artwork dolphin jungle story file develop crawl remain about loyal",
	}
	to := Account{
		AccountAddress:     "ZEQGER6MQOIOFUP6YULNPNOFHLUVXLEAL3GBYVMP7WHFPMEBUA6MX5I7MI",
		SecurityPassphrase: "endorse hybrid salon worth glory shuffle fossil explain adapt true slot neutral vivid error anchor nerve unit purchase ride then vocal enforce enemy able today",
	}
	from := Account{
		AccountAddress:     "N4LAE5ITDYDMKBCUTAEI54L5GMQ5TUIBSLEKWLPVNE2AW5QZ4ZMOLLDLSQ",
		SecurityPassphrase: "poverty wide soccer dance wink sad fold chase pulp swap almost wool remind cable police say property gown exotic bacon allow basket always able wrist",
	}
	l := log.New()
	a := New(&eventers, "https://testnet-algorand.api.purestake.io/ps1", "LDV76UoaH15icurAUz6Hd3CvmfQpKRZj8CkoYUM2", 1000000, 1000, 100, l)
	err := a.Send(&to, 100)
	assert.Nil(t, err)

	assetID, err := a.CreateAsset(&from)
	fmt.Printf("%d, %+v", assetID, err)

	err = a.OptIn(&to, assetID)
	fmt.Printf("%+v", err)

	err = a.SendAsset(&from, &to, assetID)
	fmt.Printf("%+v", err)
}
