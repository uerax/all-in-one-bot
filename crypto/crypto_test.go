package crypto

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewCrypto(t *testing.T) {
	s := []string{"2312"}
	symbols := fmt.Sprintf(`?symbols=["%s"]`, strings.Join(s, `","`))

	fmt.Print(symbols)
	teee()
}


func teee() (t map[string]string) {
	t["t"] = "t"
	return t
}