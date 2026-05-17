package rconv2

func (p GetPlayersPosition) ToGetPlayerPosition() GetPlayerPosition {
	return GetPlayerPosition{
		X: p.X,
		Y: p.Y,
		Z: p.Z,
	}
}
