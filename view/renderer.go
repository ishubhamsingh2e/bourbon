// Package view defines view rendering abstractions and interfaces.
package view

import "io"

// Renderer defines the contract for rendering named templates/views with data to an output stream.
type Renderer interface {
	Render(w io.Writer, name string, data any) error
}
