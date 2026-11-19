package jianfan

import "testing"

func TestT2S(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"漢字", "汉字"},
		{"臺灣", "台湾"},
		{"abc123", "abc123"},
		{"", ""},
	}
	for _, c := range cases {
		if got := T2S(c.in); got != c.want {
			t.Fatalf("T2S(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestS2T(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"汉字", "漢字"},
		{"台湾", "臺灣"},
	}
	for _, c := range cases {
		if got := S2T(c.in); got != c.want {
			t.Fatalf("S2T(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
