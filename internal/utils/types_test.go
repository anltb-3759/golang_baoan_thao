package utils

import "testing"

func TestNewPagination_FirstPage(t *testing.T) {
	p := NewPagination(1, 10, 25)
	if p.Page != 1 {
		t.Fatalf("expected page 1, got %d", p.Page)
	}
	if p.From != 1 {
		t.Fatalf("expected from 1, got %d", p.From)
	}
	if p.To != 10 {
		t.Fatalf("expected to 10, got %d", p.To)
	}
	if p.Total != 25 {
		t.Fatalf("expected total 25, got %d", p.Total)
	}
	if p.PrevPage != 0 {
		t.Fatalf("expected prev 0, got %d", p.PrevPage)
	}
	if p.NextPage != 2 {
		t.Fatalf("expected next 2, got %d", p.NextPage)
	}
	if p.HasPrev {
		t.Fatal("expected HasPrev false")
	}
	if !p.HasNext {
		t.Fatal("expected HasNext true")
	}
}

func TestNewPagination_LastPage(t *testing.T) {
	p := NewPagination(3, 10, 25)
	if p.From != 21 {
		t.Fatalf("expected from 21, got %d", p.From)
	}
	if p.To != 25 {
		t.Fatalf("expected to 25, got %d", p.To)
	}
	if p.PrevPage != 2 {
		t.Fatalf("expected prev 2, got %d", p.PrevPage)
	}
	if p.HasNext {
		t.Fatal("expected HasNext false on last page")
	}
	if !p.HasPrev {
		t.Fatal("expected HasPrev true on last page")
	}
}

func TestNewPagination_EmptyResult(t *testing.T) {
	p := NewPagination(1, 10, 0)
	if p.From != 0 {
		t.Fatalf("expected from 0 for empty result, got %d", p.From)
	}
	if p.To != 0 {
		t.Fatalf("expected to 0 for empty result, got %d", p.To)
	}
	if p.HasNext {
		t.Fatal("expected HasNext false for empty result")
	}
}

func TestNewPagination_SinglePage(t *testing.T) {
	p := NewPagination(1, 10, 5)
	if p.To != 5 {
		t.Fatalf("expected to 5, got %d", p.To)
	}
	if p.HasPrev {
		t.Fatal("expected HasPrev false")
	}
	if p.HasNext {
		t.Fatal("expected HasNext false")
	}
}

func TestNewPagination_MiddlePage(t *testing.T) {
	p := NewPagination(2, 10, 30)
	if p.From != 11 {
		t.Fatalf("expected from 11, got %d", p.From)
	}
	if p.To != 20 {
		t.Fatalf("expected to 20, got %d", p.To)
	}
	if p.PrevPage != 1 {
		t.Fatalf("expected prev 1, got %d", p.PrevPage)
	}
	if p.NextPage != 3 {
		t.Fatalf("expected next 3, got %d", p.NextPage)
	}
	if !p.HasPrev {
		t.Fatal("expected HasPrev true")
	}
	if !p.HasNext {
		t.Fatal("expected HasNext true")
	}
}
