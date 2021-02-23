package main

import "testing"

func BenchmarkTextMatcher(b *testing.B) {
	tm := NewTextMatcher(completeworks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Search("hamlet")
	}
}

func BenchmarkSuffixArrayMatcher(b *testing.B) {
	sm := NewSuffixArrayMatcher(completeworks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Search("hamlet")
	}
}

func BenchmarkSuffixArrayIgnoreCaseMatcher(b *testing.B) {
	sm := NewSuffixArrayIgnoreCaseMatcher(completeworks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Search("hamlet")
	}
}

func BenchmarkFuzzyMatcher(b *testing.B) {
	fm := NewFuzzyMatcher(completeworks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fm.Search("hamlet")
	}
}

func BenchmarkBleveMatcher(b *testing.B) {
	bm, _ := NewBleveMatcher(completeworks)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bm.Search("hamlet")
	}
}
