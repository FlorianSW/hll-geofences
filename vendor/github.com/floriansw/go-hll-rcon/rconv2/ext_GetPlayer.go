package rconv2

import (
	"fmt"
	"math"
)

var (
	xs     = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	numpad = [][]int{
		{7, 8, 9},
		{4, 5, 6},
		{1, 2, 3},
	}
)

// Distance is supposed to be in centimeters (default unit of worlds in Unreal Engine)
type Distance float64

func (d Distance) Meters() float64 {
	return float64(d) / 100
}

func (d Distance) Add(o Distance) Distance {
	return d + o
}

func (w GetPlayerPosition) Equal(o GetPlayerPosition) bool {
	return w.X == o.X && w.Y == o.Y && w.Z == o.Z
}

// IsSpawned indicates that the player is currently not on the map, e.g. in the spawn or team selection screen.
func (w GetPlayerPosition) IsSpawned() bool {
	return (w.X + w.Y + w.Z) != 0
}

// Distance calculates the distance of this and another position in the game world. This includes movement on the x-axis
// (as represented in changed values of X and Y) as well as on the y-axis (represented by changed Z values).
// This is calculated as if the distance was traveled in a straight line without observing obstacles. It depends on the
// resolution of when the two involved positions were obtained how accurate the calculated distance is.
func (w GetPlayerPosition) Distance(o GetPlayerPosition) Distance {
	return Distance(math.Sqrt(math.Pow(float64(w.X-o.X), 2) + math.Pow(float64(w.Y-o.Y), 2) + math.Pow(float64(w.Z-o.Z), 2)))
}

func (w GetPlayerPosition) Grid(m DrawableMap) Grid {
	d := m.mapData()
	if d == nil {
		return Grid{}
	}
	x := float64(w.X) - d.MapCenterOffset.X
	y := float64(w.Y) - d.MapCenterOffset.Y
	xGrid, yGrid := math.Floor(x/d.SectorSize), math.Floor(y/d.SectorSize)

	xInGrid := x - xGrid*d.SectorSize
	yInGrid := y - yGrid*d.SectorSize
	num := d.SectorSize / 3
	return Grid{
		X:      xs[int(xGrid)+5],
		Y:      int(yGrid) + 6,
		Numpad: numpad[int(math.Abs(math.Floor(yInGrid/num)))][int(math.Abs(math.Floor(xInGrid/num)))],
	}
}

type Grid struct {
	X      string
	Y      int
	Numpad int
}

func (g Grid) String() string {
	return fmt.Sprintf("%s%d Numpad %d", g.X, g.Y, g.Numpad)
}

type DrawableMap interface {
	mapData() *mapData
}
