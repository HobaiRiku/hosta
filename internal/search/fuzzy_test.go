package search

import "testing"

func TestFindUnicode(t *testing.T) {
	candidates := []string{"Home Server", "MT2500 路由器", "Development"}
	matches := Find("路由", candidates)
	if len(matches) != 1 || matches[0].Index != 1 {
		t.Fatalf("Find() = %#v, want candidate 1", matches)
	}
}

func TestFindEmptyKeepsInputOrder(t *testing.T) {
	matches := Find("", []string{"b", "a"})
	if len(matches) != 2 || matches[0].Index != 0 || matches[1].Index != 1 {
		t.Fatalf("Find() = %#v, want original order", matches)
	}
}

func TestMatchOneReturnsPositions(t *testing.T) {
	match, ok := MatchOne("hs", "Home Server")
	if !ok {
		t.Fatal("MatchOne() did not match")
	}
	if len(match.Positions) != 2 {
		t.Fatalf("positions = %#v, want two positions", match.Positions)
	}
}
