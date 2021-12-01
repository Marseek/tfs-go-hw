package repository

import (
	"encoding/json"
	"io/ioutil"
)

func (r *Repo) GetUsersMap(file string) map[string]string {
	var users map[string]string

	data, err := ioutil.ReadFile(file)
	if err != nil {
		r.logger.Errorln("Can't open file with user's data: ", err)
	}

	err = json.Unmarshal(data, &users)
	if err != nil {
		r.logger.Errorln("Can't unmarshall file with user data.")
		users["jlexie"] = "passwd"
		return users
	}

	return users
}
