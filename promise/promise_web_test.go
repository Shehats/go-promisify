package promise

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type user struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type post struct {
	UserID  string `json:"user_id"`
	PostID  string `json:"post_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var users []user = []user{
	{
		UserID: "1990998jfdhsdjhds",
		Name:   "John Doe",
	},
	{
		UserID: "457847845jdkgfjfg",
		Name:   "Jane Doe",
	},
	{
		UserID: "34834573487354dfgfdg",
		Name:   "Johnathon Bard",
	},
}

var posts []post = []post{
	{
		UserID:  "1990998jfdhsdjhds",
		PostID:  "58857686786dfjfdjfjfgj",
		Title:   "Ninja Turtles",
		Content: "Teenage Mutant Ninja Turtles is an American media franchise created by the comic book artists Kevin Eastman and Peter Laird.",
	},
	{
		UserID:  "457847845jdkgfjfg",
		PostID:  "877989ufjdhdfjhdfjfdhf",
		Title:   "The Walkind Dead",
		Content: "The Walking Dead is an American post-apocalyptic horror drama television series based on the comic book series of the same name by Robert Kirkman, Tony Moore, and Charlie Adlard—together forming the core of The Walking Dead franchise.",
	},
	{
		UserID:  "34834573487354dfgfdg",
		PostID:  "67988999fhdjdhsfdkhdsjj",
		Title:   "Hot Takes",
		Content: "In journalism, a hot take is a \"piece of deliberately provocative commentary that is based almost entirely on shallow moralizing\" in response to a news story,",
	},
	{
		UserID:  "1990998jfdhsdjhds",
		PostID:  "989894589458dfjkgfjkfgdgdkjgd",
		Title:   "Hot Dogs",
		Content: "A hot dog (uncommonly spelled hotdog) is a food consisting of a grilled or steamed sausage served in the slit of a partially sliced bun.",
	},
	{
		UserID:  "457847845jdkgfjfg",
		PostID:  "090998989893434hgfdjfsdgjfdsfh",
		Title:   "Rolling the Dice",
		Content: "It means to take a chance or a risk or take a gamble. To roll a set of dice is to gamble.",
	},
	{
		UserID:  "34834573487354dfgfdg",
		PostID:  "223433500dkfjfdkdfjkfjfd",
		Title:   "Songs",
		Content: "Music is generally defined as the art of arranging sound to create some combination of form, harmony, melody, rhythm or otherwise expressive content. Exact definitions of music vary considerably around the world, though it is an aspect of all human societies, a cultural universal.",
	},
	{
		UserID:  "1990998jfdhsdjhds",
		PostID:  "4578457458745djfshdfjkdhfjkdfhj",
		Title:   "Drive",
		Content: "Drive is a 2011 American action drama film directed by Nicolas Winding Refn. The screenplay, written by Hossein Amini, is based on James Sallis's 2005 novel of the same name. The film stars Ryan Gosling as an unnamed Hollywood stunt driver who moonlights as a getaway driver",
	},
	{
		UserID:  "457847845jdkgfjfg",
		PostID:  "023948375885djfhdjfdhjfd",
		Title:   "Fast Car",
		Content: "Song written and composed by Tracy Chapman, originally recorded by Tracy Chapman in 1987 and released in 1988",
	},
	{
		UserID:  "34834573487354dfgfdg",
		PostID:  "49534983459kfdjkjdfjfdghkjgk",
		Title:   "Food",
		Content: "One needs to eat.",
	},
	{
		UserID:  "1990998jfdhsdjhds",
		PostID:  "09098948874583487745dfjhfdjfhfdjhdfh",
		Title:   "Milkshake",
		Content: "A milkshake (sometimes simply called a shake) is a sweet beverage made by blending milk, ice cream, and flavorings or sweeteners such as butterscotch, caramel sauce, chocolate syrup, fruit syrup, or whole fruit into a thick, sweet, cold mixture. It may also be made using a base made from non-dairy",
	},
}

type postsResponse struct {
	Posts []post `json:"posts"`
}

type usersResponse struct {
	Users []user `json:"users"`
}

type errorResponse struct {
	Message string `json:"message"`
}

var userMappings map[string][]post = map[string][]post{
	"1990998jfdhsdjhds": {
		{
			UserID:  "1990998jfdhsdjhds",
			PostID:  "58857686786dfjfdjfjfgj",
			Title:   "Ninja Turtles",
			Content: "Teenage Mutant Ninja Turtles is an American media franchise created by the comic book artists Kevin Eastman and Peter Laird.",
		},
		{
			UserID:  "1990998jfdhsdjhds",
			PostID:  "989894589458dfjkgfjkfgdgdkjgd",
			Title:   "Hot Dogs",
			Content: "A hot dog (uncommonly spelled hotdog) is a food consisting of a grilled or steamed sausage served in the slit of a partially sliced bun.",
		},
		{
			UserID:  "1990998jfdhsdjhds",
			PostID:  "4578457458745djfshdfjkdhfjkdfhj",
			Title:   "Drive",
			Content: "Drive is a 2011 American action drama film directed by Nicolas Winding Refn. The screenplay, written by Hossein Amini, is based on James Sallis's 2005 novel of the same name. The film stars Ryan Gosling as an unnamed Hollywood stunt driver who moonlights as a getaway driver",
		},
		{
			UserID:  "1990998jfdhsdjhds",
			PostID:  "09098948874583487745dfjhfdjfhfdjhdfh",
			Title:   "Milkshake",
			Content: "A milkshake (sometimes simply called a shake) is a sweet beverage made by blending milk, ice cream, and flavorings or sweeteners such as butterscotch, caramel sauce, chocolate syrup, fruit syrup, or whole fruit into a thick, sweet, cold mixture. It may also be made using a base made from non-dairy",
		},
	},
	"457847845jdkgfjfg": {
		{
			UserID:  "457847845jdkgfjfg",
			PostID:  "877989ufjdhdfjhdfjfdhf",
			Title:   "The Walkind Dead",
			Content: "The Walking Dead is an American post-apocalyptic horror drama television series based on the comic book series of the same name by Robert Kirkman, Tony Moore, and Charlie Adlard—together forming the core of The Walking Dead franchise.",
		},
		{
			UserID:  "457847845jdkgfjfg",
			PostID:  "090998989893434hgfdjfsdgjfdsfh",
			Title:   "Rolling the Dice",
			Content: "It means to take a chance or a risk or take a gamble. To roll a set of dice is to gamble.",
		},
		{
			UserID:  "457847845jdkgfjfg",
			PostID:  "023948375885djfhdjfdhjfd",
			Title:   "Fast Car",
			Content: "Song written and composed by Tracy Chapman, originally recorded by Tracy Chapman in 1987 and released in 1988",
		},
	},
	"34834573487354dfgfdg": {
		{
			UserID:  "34834573487354dfgfdg",
			PostID:  "67988999fhdjdhsfdkhdsjj",
			Title:   "Hot Takes",
			Content: "In journalism, a hot take is a \"piece of deliberately provocative commentary that is based almost entirely on shallow moralizing\" in response to a news story,",
		},
		{
			UserID:  "34834573487354dfgfdg",
			PostID:  "223433500dkfjfdkdfjkfjfd",
			Title:   "Songs",
			Content: "Music is generally defined as the art of arranging sound to create some combination of form, harmony, melody, rhythm or otherwise expressive content. Exact definitions of music vary considerably around the world, though it is an aspect of all human societies, a cultural universal.",
		},
		{
			UserID:  "34834573487354dfgfdg",
			PostID:  "49534983459kfdjkjdfjfdghkjgk",
			Title:   "Food",
			Content: "One needs to eat.",
		},
	},
}

func createHttpHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			time.Sleep(time.Second)
			if r.URL.Path == "/users" || r.URL.Path == "users" || r.URL.Path == "users/" || r.URL.Path == "/users/" {
				resp := usersResponse{
					Users: users,
				}
				b, _ := json.Marshal(resp)
				w.Write(b)
			} else if strings.Contains(r.URL.Path, "users") && r.URL.Query().Has("userId") {
				userId := r.URL.Query().Get("userId")
				for _, u := range users {
					if u.UserID == userId {
						b, _ := json.Marshal(u)
						w.Write(b)
						return
					}
				}
			} else if strings.Contains(r.URL.Path, "posts") && r.URL.Query().Has("userId") {
				userId := r.URL.Query().Get("userId")
				var resp postsResponse
				p := make([]post, 0)
				for _, v := range posts {
					if v.UserID == userId {
						p = append(p, v)
					}
				}
				resp.Posts = p
				b, _ := json.Marshal(resp)
				w.Write(b)
			} else {
				err := errorResponse{
					Message: "invalid reques",
				}
				b, _ := json.Marshal(err)
				w.Write(b)
			}
		}
	})
}

func callApi(handler http.HandlerFunc, method, url string) (*bytes.Buffer, error) {
	req, err := http.NewRequest(method, url, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Body, err
}
func getUsers(t *testing.T, handler http.HandlerFunc) ([]user, error) {
	p := Promisify[*bytes.Buffer](callApi, handler, "GET", "/users")
	p1 := Then(p, func(resp *bytes.Buffer) ([]user, error) {
		var r usersResponse
		json.Unmarshal(resp.Bytes(), &r)
		return r.Users, nil
	})
	ret, err := p1.Await()
	assert.NoError(t, err)
	assert.Equal(t, ret, users)
	return ret, err
}

func getPosts(t *testing.T, handler http.HandlerFunc, userId string) ([]post, error) {
	url := "/posts?userId=" + userId
	p := Promisify[*bytes.Buffer](callApi, handler, "GET", url)
	p1 := Then(p, func(resp *bytes.Buffer) ([]post, error) {
		var r postsResponse
		json.Unmarshal(resp.Bytes(), &r)
		return r.Posts, nil
	})
	ret, err := p1.Await()
	assert.NoError(t, err)
	assert.Equal(t, ret, userMappings[userId])
	return ret, err
}

func TestGetDataFromServer(t *testing.T) {
	handler := createHttpHandler()
	p := Promisify[[]user](getUsers, t, handler)
	p.Then(func(usrs []user) {
		ret := map[string][]post{}
		for _, u := range usrs {
			usrsPosts, err := getPosts(t, handler, u.UserID)
			if err != nil {
				return
			}
			ret[u.UserID] = usrsPosts
		}
		assert.Equal(t, ret, userMappings)
	})
	p.Catch(func(err error) {
		assert.Fail(t, "This shouldn't get called")
	})
}
