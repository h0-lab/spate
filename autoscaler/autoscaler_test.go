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

package autoscaler_test

import (
	"testing"

	"github.com/docker/docker/api/types/swarm"
	"github.com/mtneug/spate/autoscaler"
	"github.com/mtneug/spate/metric"
	"github.com/stretchr/testify/require"
)

// TODO:

func TestNew(t *testing.T) {
	srv := swarm.Service{}
	observer := make([]metric.Observer, 0)

	a := autoscaler.New(srv, observer)
	require.Equal(t, srv, a.Service)
	require.Equal(t, observer, a.Observer)
}
