package datastructures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoDuplicates(t *testing.T) {
	set := NewSet[string]()
	set.Add("test")
	set.Add("test")
	assert.True(t, set.Contains("test"), "Set should contain 'test'")
	assert.Equal(t, len(set.elements), 1, "Set should contain 1 element")
}

func TestAddRemove(t *testing.T) {
	set := NewSet[string]()

	set.Add("test")
	assert.True(t, set.Contains("test"), "Set should contain 'test'")
	assert.Equal(t, len(set.elements), 1, "Set should contain 1 element")

	set.Remove("test")
	assert.False(t, set.Contains("test"), "Set should not contain 'test'")
	assert.Equal(t, len(set.elements), 0, "Set should be empty")
}

func TestInitiallyEmpty(t *testing.T) {
	set := NewSet[string]()
	assert.False(t, set.Contains("test"), "Set should not contain 'test'")
	assert.Equal(t, len(set.elements), 0, "Set should be empty")
}

func TestIterator(t *testing.T) {
	set := NewSet[string]()
	set.Add("test1")
	set.Add("test2")
	set.Add("test3")

	// Collect elements from the iterator
	var elements []string
	for elem := range set.Iterator() {
		elements = append(elements, elem)
	}

	assert.ElementsMatch(t, elements, []string{"test1", "test2", "test3"}, "Iterator should return all elements")
}

func TestSize(t *testing.T) {
	set := NewSet[string]()
	assert.Equal(t, 0, set.Size(), "Set should be empty")

	set.Add("test1")
	assert.Equal(t, 1, set.Size(), "Set should contain 1 element")

	set.Add("test2")
	assert.Equal(t, 2, set.Size(), "Set should contain 2 elements")

	set.Remove("test1")
	assert.Equal(t, 1, set.Size(), "Set should contain 1 element after removal")

	set.Remove("test2")
	assert.Equal(t, 0, set.Size(), "Set should be empty after removing all elements")
}
