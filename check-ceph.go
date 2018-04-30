/*
 * Copyright (C) 2018  Maximilian Falkenstein <mfalkenstein@sos.ethz.ch>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	RC_OK      = 0
	RC_WARN    = 1
	RC_CRIT    = 2
	RC_UNKNOWN = 3
)

const (
	CSTR_WARN = "HEALTH_WARN"
	CSTR_CRIT = "HEALTH_CRIT"
	CSTR_OK   = "HEALTH_OK"
)

func main() {
	fileName := flag.String("from-file", "", "Read ceph status from file instead of invoking ceph. Useful for debugging.")
	sudo := flag.Bool("sudo", true, "Use sudo to execute ceph")
	flag.Parse()

	var dec *json.Decoder = nil

	if *fileName != "" {
		if f, err := os.Open(*fileName); err == nil {
			reader := bufio.NewReader(f)
			dec = json.NewDecoder(reader)
		} else {
			fmt.Println("UNKNOWN - Couldn't open specified file: ", err)
			os.Exit(RC_UNKNOWN)
		}
	} else {
		// Run CEPH directly
		var cephCmd *exec.Cmd
		if *sudo {
			cephCmd = exec.Command("sudo", "ceph", "status", "--format" ,"json-pretty")
		} else {
			cephCmd = exec.Command("ceph", "status", "--format" ,"json-pretty")
		}
		if cephOut, err := cephCmd.Output(); err != nil {
			fmt.Println("UNKNOWN - Couldn't check ceph")
			fmt.Println(err)
			os.Exit(RC_UNKNOWN)
		} else {
			dec = json.NewDecoder(strings.NewReader(string(cephOut)))
		}
	}

	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		fmt.Println(err)
		return
	}

	rc := RC_UNKNOWN
	if val, ok := v["health"]; ok {
		if valMap, ok := val.(map[string]interface{}); ok {
			// status should always be available
			if val, ok := valMap["status"]; ok {
				if val == CSTR_OK {
					fmt.Println("CEPH OK")
					rc = RC_OK
				} else if val == CSTR_WARN {
					fmt.Println("CEPH WARNING - Cluster operates with warnings, see below")
					rc = RC_WARN
				} else if val == CSTR_CRIT {
					fmt.Println("CEPH CRITICAL - Cluster inoperative")
					rc = RC_CRIT
				} else {
					fmt.Println("UNKNOWN - Invalid JSON format")
					rc = RC_UNKNOWN
				}
			} else {
				fmt.Println("UNKNOWN - Invalid JSON format")
				os.Exit(RC_UNKNOWN)
			}
			// If the status isn't okay, checks should contain the reason(s)
			if val, ok := valMap["checks"]; ok {
				if valMap, ok := val.(map[string]interface{}); ok {
					for key := range valMap {
						fmt.Print(key, ": ")
						if val, ok := valMap[key].(map[string]interface{}); ok {
							if valMap, ok := val["summary"].(map[string]interface{}); ok {
								fmt.Println(valMap["message"])
							}
						}
					}
				} else {
					fmt.Println("UNKNOWN - Invalid JSON format")
					os.Exit(RC_UNKNOWN)
				}
			}
		} else {
			fmt.Println("UNKNOWN - Invalid JSON format")
			os.Exit(RC_UNKNOWN)
		}
	} else {
		fmt.Println("UNKNOWN - JSON didn't contain health information")
		os.Exit(RC_UNKNOWN)
	}
	os.Exit(rc)
}
