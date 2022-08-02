/*
Copyright 2022 Google LLC.

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

package controllers

import (
	"bufio"
	"os"
	"strings"
)

// Helper function to check string exists in a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Helper function to remove string from a slice of string
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// Helper function to get the nameservers from /etc/resolv.conf
func getNameservers() (nameservers []string, err error) {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	nameservers = make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, " ")
		if parts[0] != "nameserver" {
			continue
		}

		n := strings.Join(parts[1:], "")
		n = strings.TrimSpace(n)
		nameservers = append(nameservers, n)
	}
	return nameservers, nil
}
