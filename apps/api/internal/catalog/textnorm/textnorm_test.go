package textnorm

import (
	"reflect"
	"testing"
)

func TestTokens(t *testing.T) {
	got := Tokens("Air Fryer Mondial 5 Litros — Preta")
	want := []string{"air", "fryer", "mondial", "5l", "preta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got := Tokens("Panela elétrica de arroz"); !reflect.DeepEqual(got, []string{"panela", "eletrica", "arroz"}) {
		t.Fatalf("got %v", got)
	}
}
