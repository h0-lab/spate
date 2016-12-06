// Copyright © 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"fmt"
	"io"
	"text/template"
)

var printTemplate = `{{.Name}}:
  Version:        {{.Major}}.{{.Minor}}.{{.Patch}}+{{.GitCommit}}
  Git Commit:     {{.GitCommit}}
  Git Tree State: {{.GitTreeState}}
  Build Date:     {{.BuildDate}}
  Go Version:     {{.GoVersion}}
  Compiler:       {{.Compiler}}
  Platform:       {{.Platform}}
`

// Info holds version and build information. The fields are largly the same as
// in the `k8s.io/kubernetes/pkg/version` package of the Kubernetes project.
type Info struct {
	// Name of the versioned object.
	Name string `json:"name"`
	// Major version number.
	Major string `json:"major"`
	// Minor version number.
	Minor string `json:"minor"`
	// Patch version number.
	Patch string `json:"patch"`
	// GitCommit SHA.
	GitCommit string `json:"gitCommit"`
	// GitTreeState is either "clean" or "dirty".
	GitTreeState string `json:"gitTreeState"`
	// BuildDate of the binary.
	BuildDate string `json:"buildDate"`
	// GoVersion of the binary.
	GoVersion string `json:"goVersion"`
	// Compiler used for the binary.
	Compiler string `json:"compiler"`
	// Platform the binary is compiled for.
	Platform string `json:"platform"`
}

// String returns a formated version string.
func (i Info) String() string {
	return fmt.Sprintf("%v.%v.%v+%v", i.Major, i.Minor, i.Patch, i.GitCommit)
}

// PrintFull writes the version
func (i Info) PrintFull(w io.Writer) (err error) {
	t, err := template.New("info").Parse(printTemplate)
	if err != nil {
		return
	}
	err = t.Execute(w, i)
	return
}
