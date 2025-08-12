package internal

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func InterpolateHexColors(a, b string, n int) []string {
	if n <= 0 {
		log.Fatal("n must be greater than 0")
	}

	r1, g1, b1 := parseHex(a)
	r2, g2, b2 := parseHex(b)

	out := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		t := 0.0
		if n == 1 {
			t = 0
		} else {
			t = float64(i-1) / float64(n-1)
		}
		r := uint8(math.Round(float64(r1) + t*float64(int(r2)-int(r1))))
		g := uint8(math.Round(float64(g1) + t*float64(int(g2)-int(g1))))
		bc := uint8(math.Round(float64(b1) + t*float64(int(b2)-int(b1))))
		out = append(out, fmt.Sprintf("#%02X%02X%02X", r, g, bc))
	}
	return out
}

func GetUser() (string, error) {
	gamicsDir := path.Join(".", ".gamics")
	if _, err := os.Stat(gamicsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("gamics directory does not exist, please register in first")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(gamicsDir)

	if err := viper.ReadInConfig(); err != nil {
		return "", fmt.Errorf("could not read config file: %w", err)
	}

	user := viper.GetString("logged-user")
	return user, nil
}

func parseHex(s string) (r, g, b uint8) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	switch len(s) {
	case 3:
		s = strings.ToUpper(s)
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
	default:
		log.Fatalf("expected format #RRGGBB or #RGB, got %s", s)
		return 0, 0, 0
	}

	rv, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		log.Fatalf("invalid red value in color %s: %s", s, err.Error())
	}
	gv, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		log.Fatalf("invalid green value in color %s: %s", s, err.Error())
	}
	bv, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		log.Fatalf("invalid blue value in color %s: %s", s, err.Error())
	}

	return uint8(rv), uint8(gv), uint8(bv)
}
