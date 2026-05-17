package rconv2

import "context"

// SetSectorLayoutBySlice same as SetSectorLayout but takes a slice of sectors (1st element sector 1, 2nd element sector 2
// and so on) instead of each of the sectors as a separate parameter. Mostly kept for compatibility reasons.
//
// Deprecated: Might be removed in a future release
func (c *Connection) SetSectorLayoutBySlice(ctx context.Context, sectors []string) error {
	r := SetSectorLayout{}
	for i, sector := range sectors {
		if i == 0 {
			r.Sector_1 = sector
		}
		if i == 1 {
			r.Sector_2 = sector
		}
		if i == 2 {
			r.Sector_3 = sector
		}
		if i == 3 {
			r.Sector_4 = sector
		}
		if i == 4 {
			r.Sector_5 = sector
		}
	}
	_, err := execCommand[SetSectorLayout, any](ctx, c.socket, r)
	return err
}
