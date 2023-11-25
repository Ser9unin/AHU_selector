package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"pump"
	"static_bv"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

const K float32 = 0.86

var tpl *template.Template

type Single_AHU struct {
	AHU_Name         string
	LoadQ            int // measure watts
	Supply_T1        int // measure degrees Celsium
	Return_T2        int // measure degrees Celsium
	Pressure_loss_dP int // measure Pascal
	Connection_side  string
}

type Single_AUU struct {
	Unit_ID         string  `csv:"Unit_ID"`
	Connection_side string  `csv:"Connection_side"`
	Code_number     string  `csv:"Code"`
	AQT_type        string  `csv:"AQT_type"`
	AQT_set_min     uint8   `csv:"AQT_set_min"`
	AQT_dPmin       float32 `csv:"AQT_dPmin"` //measure kPa
	AQT_Gnom        float32 `csv:"Gnom"`      //measure l/h
	Tmax            string  `csv:"Tmax"`
	Actuator        string  `csv:"Actuator"`
	Static_valve    string  `csv:"Static_valve"`
	SBV_Kvs         int     `csv:"SBV_Kvs"` //measure l/h3
	SBV_setting     string  `csv:"Static_valve_setting"`
	DN              string  `csv:"DN"`
	Pump            string  `csv:"Pump"`
	Pump_setting    int     `csv:"Pump_setting"`
	AQT_setting     float32
}

func init() {
	tpl = template.Must(template.ParseGlob("web/templates/*.html"))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	http.Handle("/javascript/", http.StripPrefix("/javascript/", http.FileServer(http.Dir("./web/javascript"))))
	http.HandleFunc("/get_AHU", get_AHU)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "a_index.html", nil)
}

func get_AHU(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	AHU_bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", AHU_bytes)

	AHU_string := string(AHU_bytes)
	AHU_table := parse_AHU_table(AHU_string)

	AUU_unit_list, err := AUU_get_list(AHU_table)

	fmt.Println(err)

	if err != nil {
		// change this to flash error messages without rerendering page
		tpl.ExecuteTemplate(w, "a_index.html", err)
	} else {
		tpl.ExecuteTemplate(w, "ahulist.html", AUU_unit_list)
	}

	for _, element := range AHU_table {
		fmt.Println(element)
	}
	for _, element := range AUU_unit_list {
		fmt.Println(element)
	}
}

// get data from form with multiple inputs as []byte and put it in []Single_AHU structs
func parse_AHU_table(ahu_string string) []Single_AHU {
	var AHU_unit Single_AHU
	var AHU_table []Single_AHU
	var err error
	// cut first key-value pair from []byte on "&" symbol
	for ahu_string != "" {
		var key string
		key, ahu_string, _ = strings.Cut(ahu_string, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		// split first key-value pair on "=" symbol
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		// assign values to correct field of Single_AHU struct
		switch key {
		case "AHU_name":
			AHU_unit.AHU_Name = value
		case "Load":
			AHU_unit.LoadQ = str_to_int(value)
		case "T1":
			AHU_unit.Supply_T1 = str_to_int(value)
		case "T2":
			AHU_unit.Return_T2 = str_to_int(value)
		case "dPahu":
			AHU_unit.Pressure_loss_dP = str_to_int(value) * 1000
		case "Connection_side":
			AHU_unit.Connection_side = value
		}

		/*if AHU_unit.LoadQ < 0 {
			error placeholder
		}*/

		/*if AHU_unit.Supply_T1 <= AHU_unit.Return_T2 {
			error placeholder
		}*/

		/*if AHU_unit.Pressure_loss_dP < 0 {
			error placeholder
		}*/

		// put all Single_AHU into AHU_table
		if AHU_unit.Connection_side != "" {
			AHU_table = append(AHU_table, AHU_unit)
			AHU_unit.Connection_side = ""
		}
	}
	return AHU_table
}

/*func float_32(value string) float32 {
	var float_conv float32
	val, err := strconv.ParseFloat(value, 32)
	if err != nil {
		// keep calm and do nothing
	}
	float_conv = float32(val)
	return float_conv
}*/

func str_to_int(value string) int {
	var int_conv int
	int_conv, err := strconv.Atoi(value)
	if err != nil {
		// keep calm and do nothing
	}
	return int_conv
}

// creates list of control units based on list of airhandling units
func AUU_get_list(ahu_table []Single_AHU) ([]Single_AUU, error) {
	var auu_new Single_AUU
	var auu_table []Single_AUU
	var gcalc float32
	var auu_list []Single_AUU

	//csv unmarshall, it is done in order not to hardcode units in this file,
	//so anyone can update unit list
	auu_file, err := os.ReadFile("./tools/AUU_list2.csv")
	if err != nil {
		log.Fatal(err)
	}

	_ = gocsv.UnmarshalBytes(auu_file, &auu_list)

	//calculate flow based on Load in watts, flow measures in l/h
	//chose unit based on flow
	for _, element := range ahu_table {
		gcalc = float32((element.LoadQ / (element.Supply_T1 - element.Return_T2))) * K
		auu_new, err = AUU_calc(gcalc, element.Connection_side, auu_list, element.Pressure_loss_dP)
		auu_table = append(auu_table, auu_new)
	}
	return auu_table, err
}

// chosing unit based on flow,
func AUU_calc(gcalc float32, connection_side string, auu_list []Single_AUU, pressure_loss_dp int) (Single_AUU, error) {
	var chosen_unit Single_AUU
	var aqt_setting float32
	var found bool
	var pump_head int
	var errPump error

	// choose unit from list of units
	for _, element := range auu_list {

		if gcalc < element.AQT_Gnom && connection_side == element.Connection_side {
			found = true
			fmt.Println("FOUND!")
			fmt.Print(element.Unit_ID)
			// if found - calculate parameters of chosen unit AQT_setting, Pump_setting, SBV_setting
			if found {
				aqt_setting = gcalc / element.AQT_Gnom * 100
				element.AQT_setting = float32(math.Round(float64(aqt_setting)))
				gcalc_int := int(gcalc)
				sbv_Kvs_f32 := float32(element.SBV_Kvs)

				// for chosen AUU unit add pressure loss on fully open static valve, this how we will be sure that choosen pump will cover dP on AHU and dP on static valve
				addon_dp_sbv := (gcalc / sbv_Kvs_f32) * (gcalc / sbv_Kvs_f32) * 100000
				pressure_loss_dp += int(addon_dp_sbv)

				element.Pump_setting, pump_head, errPump = pump.Get_pump_setting(element.Pump, gcalc_int, pressure_loss_dp)

				if pump_head > pressure_loss_dp {
					element.SBV_setting = static_bv.Get_SBV_setting(element.Static_valve, gcalc_int, pressure_loss_dp, pump_head)
					chosen_unit = element
					return chosen_unit, nil
				} else {
					continue
				}
			}
		}
	}
	return Single_AUU{}, errPump
}
