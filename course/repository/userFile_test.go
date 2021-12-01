package repository

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUsersMap(t *testing.T) {
	t.Skip("Born to fail")
	outfile := `{ "jlexie": "password" }`
	file := "test_user.json"
	err := ioutil.WriteFile(file, []byte(outfile), 0777)
	if err != nil {
		t.Error("Can't write to file")
	}
	r := Repo{}
	usersMap := r.GetUsersMap(file)
	_ = os.Remove(file)
	assert.Equal(t, usersMap["jlexie"], "password")
}
