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

package ddns

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func externalIP(lookupURL string) (string, error) {
	resp, err := http.Get(lookupURL)
	if err != nil {
		return "", fmt.Errorf("GET request failed: %s: %w", lookupURL, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
