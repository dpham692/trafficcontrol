package orttest

/*
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	"github.com/apache/trafficcontrol/cache-config/testing/ort-tests/tcdata"
	"github.com/apache/trafficcontrol/cache-config/testing/ort-tests/util"
	"os"
	"testing"
	"time"
)

var (
	base_line_dir   = "baseline-configs"
	test_config_dir = "/opt/trafficserver/etc/trafficserver"

	testFiles = [8]string{
		"astats.config",
		"hdr_rw_first_ds-top.config",
		"hosting.config",
		"parent.config",
		"records.config",
		"remap.config",
		"storage.config",
		"volume.config",
	}
)

func TestT3cBadassAndSyncDs(t *testing.T) {
	fmt.Println("------------- Starting TestT3cBadassAndSyncDs ---------------")
	tcd.WithObjs(t, []tcdata.TCObj{
		tcdata.CDNs, tcdata.Types, tcdata.Tenants, tcdata.Parameters,
		tcdata.Profiles, tcdata.ProfileParameters, tcdata.Statuses,
		tcdata.Divisions, tcdata.Regions, tcdata.PhysLocations,
		tcdata.CacheGroups, tcdata.Servers, tcdata.Topologies,
		tcdata.DeliveryServices}, func() {

		// run badass and check config files.
		err := runApply("atlanta-edge-03", "badass")
		if err != nil {
			t.Fatalf("ERROR: t3c badass failed: %v\n", err)
		}
		for _, v := range testFiles {
			bfn := base_line_dir + "/" + v
			if !util.FileExists(bfn) {
				t.Fatalf("ERROR: missing baseline config file, %s,  needed for tests", bfn)
			}
			tfn := test_config_dir + "/" + v
			if !util.FileExists(tfn) {
				t.Fatalf("ERROR: missing the expected config file, %s", tfn)
			}

			diffStr, err := util.DiffFiles(bfn, tfn)
			if err != nil {
				t.Fatalf("diffing %s and %s: %v", tfn, bfn, err)
			} else if diffStr != "" {
				t.Errorf("%s and %s differ: %v", tfn, bfn, diffStr)
			} else {
				t.Logf("%s and %s diff clean", tfn, bfn)
			}
		}

		time.Sleep(time.Second * 5)

		fmt.Println("------------------------ Verify Plugin Configs ----------------")
		err = runVerify("/opt/trafficserver/etc/trafficserver/remap.config")
		if err != nil {
			t.Errorf("Plugin verification failed for remap.config: " + err.Error())
		}
		err = runVerify("/opt/trafficserver/etc/trafficserver/plugin.config")
		if err != nil {
			t.Errorf("Plugin verification failed for plugin.config: " + err.Error())
		}

		fmt.Println("----------------- End of Verify Plugin Configs ----------------")

		fmt.Println("------------------------ running SYNCDS Test ------------------")
		// remove the remap.config in preparation for running syncds
		remap := test_config_dir + "/remap.config"
		err = os.Remove(remap)
		if err != nil {
			t.Fatalf("ERROR: unable to remove %s\n", remap)
		}
		// prepare for running syncds.
		err = ExecTOUpdater("atlanta-edge-03", false, true)
		if err != nil {
			t.Fatalf("ERROR: queue updates failed: %v\n", err)
		}

		// remap.config is removed and atlanta-edge-03 should have
		// queue updates enabled.  run t3c to verify a new remap.config
		// is pulled down.
		err = runApply("atlanta-edge-03", "syncds")
		if err != nil {
			t.Fatalf("ERROR: t3c syncds failed: %v\n", err)
		}
		if !util.FileExists(remap) {
			t.Fatalf("ERROR: syncds failed to pull down %s\n", remap)
		}
		fmt.Println("------------------------ end SYNCDS Test ------------------")

	})
	fmt.Println("------------- End of TestT3cBadassAndSyncDs ---------------")
}
