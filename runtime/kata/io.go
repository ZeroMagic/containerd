/*
   Copyright The containerd Authors.

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

package kata

import (
	"context"
	"fmt"
	"net"

	"github.com/containerd/containerd/runtime"
)

type pipeSet struct {
	src    runtime.IO
	stdin  net.Conn
	stdout net.Conn
	stderr net.Conn
}

// NewIO connects to the provided pipe addresses
func newPipeSet(ctx context.Context, io runtime.IO) (*pipeSet, error) {
	
	fmt.Println("pipe is not implemented")

	return &pipeSet{}, nil
}

// Close terminates all successfully dialed IO connections
func (p *pipeSet) Close() {
	for _, cn := range []net.Conn{p.stdin, p.stdout, p.stderr} {
		if cn != nil {
			cn.Close()
		}
	}
}
