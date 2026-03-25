package data_test

import (
	"github.com/floriansw/go-hll-rcon/rconv2/api"
	"github.com/floriansw/hll-geofences/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"log/slog"
	"os"
)

var _ = Describe("Config", func() {
	Describe("Persistence", func() {
		It("persists config change of server", func() {
			l := slog.New(slog.NewTextHandler(os.Stdout, nil))
			f, err := os.CreateTemp(os.TempDir(), "config")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())
			Expect(os.WriteFile(f.Name(), []byte("{}"), 0655)).ToNot(HaveOccurred())
			c, err := data.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.Save()).ToNot(HaveOccurred())
			c, err = data.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())

			Expect(c.Save()).ToNot(HaveOccurred())

			c, err = data.NewConfig(f.Name(), l)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Fence", func() {
		Context("Includes", func() {
			It("returns false when not includes", func() {
				Expect(data.Fence{
					X:       new("G"),
					Y:       new(4),
					Numpads: []int{4, 5, 6},
				}.Includes(api.Grid{X: "H", Y: 1, Numpad: 7})).To(BeFalse())
			})

			It("includes fence when direct match", func() {
				Expect(data.Fence{
					X:       new("G"),
					Y:       new(4),
					Numpads: []int{4, 5, 6},
				}.Includes(api.Grid{X: "G", Y: 4, Numpad: 5})).To(BeTrue())
			})

			It("includes fence when matching whole line X-axis", func() {
				Expect(data.Fence{
					X:       new("G"),
					Numpads: []int{4, 5, 6},
				}.Includes(api.Grid{X: "G", Y: 8, Numpad: 5})).To(BeTrue())
			})

			It("includes fence when matching whole line Y-axis", func() {
				Expect(data.Fence{
					Y:       new(5),
					Numpads: []int{4, 5, 6},
				}.Includes(api.Grid{X: "A", Y: 5, Numpad: 5})).To(BeTrue())
			})

			It("includes fence when no numpad specified", func() {
				Expect(data.Fence{
					X: new("G"),
					Y: new(5),
				}.Includes(api.Grid{X: "G", Y: 5, Numpad: 7})).To(BeTrue())
			})
		})

		Context("Matches", func() {
			var si *api.GetSessionResponse

			BeforeEach(func() {
				si = &api.GetSessionResponse{
					MapName:     "CARENTAN",
					GameMode:    "Warfare",
					PlayerCount: 40,
				}
			})

			It("matches when no condition", func() {
				Expect(data.Fence{}.Matches(si)).To(BeTrue())
			})

			It("map with same name and mode", func() {
				Expect(data.Fence{Condition: &data.Condition{
					Equals: map[string][]string{
						"map_name":  {si.MapName},
						"game_mode": {si.GameMode},
					},
				}}.Matches(si)).To(BeTrue())
			})

			It("does not match with wrong game mode", func() {
				Expect(data.Fence{Condition: &data.Condition{
					Equals: map[string][]string{
						"map_name":  {si.MapName},
						"game_mode": {"Skirmish"},
					},
				}}.Matches(si)).To(BeFalse())
			})

			It("does not match with wrong map name", func() {
				Expect(data.Fence{Condition: &data.Condition{
					Equals: map[string][]string{
						"map_name":  {"TOBRUK"},
						"game_mode": {si.GameMode},
					},
				}}.Matches(si)).To(BeFalse())
			})

			DescribeTable("when less than players than", func(pc int, expected bool) {
				Expect(data.Fence{Condition: &data.Condition{
					LessThan: map[string]int{
						"player_count": pc,
					},
				}}.Matches(si)).To(Equal(expected))
			},
				Entry("more players", 20, false),
				Entry("less players", 60, true),
				Entry("equal number of players", 40, false),
			)

			DescribeTable("when greater than players than", func(pc int, expected bool) {
				Expect(data.Fence{Condition: &data.Condition{
					GreaterThan: map[string]int{
						"player_count": pc,
					},
				}}.Matches(si)).To(Equal(expected))
			},
				Entry("more players", 20, true),
				Entry("less players", 60, false),
				Entry("equal number of players", 40, false),
			)
		})
	})
})
