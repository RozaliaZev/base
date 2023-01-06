package user

import(
	"fmt"
)

type User struct {
	Id int `dbpool:"id"`
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
	return fmt.Sprintf("id: %d name is %s and age is %d\n", u.Id, u.Name, u.Age)
}

func (u *User) FindFriend(nameFriend User, friendList User) bool {
	for i:=0; i < len(friendList.Friends); i++ {
		if nameFriend.Name == friendList.Friends[i] {
		return true
		} 
	}
	return false
}
