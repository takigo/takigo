package geometry

import "github.com/msorc/takigo/window"

type Elementer interface {
	GeometryElements() []window.Windower
}

type Group []window.Windower

func (g Group) GeometryElements() []window.Windower {
	return g
}
