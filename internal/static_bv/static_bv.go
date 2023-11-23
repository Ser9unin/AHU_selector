package static_bv

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/gocarina/gocsv"
)

type SBV_setting struct {
	Setting string `csv:"Setting"`
	KV      int    `csv:"KV"`
}

func Get_SBV_setting(sbv_name string, gcalc, pressure_loss_dp, pump_head int) string {
	// open file connected to Static valve name, each Static valve characteristic placed in different csv file in this project
	var setting_range []SBV_setting
	var valve SBV_setting
	var kv_required int
	var kv_required_f64 float64
	var sbv_dp float64

	sbv_dp = float64(pump_head-pressure_loss_dp) / 100000

	gcalc_f64 := float64(gcalc) / 1000

	kv_required_f64 = (gcalc_f64 / math.Sqrt(sbv_dp)) * 1000

	kv_required = int(kv_required_f64)

	fmt.Println("KV required", kv_required)

	sbv_file_name := "./tools/static_bv/" + sbv_name + ".csv"
	sbv_file, err := os.ReadFile(sbv_file_name)
	if err != nil {
		log.Fatal(err)
	}

	_ = gocsv.UnmarshalBytes(sbv_file, &setting_range)
	fmt.Println("SBV setting", setting_range)

	len := len(setting_range)

	if kv_required > setting_range[len-1].KV {
		// create error here
	}

	start_index := int(float32(kv_required) / float32(setting_range[len-1].KV) * float32(len))
	if start_index >= len-1 {
		start_index = len - 2
	}

	fmt.Println("Start index", start_index)

	for _, element := range setting_range[start_index:] {
		if kv_required <= element.KV {
			found := true
			if found {
				valve = element
				break
			}
		}
	}
	fmt.Println(valve)

	return valve.Setting
}
