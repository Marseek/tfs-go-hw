package repository

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func (r *repo) GetUsersMap() map[string]string {
	var users map[string]string
	f, err := os.Open("users.json")
	if err != nil {
		r.logger.Errorln("Can't open file with user's data.")
		users["jlexie"] = "passwd"
		return users
	}
	data, _ := ioutil.ReadAll(f)
	_ = f.Close()

	err = json.Unmarshal(data, &users)
	if err != nil {
		r.logger.Errorln("Can't unmarshall file with user data.")
		users["jlexie"] = "passwd"
		return users
	}
	return users
}
