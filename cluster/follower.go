package cluster

//this should contain the logic for followers

type Followers []Node

var followers Followers

func AddFollower(node Node) {
	followers = append(followers, node)
}

func GetFollowers() Followers {
	return followers
}
