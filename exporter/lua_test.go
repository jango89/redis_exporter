package exporter

import (
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestLuaScript(t *testing.T) {
	e := getTestExporter()

	for _, tst := range []struct {
		Script        string
		ExpectedKeys  int
		ExpectedError bool
	}{
		{
			Script:       `return {"a", "11", "b", "12", "c", "13"}`,
			ExpectedKeys: 3,
		},
		{
			Script:       `return {"key1", "6389"}`,
			ExpectedKeys: 1,
		},
		{
			Script:       `return {} `,
			ExpectedKeys: 0,
		},
		{
			Script:        `return {"key1"   BROKEN `,
			ExpectedKeys:  0,
			ExpectedError: true,
		},
	} {

		e.options.LuaScript = []byte(tst.Script)
		nKeys := tst.ExpectedKeys

		setupDBKeys(t, os.Getenv("TEST_REDIS_URI"))
		defer deleteKeysFromDB(t, os.Getenv("TEST_REDIS_URI"))

		chM := make(chan prometheus.Metric)
		go func() {
			e.Collect(chM)
			close(chM)
		}()
		scrapeErrorFound := false

		for m := range chM {
			if strings.Contains(m.Desc().String(), "test_script_value") {
				nKeys--
			}

			if strings.Contains(m.Desc().String(), "exporter_last_scrape_error") {
				g := &dto.Metric{}
				m.Write(g)
				if g.GetGauge() != nil && *g.GetGauge().Value > 0 {
					scrapeErrorFound = true
				}
			}
		}
		if nKeys != 0 {
			t.Error("didn't find expected script keys")
		}

		if tst.ExpectedError {
			if !scrapeErrorFound {
				t.Error("didn't find expected scrape errors")
			}
		}
	}
}
