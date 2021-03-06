// Package bat is a module to collect battery health information
// about your *nix device. Originally intended to collect battery metadata
// from Android smartphones, namely with root access to read the
// /sys/class/power_supply/battery/uevent file, however it is also ready
// to collect the same information without root access, referring to the
// termux-battery-status binary.
//
// On a Linux laptop, you can also read the battery information, although
// it would be recommended to change the pattern matching section in the
// Battery.Get() method.
//
// To run this module locally, you can try the following pattern by pointing
// to your local /sys/class/power_supply/* folder, as seen in message.go:
//
//	import (
//		"encoding/json"
//		"fmt"
//		"github.com/ZalgoNoise/sysprobe/bat"
//	)
//
//	func main() {
//		b := &bat.Battery{}
//		b = b.Get("battery") 	// points to /sys/class/power_supply/battery/uevent
//								// falls back to termux-battery-status if inaccessible
//		json, err := json.Marshal(b)
//		if err != nil {
//		fmt.Println(string(json))
//	}
//
package bat

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/ZalgoNoise/sysprobe/utils"
)

// Battery type will be converted to JSON
// containing important information for this module
type Battery struct {
	Status   string `json:"status"`
	Health   string `json:"health"`
	Capacity int    `json:"capacity"`
	Temp     struct {
		Internal float32 `json:"int"`
		Ambient  float32 `json:"ext"`
	} `json:"temp"`
	Source string `json:"source"`
}

// TermuxBattery struct serves as a fallback object in case the uevent
// file is not accessible to the user (root privileges are required)
// It is a slower approach and thus not being the primary way to get
// battery metadata
type TermuxBattery struct {
	Health   string  `json:"health"`
	Capacity int     `json:"percentage"`
	Plugged  string  `json:"plugged"`
	Status   string  `json:"status"`
	Temp     float32 `json:"temperature"`
	Current  int     `json:"current"`
}

// TermuxGet method for Battery objects will execute the
// `termux-battery-status` command, and retrieve the slice of bytes
// which are already a JSON object. The method will unmarshal the JSON
// object into a TermuxBattery object and its values pushed back to
// Battery
func (b *Battery) TermuxGet() *Battery {
	tb := &TermuxBattery{}
	exec, err := utils.Run("termux-battery-status")

	if err != nil {
		b.Source = "Failed to probe battery health"
		return b
	}

	json.Unmarshal(exec, tb)
	b.Capacity = tb.Capacity
	b.Health = tb.Health
	b.Status = tb.Status
	b.Temp.Internal = tb.Temp
	b.Source = "/data/data/com.termux/files/usr/bin/termux-battery-status"

	return b

}

// Get method - collects battery related values
// from /sys/class/power_supply/*/uevent, and returns a
// pointer to the Battery struct with this data
func (b *Battery) Get(batteryLoc string) *Battery {

	batteryFile := "/sys/class/power_supply/" + batteryLoc + "/uevent"

	bat, err := os.Open(batteryFile)
	if err != nil {
		bat.Close()
		b.TermuxGet()
		return b
	}
	defer bat.Close()

	b.Source = batteryFile

	scanner := bufio.NewScanner(bat)

	line := 0

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "STATUS=") {
			val := utils.Splitter(scanner.Text(), "=", 1)
			b.Status = val

		} else if strings.Contains(scanner.Text(), "HEALTH=") {
			val := utils.Splitter(scanner.Text(), "=", 1)
			b.Health = val

		} else if strings.Contains(scanner.Text(), "CAPACITY=") {
			val := utils.Splitter(scanner.Text(), "=", 1)
			intVal, err := strconv.Atoi(val)
			utils.Check(err)
			b.Capacity = intVal

		} else if strings.Contains(scanner.Text(), "TEMP=") {
			val := utils.Splitter(scanner.Text(), "=", 1)
			intVal, err := strconv.Atoi(val)
			utils.Check(err)
			var floatVal float32 = (float32(intVal) / 10)
			b.Temp.Internal = floatVal

		} else if strings.Contains(scanner.Text(), "TEMP_AMBIENT=") {
			val := utils.Splitter(scanner.Text(), "=", 1)
			intVal, err := strconv.Atoi(val)
			utils.Check(err)
			var floatVal float32 = (float32(intVal) / 10)
			b.Temp.Ambient = floatVal
		}

		line++
	}

	err = scanner.Err()
	utils.Check(err)

	return b

}
