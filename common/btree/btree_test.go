package btree

import (
	"fmt"
	"testing"
)

type item struct {
	num int
}

var visitor ItemIterator = func(it Item) bool {
	fmt.Println(it.(*item).num)
	return true
}

func (i *item) Less(other Item) bool {
	return i.num < other.(*item).num
}

func createItems(length int) items {
	its := make(items, length)
	for i := 0; i < length; i++ {
		its[i] = &item{num: i * 3}
	}
	return its
}

func TestBTree_Insert(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	if tree.Length() != len(its) {
		t.Errorf("the size of the tree is not right, should be %d", len(its))
	}
}

func TestBTree_Delete(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	tree.Delete(&item{num: 4})
	if tree.Length() != len(its) {
		t.Errorf("the size of the tree is not right, should be %d", len(its))
	}
	tree.Delete(&item{num: 9})
	if tree.Length() != len(its)-1 {
		t.Errorf("the size of the tree is not right, should be %d", len(its)-1)
	}
	minimum := tree.Min()
	_it := tree.DeleteMin()
	if _it.(*item).num != minimum.(*item).num {
		t.Errorf("should be %d", minimum.(*item).num)
	}
	maximum := tree.Max()
	_it = tree.DeleteMax()
	if _it.(*item).num != maximum.(*item).num {
		t.Errorf("should be %d", maximum.(*item).num)
	}
	if tree.Length() != len(its)-3 {
		t.Errorf("the size of the tree is not right, should be %d", len(its)-3)
	}
}

func TestBTree_Ascend(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	tree.Ascend(visitor)
}

func TestBTree_AscendRange(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	start := &item{num: 12}
	end := &item{num: 99}
	tree.AscendRange(start, end, visitor)
}

func TestBTree_AscendFromFirstToPivot(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	end := &item{num: 30}
	tree.AscendFromFirstToPivot(end, visitor)
}

func TestBTree_AscendFromPivotToLast(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	start := &item{num: 168}
	tree.AscendFromPivotToLast(start, visitor)
}

func TestBTree_Descend(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	tree.Descend(visitor)
}

func TestBTree_DescendRange(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	start := &item{num: 27}
	end := &item{num: 42}
	tree.DescendRange(end, start, visitor)
}

func TestBTree_DescendFromPivotToFirst(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	end := &item{num: 27}
	tree.DescendFromPivotToFirst(end, visitor)
}

func TestBTree_DescendFromLastToPivot(t *testing.T) {
	tree := New(4)
	its := createItems(64)
	for _, it := range its {
		tree.Insert(it)
	}
	start := &item{num: 166}
	tree.DescendFromLastToPivot(start, visitor)
}

func TestBTree_Delete2(t *testing.T) {
	tree := New(4)
	its := createItems(512)
	for _, it := range its {
		tree.Insert(it)
	}
	for i := 0; i < 2000; i++ {
		it := &item{num: i}
		tree.Delete(it)
	}
	if tree.Length() != 0 {
		t.Errorf("after deleteing all items, there should be no item in tree")
	}
}
