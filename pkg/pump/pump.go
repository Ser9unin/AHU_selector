package pump

import (
	"fmt"
	"log"
	"os"

	"github.com/gocarina/gocsv"
)

type Pump_setting struct {
	G    int `csv:"G"`
	Set3 int `csv:"Set3"`
	Set2 int `csv:"Set2"`
	Set1 int `csv:"Set1"`
}

func Get_pump_setting(pump_name string, gcalc, pressure_loss_dp int) (int, int) {
	// open file connected to pump name (each pump characteristic placed in different csv file in this project)
	var setting_range []Pump_setting
	var pump_setting Pump_setting
	pump_file_name := pump_name + ".csv"
	pump_file, err := os.ReadFile(pump_file_name)
	if err != nil {
		log.Fatal(err)
	}

	_ = gocsv.UnmarshalBytes(pump_file, &setting_range)

	len := len(setting_range)
	start_index := int(float32(gcalc) / float32(setting_range[len-1].G) * float32(len))
	if start_index >= len-1 {
		start_index = len - 2
	}

	for _, element := range setting_range[start_index:] {
		if gcalc <= element.G {
			found := true
			if found {
				pump_setting = element
				break
			}
		}
	}
	fmt.Println("Pump setting", pump_setting)

	switch {
	case pressure_loss_dp < pump_setting.Set1:
		return 1, pump_setting.Set1
	case pressure_loss_dp < pump_setting.Set2:
		return 2, pump_setting.Set2
	case pressure_loss_dp < pump_setting.Set3:
		return 3, pump_setting.Set3
	}
	return 0, 0
}
