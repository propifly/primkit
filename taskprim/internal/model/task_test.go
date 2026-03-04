package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_Validate_Valid(t *testing.T) {
	task := &Task{
		List:   "andres",
		What:   "reply to accountant",
		Source: "johanna",
	}
	assert.NoError(t, task.Validate())
}

func TestTask_Validate_MissingList(t *testing.T) {
	task := &Task{What: "test", Source: "cli"}
	err := task.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list")
}

func TestTask_Validate_MissingWhat(t *testing.T) {
	task := &Task{List: "andres", Source: "cli"}
	err := task.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "what")
}

func TestTask_Validate_MissingSource(t *testing.T) {
	task := &Task{List: "andres", What: "test"}
	err := task.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source")
}
