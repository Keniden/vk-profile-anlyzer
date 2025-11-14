package domain

import "time"

type VKUser struct {
	ID         int64
	ScreenName string
	FirstName  string
	LastName   string
	Sex        int
	BirthDate  string
	City       string
	About      string
}

type WallPost struct {
	ID       int64
	Date     time.Time
	Text     string
	Likes    int
	Reposts  int
	Comments int
	Views    int
	PostType string
	IsPinned bool
}

type Gift struct {
	ID   int64
	Text string
}

type Friend struct {
	ID        int64
	FirstName string
	LastName  string
	Sex       int
}

type ActivityVector struct {
	PostsPerMonth       float64
	AveragePostLen      float64
	EngagementRate      float64
	GiftsCount          int
	FriendsCount        int
	ProfileCompleteness float64
}

type ProfileData struct {
	User    VKUser
	Wall    []WallPost
	Gifts   []Gift
	Friends []Friend
	Vector  ActivityVector
}

type Profile struct {
	ID         uint      `gorm:"primaryKey"`
	VKID       int64     `gorm:"uniqueIndex;not null"`
	ScreenName string    `gorm:"size:255"`
	FullName   string    `gorm:"size:255"`
	RawJSON    string    `gorm:"type:text"`
	Summary    string    `gorm:"type:text"`
	UpdatedAt  time.Time
	CreatedAt  time.Time
}

