// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package panel

import (
	"github.com/mproffitt/bmx/pkg/components/optionlist"
	"github.com/mproffitt/bmx/pkg/kubernetes"
)

func (m *Model) getNamespaceList() (optionlist.Options, error) {
	return newNamespaceList(m.context, m.kubeconfig)
}

type namespaces struct {
	title      string
	context    string
	filename   string
	namespaces []string
}

func newNamespaceList(context, filename string) (*namespaces, error) {
	n := namespaces{
		title:    "Namespaces",
		context:  kubernetes.GetFullName(context, filename),
		filename: filename,
	}
	var err error
	n.namespaces, err = kubernetes.GetNamespaces(n.context, n.filename)
	return &n, err
}

func (n *namespaces) Title() string {
	return n.title
}

func (n *namespaces) Options() optionlist.Iterator {
	return func(yield func(key int, val optionlist.Row) bool) {
		func(yield func(key int, val optionlist.Row) bool) bool {
			for k, v := range n.namespaces {
				if !yield(k, optionlist.Option{Value: v}) {
					return false
				}
			}
			return true
		}(yield)
	}
}
