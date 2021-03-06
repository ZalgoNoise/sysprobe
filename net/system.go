package net

import (
	"strconv"
	"sync"

	"github.com/ZalgoNoise/sysprobe/utils"
)

// System type will contain this device's network information
type System struct {
	Device     string `json:"device"`
	ID         int    `json:"id"`
	IPAddress  string `json:"ipv4"`
	SubnetMask string `json:"mask"`
}

// Get method - runs a simple `ip` command to retrieve the
// current information from the active network device, which
// returns a pointer to the Internet struct with this data
func (s *System) Get(wg *sync.WaitGroup, ipDevice string) *System {

	defer wg.Done()
	ip, err := utils.Run("ip", "-f", "inet", "addr", "show", ipDevice)
	utils.Check(err)

	s.Device = utils.Splitter(string(ip), ": ", 1)

	devIndex, err := strconv.Atoi(utils.Splitter(string(ip), ": ", 0))
	utils.Check(err)
	s.ID = devIndex

	s.IPAddress = utils.Splitter(utils.Splitter(utils.Splitter(string(ip), "\n", 1), " ", 5), "/", 0)

	s.SubnetMask = utils.Splitter(utils.Splitter(string(ip), "\n", 1), " ", 7)

	return s

}
