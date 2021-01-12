package exporter

/*
  to run the tests with redis running on anything but localhost:6379 use
  $ go test   --redis.addr=<host>:<port>

  for html coverage report run
  $ go test -coverprofile=coverage.out  && go tool cover -html=coverage.out
*/

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strings"
	"testing"
)

func TestTile38(t *testing.T) {
	if os.Getenv("TEST_TILE38_URI") == "" {
		t.Skipf("TEST_TILE38_URI not set - skipping")
	}

	for _, isTile38 := range []bool{true, false} {
		e, _ := NewRedisExporter(os.Getenv("TEST_TILE38_URI"), Options{Namespace: "test", IsTile38: isTile38})

		chM := make(chan prometheus.Metric)
		go func() {
			e.Collect(chM)
			close(chM)
		}()

		found := false
		want := "tile38_threads_total"
		for m := range chM {
			if strings.Contains(m.Desc().String(), want) {
				found = true
			}
		}

		if isTile38 && !found {
			t.Errorf("%s was *not* found in tile38 metrics but expected", want)
		} else if !isTile38 && found {
			t.Errorf("%s was *found* in tile38 metrics but *not* expected", want)
		}
	}
}
