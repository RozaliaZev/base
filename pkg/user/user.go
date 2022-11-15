package user

import(
	"fmt"
)

type User struct {
	Name string `json:"name"`
	Age int `json:"age"`
	Friends []string `json:"friend"`
}

type Ids struct {
	SourseId int `json:"source_id"` 
	TargetId int `json:"target_id"`
}

type NewUserAge struct {
	NewAge int `json:"new age"` 
}

func (u *User) ToString() string {
	return fmt.Sprintf("name is %s and age is %d\n", u.Name, u.Age)
}
