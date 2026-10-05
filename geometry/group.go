package geometry

import "github.com/takigo/takigo/window"

type Elementer interface {
	GeometryElements() []window.Windower
}

type Group []window.Windower

func (g Group) GeometryElements() []window.Windower {
	return g
}
