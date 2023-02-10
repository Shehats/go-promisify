package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Shehats/go-promisify/promise"
)

type TranslatedName struct {
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
}

type Chapter struct {
	Id              int            `json:"id"`
	RevelationPlace string         `json:"revelation_place"`
	RevelationOrder int            `json:"revelation_order"`
	BismillahPre    bool           `json:"bismillah_pre"`
	NameComplex     string         `json:"name_complex"`
	NameArabic      string         `json:"name_arabic"`
	VersesCount     int            `json:"verses_count"`
	Pages           []int          `json:"pages"`
	TranslatedName  TranslatedName `json:"translated_name"`
}

type ChaptersResponse struct {
	Chapters []Chapter `json:"chapters"`
}

func getData(url string) (*http.Response, error) {
	fmt.Println("In get data")
	resp, err := http.Get(url)
	return resp, err
}

func runStuff() <-chan ChaptersResponse {
	c := make(chan ChaptersResponse, 1)
	chaptersUrl := "https://api.quran.com/api/v4/chapters?language=en"
	p1 := promise.Promisify[*http.Response](getData, chaptersUrl)
	p2 := promise.Then(p1, func(r *http.Response) (ChaptersResponse, error) {
		var chapters ChaptersResponse
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return ChaptersResponse{}, err
		}
		err = json.Unmarshal(b, &chapters)
		if err != nil {
			return ChaptersResponse{}, err
		}
		return chapters, nil
	})
	p2.Finally(func(cr ChaptersResponse, err error) {
		c <- cr
	})
	return c
}

func main() {
	c := <-runStuff()
	fmt.Println(c)
}
