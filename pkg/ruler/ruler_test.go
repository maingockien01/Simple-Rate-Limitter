package ruler_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	rulerPkg "github.com/maingockien01/proxy/pkg/ruler"
	"github.com/stretchr/testify/require"
)

var onFetchCalled bool
var ruler *rulerPkg.Ruler

func setup() {
	onFetchCalled = false
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "../../configs/rules_test.json")
	ruler = rulerPkg.NewRuler(path, onFetch, 5*time.Second)
	ruler.FetchFile()
	//ruler.FetchFileAndPushRedisInterval()
}

func teardown() {
	if len(ruler.Rules) > 2 {
		ruler.Rules = ruler.Rules[:2]
	}
	jsonRules, _ := json.MarshalIndent(ruler.Rules, "", "\t")

	writeErr := os.WriteFile(ruler.Filepath, jsonRules, os.ModeExclusive)

	if writeErr != nil {
		panic(writeErr)
	}
}

func onFetch(ruler *rulerPkg.Ruler) {
	onFetchCalled = true
}

// Test fetch
func TestFetch(t *testing.T) {
	setup()
	defer teardown()

	ruler.FetchFile()

	require.True(t, onFetchCalled)

	require.Len(t, ruler.Rules, 2)
}

// Test fetch interval
func TestFetchInterval(t *testing.T) {
	setup()
	defer teardown()

	ruler.Rules = append(ruler.Rules, &rulerPkg.Rule{
		ApiPath:   "/whoami",
		MaxTokens: 5,
		Rate:      1,
	})

	jsonRules, _ := json.MarshalIndent(ruler.Rules, "", "\t")

	writeErr := os.WriteFile(ruler.Filepath, jsonRules, os.ModeExclusive)

	if writeErr != nil {
		panic(writeErr)
	}

	onFetchCalled = false
	time.Sleep(6 * time.Second)

	require.True(t, onFetchCalled)

	require.Len(t, ruler.Rules, 3)

}

func TestGetRule(t *testing.T) {
	setup()
	defer teardown()

	rule := ruler.GetRule("/whoami/kevin")

	require.NotNil(t, rule)

	require.Equal(t, rule.ApiPath, "/whoami")

	rule2 := ruler.GetRule("/whoareyou/kevin")

	require.NotNil(t, rule2)

	require.Equal(t, rule2.ApiPath, "/")
}
