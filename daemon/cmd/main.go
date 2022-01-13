/*
   Copyright 2022 https://github.com/geebee

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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	ddns "github.com/geebee/ddns/daemon"
)

func dotEnv() error {
	file, err := os.Open("./.env")
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to load: .env: %w", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		kv := strings.SplitN(scanner.Text(), "=", 2)
		v := kv[1]
		if s, err := strconv.Unquote(v); err == nil {
			v = s
		}

		err = os.Setenv(kv[0], v)
		if err != nil {
			return fmt.Errorf("failed to set environment variable: %s = %s: %w", kv[0], v, err)
		}
	}

	return nil
}

func main() {
	if err := dotEnv(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	daemon := ddns.NewDynamicDNSFromEnv()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	daemon.Start()
	<-interrupt
	daemon.Stop()
}
