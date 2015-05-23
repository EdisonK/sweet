package sweet

import (
	"fmt"
	"strings"
)

type Cisco struct {
}

func newCiscoCollector() Collector {
	return Cisco{}
}

func (collector Cisco) Collect(device DeviceConfig, c *Connection) (CollectionResults, error) {
	result := CollectionResults{}
	if _, err := expectSaveTimeout("assword:", c.Receive, device.Timeout); err != nil {
		return result, fmt.Errorf("Missing password prompt: %s", err.Error())
	}
	c.Send <- device.Config["pass"] + "\n"
	multi := []string{"#", ">", "assword:"}
	m, err := expectMulti(multi, c.Receive)
	if err != nil {
		return result, fmt.Errorf("Invalid response to password: %s", err.Error())
	}
	if m == "assword:" {
		return result, fmt.Errorf("Bad username or password.")
	} else if m == ">" {
		c.Send <- "enable\n"
		if err := expect("assword:", c.Receive); err != nil {
			return result, fmt.Errorf("Missing enable password prompt: %s", err.Error())
		}
		c.Send <- device.Config["enable"] + "\n"
		if err := expect("#", c.Receive); err != nil {
			return result, fmt.Errorf("Enable attempt failed: %s", err.Error())
		}
	}
	c.Send <- "terminal length 0\n"
	if err := expect("#", c.Receive); err != nil {
		return result, fmt.Errorf("Command 'terminal length 0' failed: %s", err.Error())
	}
	c.Send <- "terminal pager 0\n"
	if err := expect("#", c.Receive); err != nil {
		return result, fmt.Errorf("Command 'terminal pager 0' failed: %s", err.Error())
	}
	c.Send <- "show running-config\n"
	result["config"], err = expectSave("#", c.Receive)
	if err != nil {
		return result, fmt.Errorf("Command 'show running-config' failed: %s", err.Error())
	}

	// cleanup config results
	result["config"] = strings.TrimSpace(strings.TrimPrefix(result["config"], "show running-config"))
	result["config"] = strings.TrimSpace(strings.TrimPrefix(result["config"], "Building configuration..."))

	c.Send <- "exit\n"
	return result, nil
}
