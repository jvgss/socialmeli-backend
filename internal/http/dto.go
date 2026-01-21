package http

type FollowersCountResponse struct {
	UserID         int    `json:"userId"`
	UserName       string `json:"userName"`
	FollowersCount int    `json:"followersCount"`
}

type SimpleUser struct {
	UserID    int    `json:"userId,omitempty"`
	UserName  string `json:"userName,omitempty"`
	UserID2   int    `json:"user_id,omitempty"`
	UserName2 string `json:"user_name,omitempty"`
}

type FollowersListResponse struct {
	UserID    int          `json:"userId"`
	UserName  string       `json:"userName"`
	Followers []SimpleUser `json:"followers"`
}

type FollowedListResponse struct {
	UserID   int          `json:"user_id"`
	UserName string       `json:"user_name"`
	Followed []SimpleUser `json:"followed"`
}

type PublishResponse struct {
	PostID int `json:"post_id"`
}

type FollowedPostsResponse struct {
	UserID int `json:"user_id"`
	Posts  any `json:"posts"`
}

type PromoCountResponse struct {
	UserID             int    `json:"user_id"`
	UserName           string `json:"user_name"`
	PromoProductsCount int    `json:"promo_products_count"`
}

type PromoListResponse struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Posts    any    `json:"posts"`
}
