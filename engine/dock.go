package engine

import (
	"github.com/andrewloable/go-fastreport/report"
)

// dockable is satisfied by any component that exposes both dock geometry
// getters and setters (i.e. embeds report.ComponentBase).
type dockable interface {
	report.Base
	Left() float32
	Top() float32
	Width() float32
	Height() float32
	SetLeft(float32)
	SetTop(float32)
	SetWidth(float32)
	SetHeight(float32)
	Dock() report.DockStyle
}

// applyDockLayout adjusts the positions and sizes of objects in objs according
// to their Dock property, using containerW x containerH as the available
// space.  Objects are processed in collection order, which determines docking
// priority (earlier objects claim their edges first).
//
// The algorithm mirrors WinForms DockStyle semantics:
//   - DockTop    — object is placed at the top of the remaining rect; its width
//     is stretched to fill the remaining horizontal space.
//   - DockBottom — placed at the bottom; width stretched.
//   - DockLeft   — placed at the left; height stretched to fill remaining vertical space.
//   - DockRight  — placed at the right; height stretched.
//   - DockFill   — placed at and sized to the entire remaining rect.
//   - DockNone   — object is left untouched.
//
// The function mutates object geometry in place so that buildPreparedObject
// picks up the docked coordinates when it reads Left/Top/Width/Height.
func applyDockLayout(objs *report.ObjectCollection, containerW, containerH float32) {
	if objs == nil || objs.Len() == 0 {
		return
	}

	// "Remaining rect" starts as the full container.
	remLeft := float32(0)
	remTop := float32(0)
	remRight := containerW
	remBottom := containerH

	for i := 0; i < objs.Len(); i++ {
		obj := objs.Get(i)
		d, ok := obj.(dockable)
		if !ok {
			continue
		}

		switch d.Dock() {
		case report.DockTop:
			h := d.Height()
			d.SetLeft(remLeft)
			d.SetTop(remTop)
			d.SetWidth(remRight - remLeft)
			// Height is kept as-is; only Left/Top/Width are adjusted.
			remTop += h

		case report.DockBottom:
			h := d.Height()
			d.SetLeft(remLeft)
			d.SetTop(remBottom - h)
			d.SetWidth(remRight - remLeft)
			remBottom -= h

		case report.DockLeft:
			w := d.Width()
			d.SetLeft(remLeft)
			d.SetTop(remTop)
			d.SetHeight(remBottom - remTop)
			// Width is kept as-is; only Left/Top/Height are adjusted.
			remLeft += w

		case report.DockRight:
			w := d.Width()
			d.SetLeft(remRight - w)
			d.SetTop(remTop)
			d.SetHeight(remBottom - remTop)
			remRight -= w

		case report.DockFill:
			d.SetLeft(remLeft)
			d.SetTop(remTop)
			d.SetWidth(remRight - remLeft)
			d.SetHeight(remBottom - remTop)
			// DockFill consumes the entire remaining space; nothing left.
			remLeft = remRight
			remTop = remBottom

		default: // DockNone — leave object untouched
		}
	}
}
