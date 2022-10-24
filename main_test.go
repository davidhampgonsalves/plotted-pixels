package main

import (
	"testing"

	"image/color"
)

func Equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	return true
}

func ComparePositions(a, b [][]int, t *testing.T) {
	if len(a) != len(b) {
		t.Errorf("position expected / actual lengths are different: %v vs. %v.", a, b)
	}

	for i := range a {
		for j := range a[0] {
			if a[i][j] != b[i][j] {
				t.Errorf("%v - actual: %v does not equal expected: %v.", i, a, b)
			}
		}
	}
}

//func TestGetStartPositions(t *testing.T) {
//bitmask := [][]bool {
//{true, false, true, false},
//{false, true, false, false},
//{false, false, true, false},
//{false, false, true, true},
//}

//startPositions := GetStartPositions(0, 0, bitmask, CloneBitMask(bitmask))
//expectedPositions := [][]int {
//{0, 2},
//{3, 3},
//{3, 2},
//{0, 0},
//}

//ComparePositions(startPositions, expectedPositions, t)
//}

//func TestSimpleGetStartPositions(t *testing.T) {
//bitmask := [][]bool {
//{true, false},
//{false, true},
//}

//startPositions := GetStartPositions(0, 0, bitmask, CloneBitMask(bitmask))
//expectedPositions := [][]int {
//{1, 1},
//{0, 0},
//}

//ComparePositions(startPositions, expectedPositions, t)
//}

func TestLoadAndStartPositions(t *testing.T) {
	img := Load("test.4x4.png")

	bitmask := GenerateBitmask(img, color.NRGBA{0x00, 0x80, 0xf8, 0xff})

	startPositions := GetStartPositions(3, 0, bitmask, CloneBitMask(bitmask))

	expectedPositions := [][]int{
		{0, 1},
		{3, 2},
		{2, 3},
		{1, 2},
		{2, 1},
		{3, 0},
	}
	ComparePositions(startPositions, expectedPositions, t)
}
