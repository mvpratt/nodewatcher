package health

import (
	"testing"
	"time"

	"github.com/lightninglabs/lndclient"
)

// Assumes:
//
//	githubTag is of the form "v0.15.5-beta"
//	lndVersionString is of the form "0.15.5-beta commit=v0.15.5-beta.f1"
func TestCompare(t *testing.T) {
	expected := true
	actual := compareVersions("v0.15.5-beta", "0.15.5-beta commit=v0.15.5-beta.f1")
	if actual != expected {
		t.Error("error")
	}

	expected = false
	actual = compareVersions("v0.16", "0.15.5-beta")
	if actual != expected {
		t.Error("error")
	}
}

func TestGenerateStatusMessage(t *testing.T) {
	info := &lndclient.Info{
		SyncedToChain:       false,
		SyncedToGraph:       true,
		Version:             "0.15.5-beta commit=v0.15.5-beta.f1",
		BestHeaderTimeStamp: time.Now(),
	}

	expectedResult := "\n\nWARNING: Lightning node is not fully synced."
	actualResult, actualError := generateStatusMessage(info)

	if expectedResult != actualResult {
		t.Error("error")
	}
	if actualError != nil {
		t.Error("error")
	}

}
